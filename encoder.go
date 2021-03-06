package lager

import (
	"bytes"
	"io"
	"math"
	"reflect"
)

// Encoder is used to serialize objects to an encoded stream of bytes.
// Please note that the encoder is not thread-safe, and should only be
// used by a single goroutine.
type Encoder struct {
	buf     *bytes.Buffer
	writer  io.Writer
	nextId  uint
	objects int
	typeIds map[reflect.Type]uint
	ptrMap  map[uintptr]interface{}
}

// NewEncoder constructs a new encoder whose output stream is the
// given io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer:  w,
		nextId:  1,
		objects: 0,
		buf:     new(bytes.Buffer),
		typeIds: make(map[reflect.Type]uint),
		ptrMap:  make(map[uintptr]interface{}),
	}
}

// Write encodes the given object and places it into the stream.
// Objects are buffered until Finish() is called, because the header
// information must come first on the stream for decoding to work.
func (e *Encoder) Write(value interface{}) {
	e.write(value, true)
	e.objects++
}

// Finish should be called to terminate the stream. This collects
// type information and a map of pointers and pushes them to the
// output stream, followed by the buffered objects.
func (e *Encoder) Finish() {
	tmp := e.buf
	e.buf = new(bytes.Buffer)
	e.writeInt(e.objects)
	e.writeInt(len(e.typeIds))
	for t, id := range e.typeIds {
		e.writeString(t.String())
		e.writeUint(id)
	}
	e.writeInt(len(e.ptrMap))
	for ptr, v := range e.ptrMap {
		e.writeUintptr(ptr)
		e.write(v, true)
	}
	e.buf.WriteTo(e.writer)
	tmp.WriteTo(e.writer)
	e.buf = new(bytes.Buffer)
}

func (e *Encoder) registerType(t reflect.Type) uint {
	RegisterType(t)
	id, ok := e.typeIds[t]
	if !ok {
		id = e.nextId
		e.typeIds[t] = id
		e.nextId++
	}
	return id
}

func (e *Encoder) storePtr(w reflect.Value, ptr uintptr) {
	if _, ok := e.ptrMap[ptr]; !ok {
		e.ptrMap[ptr] = w.Elem().Interface()
		tmp := e.buf
		e.buf = new(bytes.Buffer)
		e.write(e.ptrMap[ptr], false)
		e.buf = tmp
	}
}

func (e *Encoder) writeType(t reflect.Type) {
	e.writeUint8(uint8(t.Kind()))
	switch t.Kind() {
	case reflect.Map:
		e.writeType(t.Key())
		e.writeType(t.Elem())
	case reflect.Ptr, reflect.Slice:
		e.writeType(t.Elem())
	case reflect.Struct, reflect.Interface:
		id := e.registerType(t)
		e.writeUint(id)
	}
}

func (e *Encoder) writeBool(v bool) {
	if v {
		e.writeUint8(1)
	} else {
		e.writeUint8(0)
	}
}

func (e *Encoder) writeInt(v int) {
	e.writeInt64(int64(v))
}

func (e *Encoder) writeInt8(v int8) {
	var u uint8
	if v < 0 {
		u = uint8(^v<<1) | 1
	} else {
		u = uint8(v << 1)
	}
	e.writeUint8(u)
}

func (e *Encoder) writeInt16(v int16) {
	var u uint16
	if v < 0 {
		u = uint16(^v<<1) | 1
	} else {
		u = uint16(v << 1)
	}
	e.writeUint16(u)
}

func (e *Encoder) writeInt32(v int32) {
	var u uint32
	if v < 0 {
		u = uint32(^v<<1) | 1
	} else {
		u = uint32(v << 1)
	}
	e.writeUint32(u)
}

func (e *Encoder) writeInt64(v int64) {
	var u uint64
	if v < 0 {
		u = uint64(^v<<1) | 1
	} else {
		u = uint64(v << 1)
	}
	e.writeUint64(u)
}

func (e *Encoder) writeUint(v uint) {
	e.writeUint64(uint64(v))
}

func (e *Encoder) writeUint8(v uint8) {
	e.buf.WriteByte(v)
}

func (e *Encoder) writeUint16(v uint16) {
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
}

func (e *Encoder) writeUint32(v uint32) {
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
}

func (e *Encoder) writeUint64(v uint64) {
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
	v >>= 8
	e.buf.WriteByte(byte(v))
}

func (e *Encoder) writeUintptr(v uintptr) {
	e.writeUint64(uint64(v))
}

func (e *Encoder) writeFloat32(v float32) {
	e.writeUint32(math.Float32bits(v))
}

func (e *Encoder) writeFloat64(v float64) {
	e.writeUint64(math.Float64bits(v))
}

func (e *Encoder) writeComplex64(v complex64) {
	e.writeUint32(math.Float32bits(real(v)))
	e.writeUint32(math.Float32bits(imag(v)))
}

func (e *Encoder) writeComplex128(v complex128) {
	e.writeUint64(math.Float64bits(real(v)))
	e.writeUint64(math.Float64bits(imag(v)))
}

func (e *Encoder) writeMap(v interface{}) {
	w := reflect.ValueOf(v)
	e.writeInt(w.Len())
	keyIsInterface := w.Type().Key().Kind() == reflect.Interface
	valIsInterface := w.Type().Elem().Kind() == reflect.Interface
	for _, key := range w.MapKeys() {
		e.write(key.Interface(), keyIsInterface)
		e.write(w.MapIndex(key).Interface(), valIsInterface)
	}
}

func (e *Encoder) writePtr(v interface{}) {
	w := reflect.ValueOf(v)
	ptr := w.Pointer()
	e.storePtr(w, ptr)
	e.writeUintptr(ptr)
}

func (e *Encoder) writeSlice(v interface{}) {
	w := reflect.ValueOf(v)
	e.writeInt(w.Len())
	isInterface := isInterface(w.Type().Elem())
	n := w.Len()
	for i := 0; i < n; i++ {
		e.write(w.Index(i).Interface(), isInterface)
	}
}

func (e *Encoder) writeString(v string) {
	e.writeInt(len(v))
	e.buf.WriteString(v)
}

func (e *Encoder) writeStruct(v interface{}) {
	w := reflect.ValueOf(v)
	t := w.Type()
	e.registerType(t)
	e.writeInt(numPublicFields(t))
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if privateField(f) {
			continue
		}
		e.writeString(f.Name)
		e.write(w.Field(i).Interface(), isInterface(f.Type))
	}
}

func (e *Encoder) write(v interface{}, sendType bool) {
	t := reflect.TypeOf(v)
	if sendType {
		e.writeType(t)
	}
	switch t.Kind() {
	case reflect.Bool:
		e.writeBool(v.(bool))
	case reflect.Int:
		e.writeInt(v.(int))
	case reflect.Int8:
		e.writeInt8(v.(int8))
	case reflect.Int16:
		e.writeInt16(v.(int16))
	case reflect.Int32:
		e.writeInt32(v.(int32))
	case reflect.Int64:
		e.writeInt64(v.(int64))
	case reflect.Uint:
		e.writeUint(v.(uint))
	case reflect.Uint8:
		e.writeUint8(v.(uint8))
	case reflect.Uint16:
		e.writeUint16(v.(uint16))
	case reflect.Uint32:
		e.writeUint32(v.(uint32))
	case reflect.Uint64:
		e.writeUint64(v.(uint64))
	case reflect.Uintptr:
		e.writeUintptr(v.(uintptr))
	case reflect.Float32:
		e.writeFloat32(v.(float32))
	case reflect.Float64:
		e.writeFloat64(v.(float64))
	case reflect.Complex64:
		e.writeComplex64(v.(complex64))
	case reflect.Complex128:
		e.writeComplex128(v.(complex128))
	case reflect.Array, reflect.Chan, reflect.Func, reflect.Interface:
		panic("Can't write " + t.Kind().String() + " types")
	case reflect.Map:
		e.writeMap(v)
	case reflect.Ptr:
		e.writePtr(v)
	case reflect.Slice:
		e.writeSlice(v)
	case reflect.String:
		e.writeString(v.(string))
	case reflect.Struct:
		e.writeStruct(v)
	default:
		panic("Unknown type kind: " + t.Kind().String())
	}
}
