package main

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

type anInterface interface {
	aMethod()
}

type aStruct struct {
	A int
	B string
	C float64
}

func (_ aStruct) aMethod() {}

func roundtrip(t *testing.T, in interface{}) interface{} {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	enc.Write(in)
	enc.Finish()
	dec, err := NewDecoder(buf)
	if err != nil {
		t.Fatalf("Could not construct decoder: %v", err)
	}
	out, err := dec.Read()
	if err != nil {
		t.Fatalf("Failed to read object: %v", err)
	}
	return out
}

func assertEncodes(t *testing.T, expected interface{}) {
	actual := roundtrip(t, expected)
	va := reflect.ValueOf(actual)
	ve := reflect.ValueOf(expected)
	if va.Type() != ve.Type() {
		t.Fatal("Types don't match")
	}
	if differs(va, ve) {
		t.Fatal("Expected", expected, "but got", actual)
	}
}

func differs(actual, expected reflect.Value) bool {
	switch expected.Type().Kind() {
	case reflect.Map:
		return mapsDiffer(actual, expected)
	case reflect.Slice:
		return slicesDiffer(actual, expected)
	}
	return actual.Interface() != expected.Interface()
}

func mapsDiffer(actual, expected reflect.Value) bool {
	if actual.Len() != expected.Len() {
		return true
	}
	for _, key := range actual.MapKeys() {
		if differs(actual.MapIndex(key), expected.MapIndex(key)) {
			return true
		}
	}
	return false
}

func slicesDiffer(actual, expected reflect.Value) bool {
	if actual.Len() != expected.Len() {
		return true
	}
	n := actual.Len()
	for i := 0; i < n; i++ {
		if differs(actual.Index(i), expected.Index(i)) {
			return true
		}
	}
	return false
}

func TestEncodeBool(t *testing.T) {
	assertEncodes(t, true)
	assertEncodes(t, false)
}

func TestEncodeInt(t *testing.T) {
	assertEncodes(t, int(3))
	assertEncodes(t, int(0))
	assertEncodes(t, int(-1))
	assertEncodes(t, int(math.MinInt64))
	assertEncodes(t, int(math.MaxInt64))
}

func TestEncodeInt8(t *testing.T) {
	assertEncodes(t, int8(3))
	assertEncodes(t, int8(0))
	assertEncodes(t, int8(-11))
	assertEncodes(t, int(math.MinInt8))
	assertEncodes(t, int(math.MaxInt8))
}

func TestEncodeInt16(t *testing.T) {
	assertEncodes(t, int16(3))
	assertEncodes(t, int16(0))
	assertEncodes(t, int16(-11))
	assertEncodes(t, int(math.MinInt16))
	assertEncodes(t, int(math.MaxInt16))
}

func TestEncodeInt32(t *testing.T) {
	assertEncodes(t, int32(3))
	assertEncodes(t, int32(0))
	assertEncodes(t, int32(-11))
	assertEncodes(t, int(math.MinInt32))
	assertEncodes(t, int(math.MaxInt32))
}

func TestEncodeInt64(t *testing.T) {
	assertEncodes(t, int64(3))
	assertEncodes(t, int64(0))
	assertEncodes(t, int64(-11))
	assertEncodes(t, int(math.MinInt64))
	assertEncodes(t, int(math.MaxInt64))
}

func TestEncodeUint(t *testing.T) {
	assertEncodes(t, uint(0))
	assertEncodes(t, uint(3))
	assertEncodes(t, uint(math.MaxUint64))
}

func TestEncodeUint8(t *testing.T) {
	assertEncodes(t, uint8(0))
	assertEncodes(t, uint8(3))
	assertEncodes(t, uint8(math.MaxUint8))
}

func TestEncodeUint16(t *testing.T) {
	assertEncodes(t, uint16(0))
	assertEncodes(t, uint16(3))
	assertEncodes(t, uint16(math.MaxUint16))
}

func TestEncodeUint32(t *testing.T) {
	assertEncodes(t, uint32(0))
	assertEncodes(t, uint32(3))
	assertEncodes(t, uint32(math.MaxUint32))
}

func TestEncodeUint64(t *testing.T) {
	assertEncodes(t, uint64(0))
	assertEncodes(t, uint64(3))
	assertEncodes(t, uint64(math.MaxUint64))
}

func TestEncodeUintptr(t *testing.T) {
	assertEncodes(t, uintptr(0))
	assertEncodes(t, uintptr(3))
	assertEncodes(t, uintptr(math.MaxUint64))
}

func TestEncodeFloat32(t *testing.T) {
	assertEncodes(t, float32(-1.5))
	assertEncodes(t, float32(0))
	assertEncodes(t, float32(-17))
	assertEncodes(t, float32(-math.MaxFloat32))
	assertEncodes(t, float32(math.MaxFloat32))
}

func TestEncodeFloat64(t *testing.T) {
	assertEncodes(t, float64(-1.5))
	assertEncodes(t, float64(0))
	assertEncodes(t, float64(-17))
	assertEncodes(t, float64(-math.MaxFloat64))
	assertEncodes(t, float64(math.MaxFloat64))
}

func TestEncodeComplex64(t *testing.T) {
	assertEncodes(t, complex64(-2+5i))
	assertEncodes(t, complex64(55-2.7i))
}

func TestEncodeComplex128(t *testing.T) {
	assertEncodes(t, complex128(-2+5i))
	assertEncodes(t, complex128(55-2.7i))
}

func TestEncodeInterface(t *testing.T) {
	slice := []anInterface{aStruct{}}
	assertEncodes(t, slice)
}

func TestEncodeMap(t *testing.T) {
	assertEncodes(t, map[int]string{3: "foo", 5: "bar"})
	assertEncodes(t, map[int]map[string]float32{42: map[string]float32{"zugzug": 3.4}})
}

func TestEncodeSlice(t *testing.T) {
	assertEncodes(t, []int{1, 2, 3, 4, 5})
}

func TestEncodeInterfaceSlice(t *testing.T) {
	assertEncodes(t, []interface{}{1, "foo", 2.5})
}

func TestEncodeString(t *testing.T) {
	assertEncodes(t, "")
	assertEncodes(t, "askjdlhfakjdhflakjdshf")
}

func TestEncodeStruct(t *testing.T) {
	assertEncodes(t, aStruct{216, "foo", 3.14})
}

func TestPointerMovementAsInterface(t *testing.T) {
	value := aStruct{216, "foo", 3.14}
	p := []interface{}{&value}
	p_ := roundtrip(t, p).([]interface{})
	if p[0] == p_[0] {
		t.Fatal("Pointer identical to original")
	}
}

func TestPointerMovementAsSlice(t *testing.T) {
	value := aStruct{216, "foo", 3.14}
	p := []*aStruct{&value}
	p_ := roundtrip(t, p).([]*aStruct)
	if p[0] == p_[0] {
		t.Fatal("Pointer identical to original")
	}
}

func TestPointerMovementAsStruct(t *testing.T) {
	value := aStruct{216, "foo", 3.14}

	type hasPtr struct {
		Ptr *aStruct
	}

	p := hasPtr{&value}
	p_ := roundtrip(t, p).(hasPtr)
	if p.Ptr == p_.Ptr {
		t.Fatal("Pointer identical to original")
	}
}

func TestSharedPointersAsStruct(t *testing.T) {
	value := aStruct{216, "foo", 3.14}

	type hasPtr struct {
		A, B *aStruct
	}

	p := hasPtr{&value, &value}
	p_ := roundtrip(t, p).(hasPtr)
	if p_.A != p_.B {
		t.Fatal("Shared pointers came back different")
	}
}

func TestSharedPointersAsSlice(t *testing.T) {
	value := aStruct{216, "foo", 3.14}
	ps := []*aStruct{&value, &value}
	ps_ := roundtrip(t, ps).([]*aStruct)
	if ps_[0] != ps_[1] {
		t.Fatal("Shared pointers came back different")
	}
}

func TestRecursivePointers(t *testing.T) {
	type hasPtr struct {
		Ptr *hasPtr
	}

	a := hasPtr{nil}
	b := hasPtr{&a}
	a.Ptr = &b

	s := []*hasPtr{&a, &b}
	s_ := roundtrip(t, s).([]*hasPtr)

	a_, b_ := s_[0], s_[1]
	if a_.Ptr != b_ || b_.Ptr != a_ {
		t.Fatal("Recursive pointers came back wrong")
	}
	if a_.Ptr == a.Ptr || b_.Ptr == b.Ptr {
		t.Fatal("Pointers were not moved")
	}
}

func TestEmbeddedPointerInPtrMap(t *testing.T) {
	s := aStruct{A: 3}
	m := [][]*aStruct{[]*aStruct{&s}}
	m_ := *roundtrip(t, &m).(*[][]*aStruct)
	m_[0][0].A = 5
	if s.A != 3 {
		t.Fatal("Embedded pointer from map was not patched")
	}
}
