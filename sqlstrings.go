package gosql

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const stdTagName = "dbcn"

//region Queries

//region InsertQuery

// TableName - Имя таблицы
// NameWrapper - нужен для оборачивания имен столбцов и таблиц если это нужно, если указать например " то имя будет "SomeName"
// ColumnName - (Insert) нужен для того чтобы можно было вернуть Id добавленной записи, если указан то в конец строки добавится:  RETURNING IdColumnName ;
// (Update) нужен для того чтобы обновить только определенные записи, в конец строки будет добавлено WHERE ColumnName = $1 ;
// (Select) нужен для того чтобы отфильтровать получаемые данные, в конец строки будет добавлено WHERE ColumnName = $1  ;
// (Delete) нужен для того чтобы удалить только определенные записи, в конец строки будет добавлено WHERE ColumnName = $1  ;
// TagName - нужен для того если вы используете нестандартный тег для полей структуры, тогда вместо стандартного dbcn будет использоваться указанный тег
// ItemToAdd - структура содержащая поля с тегами, значение которых соответсвует названиям столбцов таблицы
// ExcludedTags - список тегов которые вы хотите исключить при созданнии строки, например что типа Id, тогда конечная строка не будет содержать данного столбца
type QueryConfig struct {
	TableName    string
	NameWrapper  string
	ColumnName   string
	TagName      string
	ItemToAdd    any
	ExcludedTags []string
}

// Порядок аргументов должен соотвествовать порядку полей в передаваемой структуре
func GetInsertQuery(params QueryConfig) string {
	var builder strings.Builder

	tagName := stdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	//Выделяем память под символы сразу
	additionalSymbols := 37
	totalSymbols := len(params.TableName) + additionalSymbols + len(params.ColumnName)
	builder.Grow(totalSymbols)

	builder.WriteString("INSERT")
	builder.WriteString(" INTO ")

	tbname := params.TableName
	if len(params.NameWrapper) > 0 {
		tbname = WrapNigger(params.TableName, params.NameWrapper)
	}
	builder.WriteString(tbname)

	builder.WriteString(" (")

	counter := 0

	typeOfN := TransformToNonRefType(params.ItemToAdd)

	isFieldDb := false
	isPrevFieldDb := isFieldDb

	numFields := typeOfN.NumField()

	builder.Grow(numFields * (4 + 2*len(params.NameWrapper)))

	for i := 0; i < typeOfN.NumField(); i++ {
		tag := typeOfN.Field(i).Tag.Get(tagName)

		isPrevFieldDb = isFieldDb
		isFieldDb = len(tag) > 0 && !slices.Contains(params.ExcludedTags, tag)

		if isFieldDb && isPrevFieldDb {
			builder.WriteString(", ")
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

	if len(params.ColumnName) > 0 {
		idcolumn := params.ColumnName
		if len(params.NameWrapper) > 0 {
			idcolumn = WrapNigger(idcolumn, params.NameWrapper)
		}
		builder.WriteString(" RETURNING " + idcolumn)
	}
	return builder.String()
}

//endregion

//region UpdateQuery

// Если вы передаете ColumnName, тогда аргумент для него должен быть первым в списке аргументов, для остального порядок аргументов должен соотвествовать порядку полей в передаваемой структуре
func GetUpdateQuery(params QueryConfig) string {
	var builder strings.Builder
	additionalSymbols := 11
	totalSymbols := len(params.TableName) + len(params.ColumnName) + additionalSymbols
	builder.Grow(totalSymbols)

	tagName := stdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	builder.WriteString("UPDATE ")

	tbname := params.TableName
	if len(params.NameWrapper) > 0 {
		tbname = WrapNigger(params.TableName, params.NameWrapper)
	}
	builder.WriteString(tbname)

	builder.WriteString(" SET ")

	typeOfN := TransformToNonRefType(params.ItemToAdd)

	counter := 0

	numFields := typeOfN.NumField()

	builder.Grow(numFields * (4 + len(params.NameWrapper)))

	adder := 1

	if len(params.ColumnName) > 0 {
		adder = 2
	}

	isFieldDb := false
	isPrevFieldDb := isFieldDb

	for i := range numFields {
		tag := typeOfN.Field(i).Tag.Get(tagName)

		isPrevFieldDb = isFieldDb
		isFieldDb = len(tag) > 0 && !slices.Contains(params.ExcludedTags, tag)

		if isFieldDb && isPrevFieldDb {
			builder.WriteString(", ")
		}

		if isFieldDb {
			if len(params.NameWrapper) > 0 {
				tag = WrapNigger(tag, params.NameWrapper)
			}
			builder.WriteString(tag)
			builder.WriteString(" = $")
			builder.WriteString(strconv.Itoa(counter + adder))
			counter++
		}
	}

	if len(params.ColumnName) > 0 {
		builder.WriteString(" WHERE ")
		filterColumnName := params.ColumnName
		if len(params.NameWrapper) > 0 {
			filterColumnName = WrapNigger(filterColumnName, params.NameWrapper)
		}
		builder.WriteString(filterColumnName)
		builder.WriteString(" = $1")
	}

	return builder.String()
}

//endregion

//region Select query

// Если вы передаете ColumnName, тогда аргумент для него должен быть первым в списке аргументов
func GetSelectQuery(params QueryConfig) string {
	var builder strings.Builder
	additionalSymbols := 11
	totalSymbols := len(params.TableName) + len(params.ColumnName) + additionalSymbols
	builder.Grow(totalSymbols)

	tagName := stdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	builder.WriteString("SELECT ")

	typeOfN := TransformToNonRefType(params.ItemToAdd)

	numOfFields := typeOfN.NumField()

	builder.Grow(numOfFields * (4 + 2*len(params.NameWrapper)))

	isFieldDb := false
	isPrevFieldDb := isFieldDb

	for i := range numOfFields {
		tag := typeOfN.Field(i).Tag.Get(tagName)

		isPrevFieldDb = isFieldDb
		isFieldDb = len(tag) > 0 && !slices.Contains(params.ExcludedTags, tag)

		if isFieldDb && isPrevFieldDb {
			builder.WriteString(", ")
		}

		if isFieldDb {
			if len(params.NameWrapper) > 0 {
				tag = WrapNigger(tag, params.NameWrapper)
			}
			builder.WriteString(tag)
		}

	}

	builder.WriteString(" FROM ")

	tbname := params.TableName
	if len(params.NameWrapper) > 0 {
		tbname = WrapNigger(params.TableName, params.NameWrapper)
	}

	builder.WriteString(tbname)

	if len(params.ColumnName) > 0 {
		builder.WriteString(" WHERE ")
		filterColumnName := params.ColumnName
		if len(params.NameWrapper) > 0 {
			filterColumnName = WrapNigger(filterColumnName, params.NameWrapper)
		}
		builder.WriteString(filterColumnName)
		builder.WriteString(" = $1")
	}

	return builder.String()
}

//endregion

// region Delete query

func GetDeleteQuery(params QueryConfig) string {
	tableName := params.TableName
	columnName := params.ColumnName

	if len(params.NameWrapper) > 0 {
		tableName = WrapNigger(tableName, params.NameWrapper)
	}

	if len(columnName) > 0 {
		if len(params.NameWrapper) > 0 {
			columnName = WrapNigger(columnName, params.NameWrapper)
		}
		return "DELETE FROM " + tableName + " WHERE " + columnName + " = $1"
	}

	return "DELETE FROM " + tableName
}

//endregion

//endregion

//region Share funcs

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

//endregion

//region Query Config Change Methods

func (q QueryConfig) ChangeTable(tableName string, item any) QueryConfig {
	query := QueryConfig{
		TableName:   tableName,
		NameWrapper: q.NameWrapper,
		ColumnName:  q.ColumnName,
		TagName:     q.TagName,
		ItemToAdd:   item,
	}
	return *requiredProcessing(&query, &q)
}

func (q QueryConfig) ChangeColumnName(columnName string) QueryConfig {
	query := QueryConfig{
		TableName:   q.TableName,
		NameWrapper: q.NameWrapper,
		ColumnName:  columnName,
		TagName:     q.TagName,
		ItemToAdd:   q.ItemToAdd,
	}
	return *requiredProcessing(&query, &q)
}

func (q QueryConfig) ChangeExcludedTags(excludedTags ...string) QueryConfig {
	query := QueryConfig{
		TableName:   q.TableName,
		NameWrapper: q.NameWrapper,
		ColumnName:  q.ColumnName,
		TagName:     q.TagName,
		ItemToAdd:   q.ItemToAdd,
	}
	q.ExcludedTags = excludedTags
	return *requiredProcessing(&query, &q)
}

func (q QueryConfig) ChangeItem(item any) QueryConfig {
	query := QueryConfig{
		TableName:   q.TableName,
		NameWrapper: q.NameWrapper,
		ColumnName:  q.ColumnName,
		TagName:     q.TagName,
		ItemToAdd:   item,
	}
	return *requiredProcessing(&query, &q)
}

func (q QueryConfig) ChangeTagName(tagName string) QueryConfig {
	query := QueryConfig{
		TableName:   q.TableName,
		NameWrapper: q.NameWrapper,
		ColumnName:  q.ColumnName,
		TagName:     tagName,
		ItemToAdd:   q.ItemToAdd,
	}
	return *requiredProcessing(&query, &q)
}

func (q QueryConfig) ChangeNameWrapper(wrapper string) QueryConfig {
	query := QueryConfig{
		TableName:   q.TableName,
		NameWrapper: wrapper,
		ColumnName:  q.ColumnName,
		TagName:     q.TagName,
		ItemToAdd:   q.ItemToAdd,
	}
	return *requiredProcessing(&query, &q)
}

func requiredProcessing(new *QueryConfig, old *QueryConfig) *QueryConfig {
	var newExcTags []string
	if len(old.ExcludedTags) > 0 {
		newExcTags = make([]string, len(old.ExcludedTags))
		copy(newExcTags, old.ExcludedTags)
	}
	new.ExcludedTags = newExcTags
	return new
}

//endregion

type QueriesCacher struct {
	cacheInsert map[reflect.Type]*string
	cacheUpdate map[reflect.Type]*string
	cacheSelect map[reflect.Type]*string
}

func (qch *QueriesCacher) GetInsertQuery(q QueryConfig) string {

	return ""
}
