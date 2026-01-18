package gosql

import (
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const StdTagName = "dbcn"

//region Queries

type QueryType int

const (
	INSERT = iota
	UPDATE
	SELECT
	DELETE
)

// TableName - Имя таблицы
// NameWrapper - нужен для оборачивания имен столбцов и таблиц, если указать например " то имя будет "SomeName" ;
// ColumnName - (Insert) нужен для того чтобы можно было вернуть Id добавленной записи, если указан то в конец строки добавится:  RETURNING IdColumnName ;
// (Update) нужен для того чтобы обновить только определенные записи, в конец строки будет добавлено WHERE ColumnName = $1 ;
// (Select) нужен для того чтобы отфильтровать получаемые данные, в конец строки будет добавлено WHERE ColumnName = $1  ;
// (Delete) нужен для того чтобы удалить только определенные записи, в конец строки будет добавлено WHERE ColumnName = $1  ;
// TagName - нужен для того если вы используете нестандартный тег для полей структуры, тогда вместо стандартного dbcn будет использоваться указанный тег ;
// ItemToAdd - структура содержащая поля с тегами, значение которых соответсвует названиям столбцов таблицы ;
// ExcludedTags - список тегов которые вы хотите исключить при созданнии строки, например Id, тогда конечная строка не будет содержать данного столбца ;
// =========================================================================================================================================================
// TableName is the name of the table
// NameWrapper is needed to wrap the names of columns and tables, if you specify, for example, "then the name will be "SomeName"
// columnName - (Insert) is needed so that you can return the Id of the added record, if specified, then the following will be added to the end of the row: RETURNING IdColumnName ;
// (Update) is needed in order to update only certain records, WHERE columnName = $1 will be added to the end of the line ;
// (Select) is needed in order to filter the received data, WHERE columnName = $1 will be added to the end of the line  ;
// (Delete) is needed in order to delete only certain entries, WHERE columnName = $1 will be added to the end of the line;
// TagName is needed so that if you use a non-standard tag for the fields of the structure, then the specified tag will be used instead of the standard dbcn ;
// ItemToAdd - a structure containing fields with tags, the value of which is corresponds to the column names of the table ;
// ExcludedTags - a list of tags that you want to exclude when creating a row, for example, Id, then the final row will not contain this column. ;
type QueryConfig struct {
	TableName    string
	NameWrapper  string
	ColumnName   string
	TagName      string
	ItemToAdd    any
	ExcludedTags []string
}

// Возвращает строку указанного типа /
// Returns a string of the specified type
func GetQuery(params QueryConfig, queryType QueryType) string {
	switch queryType {
	case INSERT:
		return GetInsertQuery(params)
	case UPDATE:
		return GetUpdateQuery(params)
	case SELECT:
		return GetSelectQuery(params)
	case DELETE:
		return GetDeleteQuery(params)
	}
	return "this query type is not supported"
}

//region InsertQuery

// Возвращает строку запроса INSERT INTO TableName (ItemFieldTag1, ItemFieldTag2 ...) VALUES ($1,$2 ...) [RETURNING ColumnName] ,если указан ColumnName то в конец строки добавится: RETURNING IdColumnName, Порядок аргументов должен соотвествовать порядку полей в передаваемой структуре
// ============================================================================================================================================================
// Returns the INSERT INTO TableName (ItemFieldTag1, ItemFieldTag2 ...) VALUES ($1,$2 ...) [RETURNING ColumnName] query string ... If columnName is specified, then the following is added to the end of the line: RETURNING IdColumnName, the order of the arguments must match the order of the fields in the passed structure.
func GetInsertQuery(params QueryConfig) string {
	var builder strings.Builder

	tagName := StdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	typeOfN := ConversionValToNonRefType(params.ItemToAdd)

	numFields := typeOfN.NumField()

	additionalSymbols := 37
	totalSymbols := len(params.TableName) + len(params.ColumnName) + additionalSymbols + numFields*(4+2*len(params.NameWrapper))
	//Выделяем память под символы сразу
	builder.Grow(totalSymbols)

	builder.WriteString("INSERT")
	builder.WriteString(" INTO ")

	//Если указан NameWrapper то оборачиваем в него имя таблицы
	tbname := params.TableName
	if len(params.NameWrapper) > 0 {
		tbname = WrapNigger(params.TableName, params.NameWrapper)
	}
	builder.WriteString(tbname)

	builder.WriteString(" (")

	counter := 0

	isFieldDb := false
	isPrevFieldDb := isFieldDb

	//Проходим по всем полям переданной структуры
	for i := range numFields {
		//Читаем значение тега
		tag := typeOfN.Field(i).Tag.Get(tagName)

		isPrevFieldDb = isFieldDb
		isFieldDb = len(tag) > 0 && !slices.Contains(params.ExcludedTags, tag)

		if isFieldDb && isPrevFieldDb {
			builder.WriteString(", ")
		}

		//Если данное поле структуры имеет тег и тег не входит в список исключений, тогда добавляем содержимое тега в нашу строку НЕГРЫ! 14.01.2026
		if isFieldDb {
			name := tag
			if len(params.NameWrapper) > 0 {
				name = WrapNigger(name, params.NameWrapper)
			}
			builder.WriteString(name)
			counter++
		}

	}

	// формируем такую штуку VALUES ($1,$2 ....)
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

// Возвращает строку типа UPDATE TableName SET ItemFieldTag1=$1, ItemFieldTag2=$2 ... [WHERE ColumnName = $1] , eсли вы передаете ColumnName, тогда в конец строки будет добавлено WHERE ColumnName = $1 и аргумент для него должен быть первым в списке аргументов, для остального порядок аргументов должен соотвествовать порядку полей в передаваемой структуре
// ===============================================================================================================================
// Returns a string like UPDATE TableName SET ColumnName1=$1  ItemFieldTag2=$2 ... [WHERE ColumnName = $1] , if you pass ColumnName, then WHERE ColumnName = $1 will be added to the end of the string and the argument for it must be the first in the argument list. For the rest, the order of the arguments must match the order of the fields in the passed structure
func GetUpdateQuery(params QueryConfig) string {
	typeOfN := ConversionValToNonRefType(params.ItemToAdd)

	counter := 0

	numFields := typeOfN.NumField()

	var builder strings.Builder
	additionalSymbols := 11
	totalSymbols := len(params.TableName) + len(params.ColumnName) + additionalSymbols + numFields*(4+2*len(params.NameWrapper))
	builder.Grow(totalSymbols)

	tagName := StdTagName

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

// Возвращает строку типа SELECT ItemFieldTag1, ItemFieldTag2 ... FROM TableName [WHERE ColumnName = $1], если вы передаете ColumnName, в конец строки будет добавлено WHERE ColumnName = $1, аргумент для него должен быть первым в списке аргументов
// ==============================================================================================================================
// Returns a string of type SELECT ItemFieldTag1, ItemFieldTag2 ... FROM TableName, if you pass columnName, WHERE columnName = $1 will be added to the end of the line, the argument for it must be the first in the argument list.
func GetSelectQuery(params QueryConfig) string {
	var builder strings.Builder
	typeOfN := ConversionValToNonRefType(params.ItemToAdd)
	numOfFields := typeOfN.NumField()

	additionalSymbols := 11
	totalSymbols := len(params.TableName) + len(params.ColumnName) + additionalSymbols + numOfFields*(4+2*len(params.NameWrapper))

	builder.Grow(totalSymbols)

	tagName := StdTagName

	if len(params.TagName) > 0 {
		tagName = params.TagName
	}

	builder.WriteString("SELECT ")

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

// Возращает строку типа DELETE FROM TableName [WHERE ColumnName = $1], при указании ColumnName в конец строки добавляет WHERE ColumnName = $1
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

func ConversionValToNonRefType(value any) reflect.Type {
	typeOfVal := reflect.TypeOf(value)
	kind := typeOfVal.Kind()

	for kind == reflect.Pointer {
		typeOfVal = typeOfVal.Elem()
		kind = typeOfVal.Kind()
	}
	return typeOfVal
}

func ConversionTypeToNonRefType(t reflect.Type) reflect.Type {
	kind := t.Kind()

	for kind == reflect.Pointer {
		t = t.Elem()
		kind = t.Kind()
	}
	return t
}

func WrapNigger(n string, wrapper string) string {
	return wrapper + n + wrapper
}

//endregion

//region Query Config Change Funcs

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

//region Caching

var (
	insertQueryCache map[cacheKey]string
	selectQueryCache map[cacheKey]string
	updateQueryCache map[cacheKey]string
	cacheMutex       sync.RWMutex
)

type cacheKey struct {
	Type         reflect.Type
	TableName    string
	TagName      string
	ColumnName   string
	NameWrapper  string
	ExcludedTags string // отсортированная строка тегов
}

func GetCachedQuery(params QueryConfig, queryType QueryType) string {

	switch queryType {
	case INSERT:
		return GetInsertQueryCached(params)
	case UPDATE:
		return GetUpdateQueryCached(params)
	case SELECT:
		return GetSelectQueryCached(params)
	case DELETE:
		return GetDeleteQuery(params)
	}

	return "this query type is not supported"
}

func GetInsertQueryCached(params QueryConfig) string {
	itemType := ConversionValToNonRefType(params.ItemToAdd)

	key := cacheKey{
		Type:         itemType,
		TableName:    params.TableName,
		TagName:      params.TagName,
		ColumnName:   params.ColumnName,
		NameWrapper:  params.NameWrapper,
		ExcludedTags: getExcludedTagsKey(params.ExcludedTags),
	}

	cacheMutex.RLock()
	query, ok := insertQueryCache[key]
	cacheMutex.RUnlock()
	if ok {
		return query
	}

	query = GetInsertQuery(params)

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	insertQueryCache[key] = query

	return query
}

func GetUpdateQueryCached(params QueryConfig) string {
	itemType := ConversionValToNonRefType(params.ItemToAdd)

	key := cacheKey{
		Type:         itemType,
		TableName:    params.TableName,
		TagName:      params.TagName,
		ColumnName:   params.ColumnName,
		NameWrapper:  params.NameWrapper,
		ExcludedTags: getExcludedTagsKey(params.ExcludedTags),
	}

	cacheMutex.RLock()
	query, ok := updateQueryCache[key]
	cacheMutex.RUnlock()
	if ok {
		return query
	}

	query = GetUpdateQuery(params)

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	updateQueryCache[key] = query

	return query
}

func GetSelectQueryCached(params QueryConfig) string {
	itemType := ConversionValToNonRefType(params.ItemToAdd)

	key := cacheKey{
		Type:         itemType,
		TableName:    params.TableName,
		TagName:      params.TagName,
		ColumnName:   params.ColumnName,
		NameWrapper:  params.NameWrapper,
		ExcludedTags: getExcludedTagsKey(params.ExcludedTags),
	}

	cacheMutex.RLock()
	query, ok := selectQueryCache[key]
	cacheMutex.RUnlock()
	if ok {
		return query
	}

	query = GetSelectQuery(params)

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	selectQueryCache[key] = query

	return query
}

func getExcludedTagsKey(excluded []string) string {
	if len(excluded) == 0 {
		return ""
	}
	tags := make([]string, len(excluded))
	copy(tags, excluded)
	sort.Strings(tags)
	return strings.Join(tags, ",")
}

//endregion
