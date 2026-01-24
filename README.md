# gosql

[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/RostokaVitaliyRIS211b/gosql/main/LICENSE)

If you are English men scroll way down

gosql - это библиотека, которая предоставляет набор инструментов для более удобной работы с
sql.DB из database/sql, также есть возможность использовать любой драйвер для sql.DB. При этом не нужно
писать запросы руками, достаточно лишь создать модель базы данных, в которой у структур описывающих таблицы, теги
полей содержат названия столбцов таблицы.

## install

go get github.com/RostokaVitaliyRIS211b/gosql

## usage

Ниже приведен пример использования библиотеки

```go
package main

import(
    "database/sql"
    "fmt"
    "log"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/RostokaVitaliyRIS211b/gosql"
    "github.com/RostokaVitaliyRIS211b/gosql/sqlstrings"
)

type User struct {
    Id          int    `db:"Id"`
    Name        string `db:"Name"`
    Password    string `db:"Password"`
    Description string `db:"Description"`
    Roles       []int32 // Поля без соотвествующего тега игнорируются
}

var dbSheme = `
    CREATE TABLE Users(
        "Id" serial,
        "Name" text,
        "Password" text,
        "Description" text
    );
    `
func main(){
    //Стандартным образом открываем соединение с базой данных
    db,err:=sql.Open("pgx","user=postgres dbname=test sslmode=disable")
    if err != nil{
        log.Fatalln(err)
    }

    // Получаем обертку над sql.DB
    DB:=gosql.GetDb(db,"pgx")

    DB.Exec(dbSheme)

    // Создаем конфигурацию запроса
    var insertQC = sqlstrings.QueryConfig{
	    TableName:    "",                   // Имя таблицы
	    NameWrapper:  "\"",                 // Указываем строку в которую будут обёрнуты имя таблицы
											// и названия столбцов. Пример имя таблицы Users -> "Users"
	    ColumnName:   "Id",                 // Если нужно получить Id добавленной записи или
											// любой другой столбец, то необходимо указать имя данного
											// столбца
	    TagName:      "db",                 // Указываем имя тега, значением которого является имя
											// столбца таблицы
	    Item:         nil,                  // Запись для добавления/обновления/получения/удаления
	    ExcludedTags: []string{"Id"},       // Указываем имена столбцов которые нужно исключить
											// из запроса, в данном случае нам не нужно при добавлении
											// указывать Id
	    QueryType:    sqlstrings.INSERT,    // Тип запроса INSERT
    }

    user := User{Name:"test",Password:"123",Description:"нигер"}

    // Методы QueryConfig работают исключительно с копиями,
	// для того чтобы оригинальная конфигурация не менялась,
	// возвращаются тоже копии конфигурации
    currentConfig := insertQC.ChangeTable("Users",user)

    // Обертка сама сгенерирует строку запроса, и возьмёт аргументы для добавления
	// из переданной структуры, если включен маппинг
    // Итоговая строка запроса будет выглядеть так: INSERT INTO "Users" 
	// ("Name","Password","Description") VALUES ($1,$2,$3) RETURNING "Id"
    // Шаблон такой INSERT INTO [NameWrapper]TableName[NameWrapper] 
    // ([NameWrapper]FieldTag1[NameWrapper] , [NameWrapper]FieldTag2[NameWrapper], 
	// ...[NameWrapper]FieldTagN[NameWrapper]) VALUES ($1,$2, ... $n)
	// [RETURNING [NameWrapper]ColumnName[NameWrapper]]
    idx,_:=DB.Insert(currentConfig)

    user.Name = "Test2"
    user.Password="admin"
    user.Description="черный"

    idx2,_:=DB.Insert(currentConfig.ChangeItem(user))

    // Получение нескольких записей

    var selectQC = sqlstrings.QueryConfig{
		TableName:    "",                   // Имя таблицы
	    NameWrapper:  "\"",                 // Указываем строку в которую будут обёрнуты имя таблицы
											//  и названия столбцов. Пример имя таблицы Users -> "Users"
	    ColumnName:   "",                   // ColumnName используется для фильтрации результатов при
											// типе запроса SELECT
	    TagName:      "db",                 // Указываем имя тега, значением которого является
											// имя столбца таблицы
	    Item:         nil,                  // Запись для добавления/обновления/получения/удаления
	    ExcludedTags: []string{},           // Указываем имена столбцов которые нужно
											// исключить из запроса
	    QueryType:    sqlstrings.SELECT,    // Тип запроса SELECT
    }

    users:=[]User{}

    currentSelect := selectQC.ChangeTable("Users",User{})

    // Обязательно нужно передавать указатель на срез,
	// обертка сама сгенерирует строку запроса, результаты записываются в переданный срез
    // Итоговая строка будет такой: SELECT "Name", "Password", "Description" FROM "Users"

    DB.Select(currentSelect,&users)

    // Шаблон строки запроса SELECT [NameWrapper]FieldTag1[NameWrapper] ,
	// [NameWrapper]FieldTag2[NameWrapper], ...[NameWrapper]FieldTagN[NameWrapper] FROM [NameWrapper]
    // TableName[NameWrapper] [WHERE [NameWrapper]ColumnName[NameWrapper] = $1]

    user1,user2:=users[0],users[1]

    fmt.Printf("%#v\n%#v", user1, user2)


    // Также можно получить одну запись для этого необходимо указать ColumnName

    var whereIdQC = sqlstrings.QueryConfig{
	    TableName:    "",                   // Имя таблицы
	    NameWrapper:  "\"",                 // Указываем строку в которую будут обёрнуты имя таблицы 
											// и названия столбцов. Пример имя таблицы Users -> "Users"
	    ColumnName:   "Id",                 // ColumnName используется для фильтрации результатов,
											// в конец строки запроса будет добавлено WHERE Id = $1
	    TagName:      "db",                 // Указываем имя тега, значением которого является
											// имя столбца таблицы
	    Item:         nil,                  // Запись для добавления/обновления/получения/удаления
	    ExcludedTags: []string{},           // Указываем имена столбцов которые нужно исключить
											// из запроса
	    QueryType:    sqlstrings.SELECT,    // Тип запроса SELECT
    }

    currentGet:=whereIdQC.ChangeTable("Users",User{})

    user = User{}

    // Обязательно нужно передать указатель на структуру,
	// а в аргументах передать значение для получения записи
    DB.Get(currentGet,&user,idx)

    fmt.Printf("%#v", user)

    // Обновление записи

    user.Name = "Jamal"
    user.Password = "nutella"
    user.Description = "choco"

    var updateQC = sqlstrings.QueryConfig{
	    TableName:    "",                   // Имя таблицы
	    NameWrapper:  "\"",                 // Указываем строку в которую будут обёрнуты имя таблицы 
											// и названия столбцов. Пример имя таблицы Users -> "Users"
	    ColumnName:   "Id",                 // ColumnName используется для обновления
											// определенной записи,
											// в конец строки запроса будет добавлено WHERE Id = $1
	    TagName:      "db",                 // Указываем имя тега, значением которого является
											// имя столбца таблицы
	    Item:         nil,                  // Запись для добавления/обновления/получения/удаления
	    ExcludedTags: []string{"Id"},       // Указываем имена столбцов которые нужно исключить
											// из запроса,
											// в данном случае нам не нужно обновлять Id
	    QueryType:    sqlstrings.UPDATE,    // Тип запроса UPDATE
    }

    currentUpdate := updateQC.ChangeTable("Users",user)

    // Если указан ColumnName и включен маппинг тогда при передаче
	// аргументов первым будет поле структуры с значением тега равным ColumnName
    // Текущий порядок передачи аргументов Id,Name,Password,Description
    // Обертка сама сгенерирует строку запроса, а аргументы будут взяты из переданной структуры,
	//  если включен маппинг
    DB.Update(currentUpdate)

    // Шаблон строки UPDATE [NameWrapper]TableName[NameWrapper] SET
	//  [NameWrapper]FieldTag1[NameWrapper] = $1[2], ... [NameWrapper]FieldTagN[NameWrapper] = $n
    // [WHERE [NameWrapper]ColumnName[NameWrapper] = $1]

    DB.Get(currentGet,&user,idx)

    fmt.Printf("%#v", user)

    //Удаление записи

    var deleteQC = sqlstrings.QueryConfig{
	    TableName:    "",                   // Имя таблицы
	    NameWrapper:  "\"",                 // Указываем строку в которую будут обёрнуты имя таблицы 
											// и названия столбцов. Пример имя таблицы Users -> "Users"
	    ColumnName:   "Id",                 // ColumnName используется для обновления
											// определенной записи,
											// в конец строки запроса будет добавлено WHERE Id = $1
	    TagName:      "db",                 // Указываем имя тега, значением которого является имя
											// столбца таблицы
	    Item:         nil,                  // Запись для добавления/обновления/получения/удаления
	    ExcludedTags: []string{},           // Указываем имена столбцов которые нужно исключить
											// из запроса
	    QueryType:    sqlstrings.DELETE,    // Тип запроса DELETE
    }

    currentDelete := deleteQC.ChangeTable("Users",nil)

    DB.Delete(currentDelete, idx)

    //Выполнение строки запроса

    query:="DELETE FROM \"Users\" WHERE \"Id\" = $1"

    DB.Exec(query,idx2)

    //Различные методы обертки

	// Можно отключить маппинг, тогда при запросах INSERT и UPDATE аргументы никогда не будут
	// взяты из переданной структуры
    DB.SetMapper(nil)

	// Можно отключить использование кэша при генерации строк запросов, если вам дорога память
    DB.UseCachedFuncs(false)

	// Можно сменить обработчика базы данных
    DB.ChangeHandler(nil) 
}
```

================================================================================================================
================================================================================================================

English men read this

gosql - this is a library that provides a set of tools for more convenient work with
sql.DB from database/sql, and it is also possible to use any driver for sql.DB. At the same time, you do not need
to write queries by hand, it is enough just to create a database model in which the structures describing the tables
have field tags containing the names of the columns of the table.

## install

go get github.com/RostokaVitaliyRIS211b/gosql

## usage

The following is an example of using the library

```go
package main

import(
    "database/sql"
    "fmt"
    "log"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/RostokaVitaliyRIS211b/gosql"
    "github.com/RostokaVitaliyRIS211b/gosql/sqlstrings"
)

type User struct {
    Id          int    `db:"Id"`
    Name        string `db:"Name"`
    Password    string `db:"Password"`
    Description string `db:"Description"`
    Roles []int32 // Fields without an appropriate tag are ignored
}

var dbSheme = `
    CREATE TABLE Users(
        "Id" serial,
        "Name" text,
        "Password" text,
        "Description" text
    );
    `
func main(){
	// Opening a database connection in the standard way
    db,err:=sql.Open("pgx","user=postgres dbname=test sslmode=disable")
    if err != nil{
        log.Fatalln(err)
    }

    // Getting a wrapper over sql.DB
	DB:=gosql.GetDb(db,"pgx")

    DB.Exec(dbSheme)

    // Creating the request configuration
    var insertQC = sqlstrings.QueryConfig{
	    TableName: "", 					// Table name
	    NameWrapper: "\"", 				// Specify the string in which the table name will be wrapped
										// and column names. Example table name Users -> "Users"
	    ColumnName: "Id", 				// If you want to get the Id of the added record or
										// any other column, you must specify the name of this column
	    TagName: "db", 					// Specify the name of the tag, the value of which is the name
										//  of the table column
	    Item: nil, 						// Entry to add/update/receive/delete
	    ExcludedTags: []string{"Id"}, 	// Specifying the names of the columns to exclude from the query,
										// in this case, we don't need to specify the Id when adding
	    QueryType: sqlstrings.INSERT,	// Query type INSERT
    }

    user := User{Name:"test",Password:"123",Description:"niger"}

    // The QueryConfig methods work exclusively with copies,
	// so that the original configuration does not change,
	// copies of the currentConfig configuration are also returned
	currentConfig := insertQC.ChangeTable("Users",user)

    // The wrapper will generate the query string itself, and will take arguments to add
	// from the passed structure if mapping is enabled
    // The final query string will look like this: INSERT INTO "Users" 
	// ("Name","Password","Description") VALUES ($1,$2,$3) RETURNING "Id"
	// The template is INSERT INTO [NameWrapper]TableName[NameWrapper] 
    // ([NameWrapper]FieldTag1[NameWrapper] , [NameWrapper]FieldTag2[NameWrapper], 
	// ...[NameWrapper]FieldTagN[NameWrapper]) VALUES ($1,$2, ... $n)
	// [RETURNING [NameWrapper]ColumnName[NameWrapper]]
    idx,_:=DB.Insert(currentConfig)

    user.Name = "Test2"
    user.Password="admin"
    user.Description="black"

    idx2,_:=DB.Insert(currentConfig.ChangeItem(user))

    // Getting multiple records

    var selectQC = sqlstrings.QueryConfig{
		TableName: "", 					// Table name
	    NameWrapper: "\"", 				// Specify the row in which the table name and column names
										// will be wrapped.
										// Example name of the Users table -> "Users"
	    ColumnName: "", 				// ColumnName is used to filter results with the
										// SELECT query type
	    TagName: "db", 					// Specify the name of the tag, the value of which is
										// the name of the table column
	    Item: nil, 						// Entry to add/update/receive/delete
	    ExcludedTags: []string{}, 		// Specifying the names of the columns to exclude from the query
	    QueryType: sqlstrings.SELECT,	// Query type SELECT
    }

    users:=[]User{}

    currentSelect := selectQC.ChangeTable("Users",User{})

    // It is necessary to pass a pointer to the slice,
	// the wrapper itself will generate a query string, the results are written to the passed slice
	// The final string will be as follows: SELECT "Name", "Password", "Description" FROM "Users"

    DB.Select(currentSelect,&users)

    // Query string template SELECT [NameWrapper]FieldTag1[NameWrapper] ,
	// [NameWrapper]FieldTag2[NameWrapper], ...[NameWrapper]FieldTagN[NameWrapper] FROM [NameWrapper]
    // TableName[NameWrapper] [WHERE [NameWrapper]ColumnName[NameWrapper] = $1]

    user1,user2:=users[0],users[1]

    fmt.Printf("%#v\n%#v", user1, user2)


    // You can also get one record, you need to specify columnName for this.

    var whereIdQC = sqlstrings.QueryConfig{
	    TableName: "", 					// Table name
	    NameWrapper: "\"", 				// Specify the row in which the table name and
										// column names will be wrapped.
										// Example name of the Users table -> "Users"
	    ColumnName: "Id", 				// ColumnName is used to filter the results,
										// WHERE Id = $1 will be added to the end of the query string
	    TagName: "db", 					// Specify the name of the tag, the value of which is
										//  the name of the table column
	    Item: nil, 						// Entry to add/update/receive/delete
	    ExcludedTags: []string{}, 		// Specifying the names of the columns to exclude from the query
	    QueryType: sqlstrings.SELECT,	// Query type SELECT
    }

    currentGet:=whereIdQC.ChangeTable("Users",User{})

    user = User{}

    // You must pass a pointer to the structure,
	// and pass the value in the arguments to get the record
    DB.Get(currentGet,&user,idx)

    fmt.Printf("%#v", user)

    // Record update

    user.Name = "Jamal"
    user.Password = "nutella"
    user.Description = "choco"

    var updateQC = sqlstrings.QueryConfig{
	    TableName: "", 					// Table name
	    NameWrapper: "\"", 				// Specify the row in which the table name and
										// column names will be wrapped.
										// Example name of the Users table -> "Users"
	    ColumnName: "Id", 				// ColumnName is used to update a specific record,
										// WHERE Id = $1 will be added to the end of the query string
	    TagName: "db", 					// Specify the name of the tag, the value of which
										// is the name of the table column
	    Item: nil, 						// Entry to add/update/receive/delete
	    ExcludedTags: []string{"Id"}, 	// Specifying the names of the columns to exclude from the query,
										// in this case, we don't need to update the Id
	    QueryType: sqlstrings.UPDATE,	// Query type UPDATE
    }

    currentUpdate := updateQC.ChangeTable("Users",user)

    // If ColumnName is specified and mapping is enabled, then when passing
	// arguments, the structure field with the tag value equal to columnName will be the first.
    // The current order of passing the arguments Id,Name,Password,Description
    // The wrapper will generate the query string itself,
	// and the arguments will be taken from the passed structure if mapping is enabled
    DB.Update(currentUpdate)

    // String template UPDATE [NameWrapper]TableName[NameWrapper] SET 
	// [NameWrapper]FieldTag1[NameWrapper] = $1[2],  ... [ NameWrapper]FieldTagN[NameWrapper] = $n
    // [WHERE [NameWrapper]ColumnName[NameWrapper] = $1]

    DB.Get(currentGet,&user,idx)

    fmt.Printf("%#v", user)

    //Deleting an entry

    var deleteQC = sqlstrings.QueryConfig{
	    TableName: "", 					// Table name
	    NameWrapper: "\"", 				// Specify the row in which the table name and
										// column names will be wrapped.
										// Example name of the Users table -> "Users"
	    ColumnName: "Id", 				// ColumnName is used to update a specific record,
										// WHERE Id = $1 will be added to the end of the query string
	    TagName: "db", 					// Specify the name of the tag, the value of which is
										// the name of the table column
	    Item: nil, 						// Entry to add/update/receive/delete
	    ExcludedTags: []string{}, 		// Specifying the names of the columns to exclude from the query
	    QueryType: sqlstrings.DELETE,	// Query type DELETE
    }

    currentDelete := deleteQC.ChangeTable("Users",nil)

    DB.Delete(currentDelete, idx)

    //Query String Execution

    query:="DELETE FROM \"Users\" WHERE \"Id\" = $1"

    DB.Exec(query,idx2)

    //Different wrapper methods

	// You can disable mapping, so that when you request INSERT and UPDATE,
	// the arguments will never be taken from the passed structure.
    DB.SetMapper(nil)

	// You can disable cache usage when generating query strings if you value memory.
    DB.UseCachedFuncs(false)

	// You can change the database handler
    DB.ChangeHandler(nil)
```
