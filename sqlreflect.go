package gosql

import "reflect"

type Mapper struct {
	tagName   string
	cacheMaps map[reflect.Type]*TypeMap
	mapFunc   func(t reflect.Type) *TypeMap
}

type TypeMap struct {
	fields []FieldInfo
}

type FieldInfo struct {
	ftypeNonRef reflect.Type
	fvalNonRef  reflect.Value

	originalType  reflect.Type
	originalValue reflect.Value
}
