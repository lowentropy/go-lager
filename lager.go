package lager

import (
	"reflect"
)

// typeMap contains types by their full package name.
// It holds both struct and interface types.
var typeMap map[string]reflect.Type

// init builds the type map, which is the only package-wide data
func init() {
	typeMap = make(map[string]reflect.Type)
}

// Register allows you to specify a struct or interface value.
// The type of that value will be registered so that serialized objects
// correctly decode as the proper type.
func Register(value interface{}) {
	RegisterType(reflect.TypeOf(value))
}

// RegisterType allows you to specify a reflected struct or interface
// type. It will be registered so that values of this type are
// properly decoded.
func RegisterType(typ reflect.Type) {
	typeMap[typ.String()] = typ
}

// privateField checks whether the given struct field is exported
// (returns false) or not (returns true).
func privateField(f reflect.StructField) bool {
	return len(f.PkgPath) > 0
}

// numPublicFields returns the number of fields of the given struct
// type which are exported, i.e. those that start with a capital letter.
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

// isInterface returns whether the given arbitrary type is an interface
func isInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Interface
}

// isPtr returns whether the given arbitrrary type is a pointer
func isPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}
