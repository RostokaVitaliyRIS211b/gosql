package sqlreflect

import (
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"sync"

	"github.com/RostokaVitaliyRIS211b/gosql/sqlstrings"
)

type Scanner interface {
	Scan(dest any, rows RowScanner, queryConfig sqlstrings.QueryConfig) error
}

type RowScanner interface {
	Scan(dest ...any) error
	Next() bool
}

type StdScanner struct {
	Mapper *Mapper
}

type Mapper struct {
	cacheMaps map[reflect.Type]*TypeMap
	MapFunc   func(t reflect.Type, tagName string) (*TypeMap, error)
	cacheLock sync.RWMutex
}

type TypeMap struct {
	NonRefType reflect.Type
	TagName    string
	Fields     []*FieldInfo
}

type FieldInfo struct {
	Name  string
	Ftype reflect.Type
	FTag  string
}

// map the item, panics if type of item isn`t struct or pointer to the struct
// ===========================================================================================================
// сопоставляет элемент, впадает в панику, если тип элемента не является struct или указателем на структуру
func MapFunc(item reflect.Type, tagName string) (*TypeMap, error) {

	ogItemType := item

	myerr := errors.New("type need to be either struct or pointer to the struct")

	if kind := ogItemType.Kind(); kind != reflect.Pointer && kind != reflect.Struct {
		return nil, myerr
	}

	nonRefItemType := ConversionTypeToNonRefType(item)

	if kind := nonRefItemType.Kind(); kind != reflect.Struct {
		return nil, myerr
	}

	fields := []*FieldInfo{}

	for i := range nonRefItemType.NumField() {
		field := nonRefItemType.Field(i)
		columnName := field.Tag.Get(tagName)
		ogType := field.Type
		nonRefType := ConversionTypeToNonRefType(ogType)
		if len(columnName) > 0 && field.IsExported() && IsScannable(nonRefType) && (ogType.Kind() != reflect.Pointer || (ogType.Kind() == reflect.Pointer && ogType.Elem() == nonRefType)) {
			fieldInfo := &FieldInfo{}

			fieldInfo.Ftype = ogType

			fieldInfo.Name = field.Name

			fieldInfo.FTag = columnName
			fields = append(fields, fieldInfo)
		}

	}

	return &TypeMap{
		NonRefType: nonRefItemType,
		Fields:     fields,
		TagName:    tagName,
	}, nil
}

var scannable = reflect.TypeFor[sql.Scanner]()

func IsScannable(t reflect.Type) bool {
	switch reflect.New(t).Interface().(type) {
	case *int:
		return true
	case *[]byte:
		return true
	case *int8:
		return true
	case *int16:
		return true
	case *int32:
		return true
	case *int64:
		return true
	case *uint:
		return true
	case *uint8:
		return true
	case *uint16:
		return true
	case *uint32:
		return true
	case *uint64:
		return true
	case *bool:
		return true
	case *float32:
		return true
	case *float64:
		return true
	case *string:
		return true
	default:
		return reflect.PointerTo(t).Implements(scannable)
	}
}

func (mapper *Mapper) Map(item reflect.Type, tagName string) (*TypeMap, error) {
	nonRefType := ConversionValToNonRefType(item)

	var err error
	mapper.cacheLock.RLock()
	typeMap, ok := mapper.cacheMaps[nonRefType]
	mapper.cacheLock.RUnlock()
	ok = ok && (tagName == typeMap.TagName)

	if ok {
		return typeMap, nil
	}

	mapper.cacheLock.Lock()
	defer mapper.cacheLock.Unlock()

	typeMap, err = mapper.MapFunc(item, tagName)
	if typeMap != nil {
		mapper.cacheMaps[nonRefType] = typeMap
	}
	return typeMap, err
}

func (sc *StdScanner) Scan(dest any, rows RowScanner, queryConfig sqlstrings.QueryConfig) error {

	ogSliceType := reflect.TypeOf(dest)
	if ogSliceType.Kind() != reflect.Pointer {
		return errors.New("dest must be a pointer to slice")
	}

	nonRefSliceType := ogSliceType.Elem()

	if nonRefSliceType.Kind() != reflect.Slice {
		return errors.New("dest must be a slice")
	}

	sliceVal := reflect.ValueOf(dest).Elem()

	sliceVal.SetLen(0)

	ogType := nonRefSliceType.Elem()
	nonRefType := ConversionTypeToNonRefType(ogType)

	typeMap, err := sc.Mapper.Map(ogType, queryConfig.TagName)

	if err != nil {
		return err
	}

	for rows.Next() {
		itemZero := reflect.New(nonRefType)
		err := rows.Scan(GetFieldsPointersOfItem(itemZero, typeMap, queryConfig.ExcludedTags)...)
		if err == nil {
			var ogVal any
			ogVal, err = ConversionToOgType(itemZero.Elem().Interface(), ogType)
			if err == nil {
				sliceVal.Set(reflect.Append(sliceVal, reflect.ValueOf(ogVal)))
			}
		}
	}

	return err
}

// Get pointers to fields of item, then give it in rows.Scan(), here you need to pass a pointer to the structure
// ===================================================================================================
// Получаем указатели на поля элемента, затем передаем их в rows.Scan(), сюда нужно передавать указатель на структуру
func GetFieldsPointersOfItem(item reflect.Value, tmap *TypeMap, excludedTags []string) []any {
	var pointers []any
	t := item.Type()
	if t.Kind() != reflect.Pointer && t.Elem().Kind() != reflect.Struct {
		return pointers
	}

	v := item.Elem()

	for _, fieldInfo := range tmap.Fields {
		if !slices.Contains(excludedTags, fieldInfo.FTag) {
			field := v.FieldByName(fieldInfo.Name)
			if fieldInfo.Ftype.Kind() == reflect.Pointer {
				if field.CanSet() {
					field.Set(reflect.New(fieldInfo.Ftype.Elem()))
				}
				pointers = append(pointers, field.Interface())
			} else if field.CanAddr() {
				pointers = append(pointers, field.Addr().Interface())
			}
		}
	}

	return pointers
}

// Conversion to the original type, it can be *User, but I can only get fields from the type from User, and I need to return *User back
// ============================================================================================================
// Приведение в исходный тип, он может быть *User, но я могу получить поля только от типа от User, и мне нужно вернуть обратно *User
func ConversionToOgType(item any, og reflect.Type) (any, error) {
	var res any
	t := reflect.TypeOf(item)

	if t == og {
		return item, nil
	}

	if og.Kind() != reflect.Pointer {
		return nil, errors.New("type need to be either struct or pointer to the struct...")
	}

	v := reflect.New(t)
	elem := v.Elem()

	if elem.CanSet() {
		elem.Set(reflect.ValueOf(item))
		t = v.Type()
	} else {
		return nil, errors.New("somehow value cannot be set")
	}

	ogV := reflect.New(og).Elem()

	for t != og {

		nv := reflect.New(t)
		elem = nv.Elem()

		if elem.CanSet() {
			elem.Set(v)
			v = nv
			t = v.Type()
		} else {
			return nil, errors.New("somehow value cannot be set")
		}
	}

	if ogV.CanSet() {
		ogV.Set(v)
		res = ogV.Interface()
	} else {
		return nil, errors.New("og type cannot be set")
	}

	return res, nil
}

func ConversionValToNonRefType(value any) reflect.Type {
	typeOfVal := reflect.TypeOf(value)
	kind := typeOfVal.Kind()

	for kind == reflect.Pointer {
		typeOfVal = typeOfVal.Elem()
		kind = typeOfVal.Kind()
	}
	return typeOfVal
}

func GetMapper() *Mapper {
	return &Mapper{
		MapFunc:   MapFunc,
		cacheMaps: map[reflect.Type]*TypeMap{},
	}
}

func GetScanner() Scanner {
	return &StdScanner{
		Mapper: GetMapper(),
	}
}

func ConversionTypeToNonRefType(t reflect.Type) reflect.Type {
	kind := t.Kind()

	for kind == reflect.Pointer {
		t = t.Elem()
		kind = t.Kind()
	}
	return t
}
