package gosql

import (
	"context"
	"database/sql"
	"reflect"
)

type DbHandler interface {
	Select(dest []any, query string, args ...any) error
	Insert(query string, args ...any) (id int, err error)
	Exec(query string, args ...any) (int, error)
	SelectContext(context context.Context, dest []any, query string, args ...any) error
	InsertContext(context context.Context, query string, args ...any) (id int, err error)
	ExecContext(context context.Context, query string, args ...any) (int, error)
}

type DB struct {
	Id             string
	TagName        string
	Handler        *DbHandler
	UseCachedFuncs bool
}

type StdDbHandler struct {
	Db *sql.DB
}

func (hn *StdDbHandler) Select(context context.Context, dest []any, query string, args ...any) error {
	destType := TransformToNonRefType(reflect.TypeOf(dest).Elem())

	rows, err := hn.Db.QueryContext(context, query, args...)

	if err != nil {
		return err
	}

	for rows.Next() {
		destValue := reflect.Zero(destType)
		err = rows.Scan()
		dest = append(dest, destValue)
	}

	return err
}
