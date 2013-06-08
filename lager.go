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

func init() {
	typeMap = make(map[string]reflect.Type)
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
