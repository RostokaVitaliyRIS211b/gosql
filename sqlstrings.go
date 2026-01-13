package gosql

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const stdTagName = "dbcn"

//region InsertQuery

// TableName - Имя таблицы
// NameWrapper - нужен для оборачивания имен столбцов и таблиц если это нужно, если указать например " то имя будет "SomeName"
// IdColumnName - нужен для того чтобы можно было вернуть Id добавленной записи, если указан то в конец строки добавится:  RETURNING IdColumnName
// TagName - нужен для того если вы хотите использовать нестандартный тег для полей структуры, тогда вместо стандартного dbcn будет использоваться указанный тег
// ItemToAdd - структура содержащая поля с тегами, значение которых соответсвует названиям столбцов таблицы
// ExcludedTags - список тегов которые вы хотите исключить при созданнии строки, например что типа Id, тогда конечная строка не будет содержать данного столбца
type QueryConfig struct {
	TableName    string
	NameWrapper  string
	IdColumnName string
	TagName      string
	ItemToAdd    any
	ExcludedTags []string
}

//region FuncWrappers

// func GetInsertQuery(tableName string, item any, excludedTags ...string) string {
// 	return GetInsertQueryReflect(QueryConfig{TableName: tableName,
// 		ItemToAdd:    item,
// 		NameWrapper:  "",
// 		IdColumnName: "",
// 		TagName:      "",
// 		ExcludedTags: excludedTags,
// 	})
// }

// func GetInsertQueryWrapNames(tableName string, item any, wrapStr string, excludedTags ...string) string {
// 	return GetInsertQueryReflect(QueryConfig{
// 		ItemToAdd:    item,
// 		NameWrapper:  wrapStr,
// 		IdColumnName: "",
// 		TagName:      "",
// 		ExcludedTags: excludedTags,
// 	})
// }

// func GetInsertQueryReturnId(tableName string, item any, idColumnName string, excludedTags ...string) string {
// 	return GetInsertQueryReflect(QueryConfig{
// 		ItemToAdd:    item,
// 		NameWrapper:  "",
// 		IdColumnName: idColumnName,
// 		TagName:      "",
// 		ExcludedTags: excludedTags,
// 	})
// }

// func GetInsertQueryCustomTag(tableName string, item any, tag string, excludedTags ...string) string {
// 	return GetInsertQueryReflect(QueryConfig{TableName: tableName,
// 		ItemToAdd:    item,
// 		NameWrapper:  "",
// 		IdColumnName: "",
// 		TagName:      tag,
// 		ExcludedTags: excludedTags,
// 	})
// }

//endregion

// Порядок аргументов должен соотвествовать порядку полей в передаваемой структуре
func GetInsertQueryReflect(params QueryConfig) string {
	var builder strings.Builder

	tagName := stdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	additionalSymbols := 37
	totalSymbols := len(params.TableName) + additionalSymbols

	builder.WriteString("INSERT")
	builder.WriteString(" INTO ")
	builder.WriteString(params.TableName)
	builder.WriteString(" (")
	builder.Grow(totalSymbols)

	counter := 0

	typeOfN := TransformToNonRefType(params.ItemToAdd)

	isFieldDb := false
	isPrevFieldDb := isFieldDb
	for i := 0; i < typeOfN.NumField(); i++ {
		tag := typeOfN.Field(i).Tag.Get(tagName)

		isPrevFieldDb = isFieldDb
		isFieldDb = len(tag) > 0 && !slices.Contains(params.ExcludedTags, tag)

		if isFieldDb && isPrevFieldDb {
			builder.WriteString(",")
		}

		if isFieldDb {
			name := tag
			if len(params.NameWrapper) > 0 {
				name = WrapNigger(name, params.NameWrapper)
			}
			builder.WriteString(name)
			counter++
		}

	}

	builder.WriteString(") VALUES (")
	for idx := range counter {
		if idx > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(idx + 1))
	}

	builder.WriteString(")")

	if len(params.IdColumnName) > 0 {
		builder.WriteString(" RETURNING " + params.IdColumnName)
	}
	return builder.String()
}

//endregion

func TransformToNonRefType(value any) reflect.Type {

	typeOfVal := reflect.TypeOf(value)
	val := reflect.ValueOf(value)

	for kind := typeOfVal.Kind(); kind == reflect.Pointer || kind == reflect.Interface; {
		val = val.Elem()
		typeOfVal = val.Type()
	}

	return typeOfVal
}

func WrapNigger(n string, wrapper string) string {
	return wrapper + n + wrapper
}
