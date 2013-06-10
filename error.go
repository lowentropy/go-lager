package lager

import (
	"reflect"
)

// UnsupportedRead is returned when the serialized data contains
// type information that the decoder doesn't know how to read.
type UnsupportedRead struct {
	kind reflect.Kind
}

func (err UnsupportedRead) Error() string {
	return "Can't read " + err.kind.String() + " types"
}

// MissingtypeId is returned when a unique type ID from the
// serialized data cannot be found by the decoder. This could happen
// if the data was invalid or corrupt.
type MissingTypeId struct {
	id uint
}

func (err MissingTypeId) Error() string {
	return "Encountered unknown type id " + string(err.id)
}

// MissingTypeName is returned when a named struct or interface type
// is present in the serialized data, but has not been registered. You
// can fix this by calling Register or RegisterType before reading from
// the decoder.
type MissingTypeName struct {
	name string
}

func (err MissingTypeName) Error() string {
	return "Encountered unknown type name " + err.name + "; you should register this type!"
}

// MissingPointer is returned when a pointer contained in a serialized
// object cannot be remapped by the deocder. This could happen if the data
// is invalid or corrupt.
type MissingPointer struct {
	ptr uintptr
}

func (err MissingPointer) Error() string {
	return "Missing pointer in map: " + string(err.ptr)
}

// MissingField is returned when a named field of a struct contained in the data
// cannot be found on the reflected type of that struct. This could happen if a
// field was renamed or removed from the struct between the time the data file was
// created and the time it was read again.
type MissingField struct {
	t    reflect.Type
	name string
}

func (err MissingField) Error() string {
	return "Missing field " + err.name + " in struct " + err.t.String()
}

// EndOfStream is returned when there are no more objects left in the encoded
// stream and a call to Read() is made.
type EndOfStream struct{}

func (_ EndOfStream) Error() string {
	return "End of stream reached, no more objects to return"
}
