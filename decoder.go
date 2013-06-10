package main

import (
	"io"
	"math"
	"reflect"
	"unsafe"
)

type Decoder struct {
	reader     io.ByteReader
	objects    int
	typeMap    map[uint]reflect.Type
	ptrMap     map[uintptr]uintptr
	postHeader bool
}

func NewDecoder(r io.ByteReader) (*Decoder, error) {
	d := &Decoder{
		reader:     r,
		objects:    0,
		typeMap:    make(map[uint]reflect.Type),
		ptrMap:     make(map[uintptr]uintptr),
		postHeader: false,
	}
	if err := d.readHeader(); err != nil {
		return nil, err
	}
	d.postHeader = true
	return d, nil
}

func (d *Decoder) Read() (interface{}, error) {
	if d.objects == 0 {
		return nil, err{"Out of objects"}
	}
	d.objects--
	return d.read(d.readType()), nil
}

func (d *Decoder) readHeader() error {
	d.objects = d.readInt()
	n := d.readInt()
	for i := 0; i < n; i++ {
		name := d.readString()
		id := d.readUint()
		t, ok := typeMap[name]
		if !ok {
			return err{"Can't find type: " + name}
		}
		d.typeMap[id] = t
	}
	n = d.readInt()
	objs := make([]reflect.Value, n)
	for i := 0; i < n; i++ {
		ptr := d.readUintptr()
		t := d.readType()
		v := reflect.New(t)
		v.Elem().Set(reflect.ValueOf(d.read(t)))
		objs[i] = v.Elem()
		d.ptrMap[ptr] = v.Pointer()
	}
	for _, obj := range objs {
		d.patch(obj)
	}
	return nil
}

func (d *Decoder) patch(v reflect.Value) {
	switch v.Type().Kind() {
	case reflect.Slice:
		d.patchSlice(v)
	case reflect.Map:
		d.patchMap(v)
	case reflect.Struct:
		d.patchStruct(v)
	case reflect.Ptr:
		d.patchPtr(v)
	}
}

func (d *Decoder) patchPtr(v reflect.Value) {
	if isPtr(v.Type()) {
		ptr := unsafe.Pointer(d.ptrMap[v.Pointer()])
		newval := reflect.NewAt(v.Type().Elem(), ptr)
		v.Set(newval)
	}
}

func (d *Decoder) patchSlice(v reflect.Value) {
	n := v.Len()
	for i := 0; i < n; i++ {
		d.patch(v.Index(i))
	}
}

func (d *Decoder) patchMap(v reflect.Value) {
	for _, key := range v.MapKeys() {
		d.patch(key)
		d.patch(v.MapIndex(key))
	}
}

func (d *Decoder) patchStruct(v reflect.Value) {
	n := v.NumField()
	t := v.Type()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if !privateField(f) {
			d.patch(v.Field(i))
		}
	}
}

func (d *Decoder) readType() reflect.Type {
	kind := reflect.Kind(d.readUint8())
	switch kind {
	case reflect.Bool:
		return reflect.TypeOf(false)
	case reflect.Int:
		return reflect.TypeOf(int(0))
	case reflect.Int8:
		return reflect.TypeOf(int8(0))
	case reflect.Int16:
		return reflect.TypeOf(int16(0))
	case reflect.Int32:
		return reflect.TypeOf(int32(0))
	case reflect.Int64:
		return reflect.TypeOf(int64(0))
	case reflect.Uint:
		return reflect.TypeOf(uint(0))
	case reflect.Uint8:
		return reflect.TypeOf(uint8(0))
	case reflect.Uint16:
		return reflect.TypeOf(uint16(0))
	case reflect.Uint32:
		return reflect.TypeOf(uint32(0))
	case reflect.Uint64:
		return reflect.TypeOf(uint64(0))
	case reflect.Uintptr:
		return reflect.TypeOf(uintptr(0))
	case reflect.Float32:
		return reflect.TypeOf(float32(0))
	case reflect.Float64:
		return reflect.TypeOf(float64(0))
	case reflect.Complex64:
		return reflect.TypeOf(complex64(0))
	case reflect.Complex128:
		return reflect.TypeOf(complex128(0))
	case reflect.Array, reflect.Chan, reflect.Func:
		panic("Can't read " + kind.String() + " types")
	case reflect.Map:
		key := d.readType()
		elem := d.readType()
		return reflect.MapOf(key, elem)
	case reflect.Ptr:
		return reflect.PtrTo(d.readType())
	case reflect.Slice:
		return reflect.SliceOf(d.readType())
	case reflect.String:
		return reflect.TypeOf("")
	case reflect.Struct, reflect.Interface:
		id := d.readUint()
		t, ok := d.typeMap[id]
		if !ok {
			panic("Can't find type id!")
		}
		return t
	}
	panic("Unknown type kind: " + kind.String())
}

func (d *Decoder) readBool() bool {
	if d.readUint8() == 0 {
		return false
	} else {
		return true
	}
}

func (d *Decoder) readInt() int {
	return int(d.readInt64())
}

func (d *Decoder) readInt8() int8 {
	u := d.readUint8()
	if u&1 != 0 {
		return ^int8(u >> 1)
	} else {
		return int8(u >> 1)
	}
}

func (d *Decoder) readInt16() int16 {
	u := d.readUint16()
	if u&1 != 0 {
		return ^int16(u >> 1)
	} else {
		return int16(u >> 1)
	}
}

func (d *Decoder) readInt32() int32 {
	u := d.readUint32()
	if u&1 != 0 {
		return ^int32(u >> 1)
	} else {
		return int32(u >> 1)
	}
}

func (d *Decoder) readInt64() int64 {
	u := d.readUint64()
	if u&1 != 0 {
		return ^int64(u >> 1)
	} else {
		return int64(u >> 1)
	}
}

func (d *Decoder) readUint() uint {
	return uint(d.readUint64())
}

func (d *Decoder) readUint8() uint8 {
	u, _ := d.reader.ReadByte()
	return u
}

func (d *Decoder) readUint16() uint16 {
	u1, _ := d.reader.ReadByte()
	u2, _ := d.reader.ReadByte()
	return (uint16(u2) << 8) | uint16(u1)
}

func (d *Decoder) readUint32() uint32 {
	u1, _ := d.reader.ReadByte()
	u2, _ := d.reader.ReadByte()
	u3, _ := d.reader.ReadByte()
	u4, _ := d.reader.ReadByte()
	return (uint32(u4) << 24) |
		(uint32(u3) << 16) |
		(uint32(u2) << 8) |
		uint32(u1)
}

func (d *Decoder) readUint64() uint64 {
	u1, _ := d.reader.ReadByte()
	u2, _ := d.reader.ReadByte()
	u3, _ := d.reader.ReadByte()
	u4, _ := d.reader.ReadByte()
	u5, _ := d.reader.ReadByte()
	u6, _ := d.reader.ReadByte()
	u7, _ := d.reader.ReadByte()
	u8, _ := d.reader.ReadByte()
	return (uint64(u8) << 56) |
		(uint64(u7) << 48) |
		(uint64(u6) << 40) |
		(uint64(u5) << 32) |
		(uint64(u4) << 24) |
		(uint64(u3) << 16) |
		(uint64(u2) << 8) |
		uint64(u1)
}

func (d *Decoder) readUintptr() uintptr {
	return uintptr(d.readUint64())
}

func (d *Decoder) readFloat32() float32 {
	return math.Float32frombits(d.readUint32())
}

func (d *Decoder) readFloat64() float64 {
	return math.Float64frombits(d.readUint64())
}

func (d *Decoder) readComplex64() complex64 {
	r := math.Float32frombits(d.readUint32())
	i := math.Float32frombits(d.readUint32())
	return complex(r, i)
}

func (d *Decoder) readComplex128() complex128 {
	r := math.Float64frombits(d.readUint64())
	i := math.Float64frombits(d.readUint64())
	return complex(r, i)
}

func (d *Decoder) readMap(t reflect.Type) interface{} {
	n := d.readInt()
	v := reflect.MakeMap(t)
	keyType := t.Key()
	elemType := t.Elem()
	for i := 0; i < n; i++ {
		key := reflect.ValueOf(d.read(keyType))
		val := reflect.ValueOf(d.read(elemType))
		v.SetMapIndex(key, val)
	}
	return v.Interface()
}

func (d *Decoder) readPtr(t reflect.Type) interface{} {
	addr := d.readUintptr()
	if d.postHeader {
		patched, ok := d.ptrMap[addr]
		if !ok {
			panic("Missing pointer " + string(addr))
		}
		addr = patched
	}
	ptr := unsafe.Pointer(addr)
	return reflect.NewAt(t.Elem(), ptr).Interface()
}

func (d *Decoder) readSlice(t reflect.Type) interface{} {
	n := d.readInt()
	inner := t.Elem()
	v := reflect.MakeSlice(t, 0, n)
	for i := 0; i < n; i++ {
		elem := reflect.ValueOf(d.read(inner))
		v = reflect.Append(v, elem)
	}
	return v.Interface()
}

func (d *Decoder) readString() string {
	n := d.readInt()
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i], _ = d.reader.ReadByte()
	}
	return string(buf)
}

func (d *Decoder) readStruct(t reflect.Type) interface{} {
	n := d.readInt()
	v := reflect.New(t).Elem()
	for i := 0; i < n; i++ {
		name := d.readString()
		field, ok := t.FieldByName(name)
		if !ok {
			panic("Can't find field " + name + " in " + t.String())
		}
		value := reflect.ValueOf(d.read(field.Type))
		v.FieldByName(name).Set(value)
	}
	return v.Interface()
}

func (d *Decoder) read(t reflect.Type) interface{} {
	if isInterface(t) {
		t = d.readType()
	}

	var value interface{}

	switch t.Kind() {
	case reflect.Bool:
		value = d.readBool()
	case reflect.Int:
		value = d.readInt()
	case reflect.Int8:
		value = d.readInt8()
	case reflect.Int16:
		value = d.readInt16()
	case reflect.Int32:
		value = d.readInt32()
	case reflect.Int64:
		value = d.readInt64()
	case reflect.Uint:
		value = d.readUint()
	case reflect.Uint8:
		value = d.readUint8()
	case reflect.Uint16:
		value = d.readUint16()
	case reflect.Uint32:
		value = d.readUint32()
	case reflect.Uint64:
		value = d.readUint64()
	case reflect.Uintptr:
		value = d.readUintptr()
	case reflect.Float32:
		value = d.readFloat32()
	case reflect.Float64:
		value = d.readFloat64()
	case reflect.Complex64:
		value = d.readComplex64()
	case reflect.Complex128:
		value = d.readComplex128()
	case reflect.Array:
		panic("Can't read arrays")
	case reflect.Chan, reflect.Func:
		panic("Can't read " + t.Kind().String() + " types")
	case reflect.Interface:
		value = d.read(d.readType())
	case reflect.Map:
		value = d.readMap(t)
	case reflect.Ptr:
		value = d.readPtr(t)
	case reflect.Slice:
		value = d.readSlice(t)
	case reflect.String:
		value = d.readString()
	case reflect.Struct:
		value = d.readStruct(t)
	default:
		panic("Unknown type kind: " + t.Kind().String())
	}
	return value
}
