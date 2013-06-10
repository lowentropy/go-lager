package main

import (
	"reflect"
)

type UnsupportedRead struct {
	kind reflect.Kind
}

func (err UnsupportedRead) Error() string {
	return "Can't read " + err.kind.String() + " types"
}

type MissingTypeId struct {
	id uint
}

func (err MissingTypeId) Error() string {
	return "Encountered unknown type id " + string(err.id)
}

type MissingTypeName struct {
	name string
}

func (err MissingTypeName) Error() string {
	return "Encountered unknown type name " + err.name + "; you should register this type!"
}

type MissingPointer struct {
	ptr uintptr
}

func (err MissingPointer) Error() string {
	return "Missing pointer in map: " + string(err.ptr)
}

type MissingField struct {
	t    reflect.Type
	name string
}

func (err MissingField) Error() string {
	return "Missing field " + err.name + " in struct " + err.t.String()
}
