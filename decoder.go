package lager

import (
	"bufio"
	"io"
	"math"
	"reflect"
	"unsafe"
)

// Decoder is used to read Go objects from a stream of encoded bytes.
// Please note that the decoder is not thread-safe, and should only be
// used by a single goroutine.
type Decoder struct {
	reader     io.ByteReader
	objects    int
	typeMap    map[uint]reflect.Type
	ptrMap     map[uintptr]uintptr
	postHeader bool
}

// NewDecoder creates a new Decoder whose input source is the given
// io.Reader. On creation, the decoder reads the header section
// from the stream. Errors can occur during this phase.
func NewDecoder(r io.Reader) (*Decoder, error) {
	d := &Decoder{
		reader:     bufio.NewReader(r),
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

// Read returns the next object from the stream. If the end of stream
// has been reached, it returns an error.
func (d *Decoder) Read() (interface{}, error) {
	if d.objects == 0 {
		return nil, EndOfStream{}
	}
	d.objects--
	t, err := d.readType()
	if err != nil {
		return nil, err
	}
	return d.read(t)
}

func (d *Decoder) readHeader() error {
	var err error
	if d.objects, err = d.readInt(); err != nil {
		return err
	}
	if err = d.readTypeMap(); err != nil {
		return err
	}
	if err = d.readPtrMap(); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) readTypeMap() error {
	n, err := d.readInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		name, err := d.readString()
		if err != nil {
			return err
		}
		id, err := d.readUint()
		if err != nil {
			return err
		}
		t, ok := typeMap[name]
		if !ok {
			return MissingTypeName{name}
		}
		d.typeMap[id] = t
	}
	return nil
}

func (d *Decoder) readPtrMap() error {
	n, err := d.readInt()
	if err != nil {
		return err
	}
	objs := make([]reflect.Value, n)
	for i := 0; i < n; i++ {
		ptr, err := d.readUintptr()
		if err != nil {
			return err
		}
		t, err := d.readType()
		if err != nil {
			return err
		}
		v := reflect.New(t)
		value, err := d.read(t)
		if err != nil {
			return err
		}
		v.Elem().Set(reflect.ValueOf(value))
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

func (d *Decoder) readType() (reflect.Type, error) {
	id, err := d.readUint8()
	if err != nil {
		return nil, err
	}
	kind := reflect.Kind(id)

	switch kind {
	case reflect.Bool:
		return reflect.TypeOf(false), nil
	case reflect.Int:
		return reflect.TypeOf(int(0)), nil
	case reflect.Int8:
		return reflect.TypeOf(int8(0)), nil
	case reflect.Int16:
		return reflect.TypeOf(int16(0)), nil
	case reflect.Int32:
		return reflect.TypeOf(int32(0)), nil
	case reflect.Int64:
		return reflect.TypeOf(int64(0)), nil
	case reflect.Uint:
		return reflect.TypeOf(uint(0)), nil
	case reflect.Uint8:
		return reflect.TypeOf(uint8(0)), nil
	case reflect.Uint16:
		return reflect.TypeOf(uint16(0)), nil
	case reflect.Uint32:
		return reflect.TypeOf(uint32(0)), nil
	case reflect.Uint64:
		return reflect.TypeOf(uint64(0)), nil
	case reflect.Uintptr:
		return reflect.TypeOf(uintptr(0)), nil
	case reflect.Float32:
		return reflect.TypeOf(float32(0)), nil
	case reflect.Float64:
		return reflect.TypeOf(float64(0)), nil
	case reflect.Complex64:
		return reflect.TypeOf(complex64(0)), nil
	case reflect.Complex128:
		return reflect.TypeOf(complex128(0)), nil
	case reflect.Map:
		key, err := d.readType()
		if err != nil {
			return nil, err
		}
		elem, err := d.readType()
		if err != nil {
			return nil, err
		}
		return reflect.MapOf(key, elem), nil
	case reflect.Ptr:
		t, err := d.readType()
		if err != nil {
			return nil, err
		}
		return reflect.PtrTo(t), nil
	case reflect.Slice:
		t, err := d.readType()
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(t), nil
	case reflect.String:
		return reflect.TypeOf(""), nil
	case reflect.Struct, reflect.Interface:
		id, err := d.readUint()
		if err != nil {
			return nil, err
		}
		t, ok := d.typeMap[id]
		if !ok {
			return nil, MissingTypeId{id}
		}
		return t, nil
	}
	return nil, UnsupportedRead{kind}
}

func (d *Decoder) readBool() (bool, error) {
	u, err := d.readUint8()
	return (u != 0), err
}

func (d *Decoder) readInt() (int, error) {
	i, err := d.readInt64()
	return int(i), err
}

func (d *Decoder) readInt8() (int8, error) {
	u, err := d.readUint8()
	if u&1 != 0 {
		return ^int8(u >> 1), err
	} else {
		return int8(u >> 1), err
	}
}

func (d *Decoder) readInt16() (int16, error) {
	u, err := d.readUint16()
	if u&1 != 0 {
		return ^int16(u >> 1), err
	} else {
		return int16(u >> 1), err
	}
}

func (d *Decoder) readInt32() (int32, error) {
	u, err := d.readUint32()
	if u&1 != 0 {
		return ^int32(u >> 1), err
	} else {
		return int32(u >> 1), err
	}
}

func (d *Decoder) readInt64() (int64, error) {
	u, err := d.readUint64()
	if u&1 != 0 {
		return ^int64(u >> 1), err
	} else {
		return int64(u >> 1), err
	}
}

func (d *Decoder) readUint() (uint, error) {
	u, err := d.readUint64()
	return uint(u), err
}

func (d *Decoder) readUint8() (uint8, error) {
	return d.reader.ReadByte()
}

func (d *Decoder) readUint16() (uint16, error) {
	u1, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u2, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return (uint16(u2) << 8) | uint16(u1), nil
}

func (d *Decoder) readUint32() (uint32, error) {
	u1, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u2, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u3, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u4, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return (uint32(u4) << 24) |
		(uint32(u3) << 16) |
		(uint32(u2) << 8) |
		uint32(u1), nil
}

func (d *Decoder) readUint64() (uint64, error) {
	u1, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u2, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u3, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u4, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u5, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u6, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u7, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	u8, err := d.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return (uint64(u8) << 56) |
		(uint64(u7) << 48) |
		(uint64(u6) << 40) |
		(uint64(u5) << 32) |
		(uint64(u4) << 24) |
		(uint64(u3) << 16) |
		(uint64(u2) << 8) |
		uint64(u1), nil
}

func (d *Decoder) readUintptr() (uintptr, error) {
	u, err := d.readUint64()
	return uintptr(u), err
}

func (d *Decoder) readFloat32() (float32, error) {
	u, err := d.readUint32()
	return math.Float32frombits(u), err
}

func (d *Decoder) readFloat64() (float64, error) {
	u, err := d.readUint64()
	return math.Float64frombits(u), err
}

func (d *Decoder) readComplex64() (complex64, error) {
	r, err := d.readUint32()
	if err != nil {
		return 0, err
	}
	i, err := d.readUint32()
	if err != nil {
		return 0, err
	}
	return complex(math.Float32frombits(r), math.Float32frombits(i)), nil
}

func (d *Decoder) readComplex128() (complex128, error) {
	r, err := d.readUint64()
	if err != nil {
		return 0, err
	}
	i, err := d.readUint64()
	if err != nil {
		return 0, err
	}
	return complex(math.Float64frombits(r), math.Float64frombits(i)), nil
}

func (d *Decoder) readMap(t reflect.Type) (interface{}, error) {
	n, err := d.readInt()
	if err != nil {
		return nil, err
	}
	m := reflect.MakeMap(t)
	keyType := t.Key()
	elemType := t.Elem()
	for i := 0; i < n; i++ {
		k, err := d.read(keyType)
		if err != nil {
			return nil, err
		}
		v, err := d.read(elemType)
		if err != nil {
			return nil, err
		}
		m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return m.Interface(), nil
}

func (d *Decoder) readPtr(t reflect.Type) (interface{}, error) {
	addr, err := d.readUintptr()
	if err != nil {
		return nil, err
	}
	if d.postHeader {
		patched, ok := d.ptrMap[addr]
		if !ok {
			return nil, MissingPointer{addr}
		}
		addr = patched
	}
	ptr := unsafe.Pointer(addr)
	return reflect.NewAt(t.Elem(), ptr).Interface(), nil
}

func (d *Decoder) readSlice(t reflect.Type) (interface{}, error) {
	n, err := d.readInt()
	if err != nil {
		return nil, err
	}
	inner := t.Elem()
	v := reflect.MakeSlice(t, 0, n)
	for i := 0; i < n; i++ {
		elem, err := d.read(inner)
		if err != nil {
			return nil, err
		}
		v = reflect.Append(v, reflect.ValueOf(elem))
	}
	return v.Interface(), nil
}

func (d *Decoder) readString() (string, error) {
	n, err := d.readInt()
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		if buf[i], err = d.reader.ReadByte(); err != nil {
			return "", err
		}
	}
	return string(buf), nil
}

func (d *Decoder) readStruct(t reflect.Type) (interface{}, error) {
	n, err := d.readInt()
	if err != nil {
		return nil, err
	}
	v := reflect.New(t).Elem()
	for i := 0; i < n; i++ {
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		field, ok := t.FieldByName(name)
		if !ok {
			return nil, MissingField{t, name}
		}
		value, err := d.read(field.Type)
		if err != nil {
			return nil, err
		}
		v.FieldByName(name).Set(reflect.ValueOf(value))
	}
	return v.Interface(), nil
}

func (d *Decoder) read(t reflect.Type) (interface{}, error) {
	var err error
	if isInterface(t) {
		if t, err = d.readType(); err != nil {
			return nil, err
		}
	}

	var value interface{}

	switch t.Kind() {
	case reflect.Bool:
		value, err = d.readBool()
	case reflect.Int:
		value, err = d.readInt()
	case reflect.Int8:
		value, err = d.readInt8()
	case reflect.Int16:
		value, err = d.readInt16()
	case reflect.Int32:
		value, err = d.readInt32()
	case reflect.Int64:
		value, err = d.readInt64()
	case reflect.Uint:
		value, err = d.readUint()
	case reflect.Uint8:
		value, err = d.readUint8()
	case reflect.Uint16:
		value, err = d.readUint16()
	case reflect.Uint32:
		value, err = d.readUint32()
	case reflect.Uint64:
		value, err = d.readUint64()
	case reflect.Uintptr:
		value, err = d.readUintptr()
	case reflect.Float32:
		value, err = d.readFloat32()
	case reflect.Float64:
		value, err = d.readFloat64()
	case reflect.Complex64:
		value, err = d.readComplex64()
	case reflect.Complex128:
		value, err = d.readComplex128()
	case reflect.Interface:
		it, err := d.readType()
		if err == nil {
			value, err = d.read(it)
		}
	case reflect.Map:
		value, err = d.readMap(t)
	case reflect.Ptr:
		value, err = d.readPtr(t)
	case reflect.Slice:
		value, err = d.readSlice(t)
	case reflect.String:
		value, err = d.readString()
	case reflect.Struct:
		value, err = d.readStruct(t)
	default:
		err = UnsupportedRead{t.Kind()}
	}
	return value, err
}
