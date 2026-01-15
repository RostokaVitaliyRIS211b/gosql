package gosql

import (
	"context"
	"database/sql"
	"reflect"
	"sync/atomic"
)

type Scanner interface {
	Scan(item any, tagName string, rows *sql.Rows, excludedTags []string) error
}

type DbHandler interface {
	Select(dest []any, queryConfig QueryConfig, args ...any) error
	Insert(queryConfig QueryConfig, args ...any) (id int, err error)
	Exec(queryConfig QueryConfig, args ...any) (int, error)
	SelectContext(context context.Context, dest []any, queryConfig QueryConfig, args ...any) error
	InsertContext(context context.Context, queryConfig QueryConfig, args ...any) (id int, err error)
	ExecContext(context context.Context, queryConfig QueryConfig, args ...any) (int, error)
	UseCachedFuncs(bool)
}

type DB struct {
	Id      string
	TagName string
	handler DbHandler
}

type StdDbHandler struct {
	db             *sql.DB
	scanner        Scanner
	useCachedFuncs *atomic.Bool
}

func (hn *StdDbHandler) SelectContext(context context.Context, dest []any, queryConfig QueryConfig, args ...any) error {
	ogType := reflect.TypeOf(dest).Elem()

	query := ""

	rows, err := hn.db.QueryContext(context, query, args...)

	if err != nil {
		return err
	}

	for rows.Next() {
		destValue := reflect.Zero(ogType).Interface()
		err = hn.scanner.Scan(destValue, queryConfig.TagName, rows, queryConfig.ExcludedTags)
		dest = append(dest, destValue)
	}

	return err
}
