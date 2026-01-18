package gosql

import (
	"context"
	"database/sql"
	"sync/atomic"
)

type Scanner interface {
	Scan(dest any, rows RowScanner, queryConfig QueryConfig) error
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

	query := ""

	if hn.useCachedFuncs.Load() {
		query = GetSelectQueryCached(queryConfig)
	} else {
		query = GetSelectQuery(queryConfig)
	}

	rows, err := hn.db.QueryContext(context, query, args...)
	if err != nil {
		return err
	}

	err = hn.scanner.Scan(dest, rows, queryConfig)

	return err
}
