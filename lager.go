package main

import (
	"reflect"
)

type err struct {
	message string
}

func (e err) Error() string {
	return e.message
}

var typeMap map[string]reflect.Type
var interfaceType reflect.Type

func init() {
	var slice []interface{}
	typeMap = make(map[string]reflect.Type)
	interfaceType = reflect.TypeOf(slice).Elem()
}

func registerType(t reflect.Type) {
	typeMap[t.String()] = t
}

func privateField(f reflect.StructField) bool {
	return len(f.PkgPath) > 0
}

func numPublicFields(t reflect.Type) int {
	n := t.NumField()
	public := 0
	for i := 0; i < n; i++ {
		field := t.Field(i)
		if !privateField(field) {
			public++
		}
	}
	return public
}

func isInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Interface
}

func isPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}
