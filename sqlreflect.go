package gosql

import (
	"database/sql"
	"errors"
	"reflect"
	"sync"
)

type StdScanner struct {
	Mapper *Mapper
}

type Mapper struct {
	TagName   string
	cacheMaps map[reflect.Type]*TypeMap
	mapFunc   func(t any, tagName string) (*TypeMap, error)
	cacheLock sync.RWMutex
}

type TypeMap struct {
	NonRefType reflect.Type
	Fields     []*FieldInfo
}

type FieldInfo struct {
	Name          string
	FtypeNonRef   reflect.Type
	FvalNonRef    reflect.Value
	FTag          string
	OriginalType  reflect.Type
	OriginalValue reflect.Value
}

// map the item, panics if type of item isn`t struct or pointer to the struct
// ===========================================================================================================
// сопоставляет элемент, впадает в панику, если тип элемента не является struct или указателем на структуру
func MapFunc(item any, tagName string) (*TypeMap, error) {

	ogItemType := reflect.TypeOf(item)

	myerr := errors.New("type need to be either struct or pointer to the struct")

	if kind := ogItemType.Kind(); kind != reflect.Pointer && kind != reflect.Struct {
		return nil, myerr
	}

	nonRefItemType := TransformToNonRefType(ogItemType)

	if kind := nonRefItemType.Kind(); kind != reflect.Struct {
		return nil, myerr
	}

	fields := []*FieldInfo{}

	for i := range nonRefItemType.NumField() {
		field := nonRefItemType.Field(i)
		columnName := field.Tag.Get(tagName)
		ogType := field.Type
		if len(columnName) > 0 && IsScannable(ogType) {
			fieldInfo := &FieldInfo{}

			ogValue := reflect.Zero(ogType)

			fieldInfo.OriginalType = ogType
			fieldInfo.OriginalValue = ogValue

			nonRefType := TransformToNonRefType(item)
			nonRefValue := reflect.Zero(nonRefItemType)

			fieldInfo.FtypeNonRef = nonRefType
			fieldInfo.FvalNonRef = nonRefValue

			fieldInfo.Name = field.Name

			fieldInfo.FTag = columnName

			fields = append(fields, fieldInfo)
		}

	}

	return &TypeMap{
		NonRefType: nonRefItemType,
		Fields:     fields,
	}, nil
}

func IsScannable(t reflect.Type) bool {
	scannable := reflect.TypeFor[sql.Scanner]()
	return reflect.PointerTo(scannable).Implements(scannable)
}

func (mapper *Mapper) Map(item any, tagName string) (*TypeMap, error) {
	nonRefType := TransformToNonRefType(item)

	var err error
	mapper.cacheLock.RLock()
	typeMap, ok := mapper.cacheMaps[nonRefType]
	mapper.cacheLock.RUnlock()

	if ok {
		return typeMap, nil
	}

	mapper.cacheLock.Lock()
	defer mapper.cacheLock.Unlock()

	typeMap, err = mapper.mapFunc(item, tagName)
	if typeMap != nil {
		mapper.cacheMaps[nonRefType] = typeMap
	}
	return typeMap, err
}

func (sc *StdScanner) Scan(item any, tagName string, rows *sql.Rows, excludedTags []string) error {
	typeMap, err := sc.Mapper.Map(item, tagName)
	_ = typeMap
	return err
}
