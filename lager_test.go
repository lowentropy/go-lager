package main

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

func assertEncodes(t *testing.T, expected interface{}) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	enc.Write(expected)
	enc.Finish()
	dec, err := NewDecoder(buf)
	if err != nil {
		t.Fatal("Could not construct decoder")
	}
	actual, err := dec.Read()
	if err != nil {
		t.Fatal("Failed to read object")
	}
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
	assertEncodes(t, int(-11))
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
	assertEncodes(t, interface{}(216))
}

func TestEncodeMap(t *testing.T) {
	x := 216
	assertEncodes(t, map[int]string{3: "foo", 5: "bar"})
	assertEncodes(t, map[string]*int{"bluh": &x})
	assertEncodes(t, map[*int]complex128{&x: 5 - 2i})
	assertEncodes(t, map[int]map[string]float32{42: map[string]float32{"zugzug": 3.4}})
}

func TestEncodePtr(t *testing.T) {
	str := "aksdjfhasd"
	num := 1294
	pnum := &num
	ppnum := &pnum
	pppnum := &ppnum
	assertEncodes(t, &str)
	assertEncodes(t, &num)
	assertEncodes(t, &pppnum)
}

func TestEncodeSlice(t *testing.T) {
	s1 := "foo"
	s2 := "bar"
	assertEncodes(t, []int{1, 2, 3, 4, 5})
	assertEncodes(t, []*string{&s1, &s2})
}

func TestEncodeString(t *testing.T) {
	assertEncodes(t, "")
	assertEncodes(t, "askjdlhfakjdhflakjdshf")
}

func TestEncodeInterfaceSlice(t *testing.T) {
	assertEncodes(t, []interface{}{1, "foo", 2.5})
}

// TODO: the rest of them...
// TODO: common ptr test
// TODO: recursive ptr test
