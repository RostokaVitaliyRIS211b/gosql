package gosql

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/RostokaVitaliyRIS211b/gosql/sqlreflect"
	"github.com/RostokaVitaliyRIS211b/gosql/sqlstrings"
)

type DbHandler interface {
	Select(dest any, query string, queryConfig sqlstrings.QueryConfig, args ...any) error
	Insert(query string, queryConfig sqlstrings.QueryConfig, args ...any) (id int, err error)
	Exec(query string, queryConfig sqlstrings.QueryConfig, args ...any) (int, error)
	SelectContext(context context.Context, dest any, query string, queryConfig sqlstrings.QueryConfig, args ...any) error
	InsertContext(context context.Context, query string, queryConfig sqlstrings.QueryConfig, args ...any) (id int, err error)
	ExecContext(context context.Context, query string, queryConfig sqlstrings.QueryConfig, args ...any) (int, error)
}

type DB struct {
	Id             string
	handler        DbHandler
	handlerMutex   sync.RWMutex
	useCachedFuncs *atomic.Bool
}

type StdDbHandler struct {
	db           *sql.DB
	scanner      sqlreflect.Scanner
	scannerMutex sync.RWMutex
	dbMutex      sync.RWMutex
}

//region StdDbHandler Realization

func (stdh *StdDbHandler) SelectContext(context context.Context, dest any, query string, queryConfig sqlstrings.QueryConfig, args ...any) error {

	stdh.dbMutex.RLock()
	rows, err := stdh.db.QueryContext(context, query, args...)
	stdh.dbMutex.RUnlock()

	if err != nil {
		return err
	}

	stdh.scannerMutex.RLock()
	defer stdh.scannerMutex.RUnlock()
	err = stdh.scanner.Scan(dest, rows, queryConfig)

	return err
}

func (stdh *StdDbHandler) InsertContext(context context.Context, query string, y_ sqlstrings.QueryConfig, args ...any) (id int, err error) {

	stdh.dbMutex.RLock()
	defer stdh.dbMutex.RUnlock()
	err = stdh.db.QueryRowContext(context, query, args...).Scan(&id)

	return id, err

}

func (stdh *StdDbHandler) ExecContext(context context.Context, query string, _ sqlstrings.QueryConfig, args ...any) (int, error) {
	stdh.dbMutex.RLock()
	defer stdh.dbMutex.RUnlock()
	res, err := stdh.db.ExecContext(context, query, args...)
	aff, _ := res.RowsAffected()
	return (int)(aff), err
}

// dest должен быть указателем на slice
// ======================================================================================
// dest should be a pointer to slice
func (stdh *StdDbHandler) Select(dest any, query string, queryConfig sqlstrings.QueryConfig, args ...any) error {

	stdh.dbMutex.RLock()
	rows, err := stdh.db.Query(query, args...)
	stdh.dbMutex.RUnlock()

	if err != nil {
		return err
	}

	stdh.scannerMutex.RLock()
	defer stdh.scannerMutex.RUnlock()
	err = stdh.scanner.Scan(dest, rows, queryConfig)

	return err
}
func (stdh *StdDbHandler) Insert(query string, queryConfig sqlstrings.QueryConfig, args ...any) (id int, err error) {

	stdh.dbMutex.RLock()
	defer stdh.dbMutex.RUnlock()
	err = stdh.db.QueryRow(query, args...).Scan(&id)

	return id, err

}

func (stdh *StdDbHandler) Exec(query string, _ sqlstrings.QueryConfig, args ...any) (int, error) {
	stdh.dbMutex.RLock()
	defer stdh.dbMutex.RUnlock()
	res, err := stdh.db.Exec(query, args...)
	aff, _ := res.RowsAffected()
	return (int)(aff), err
}

// Этот метод блокирует вызывающую горутину пока Scanner не станет доступен для записи
// ======================================================================================
// This method blocks the calling goroutine until the Scanner becomes writable.
func (stdh *StdDbHandler) ChangeScanner(sc sqlreflect.Scanner) {
	defer stdh.scannerMutex.Unlock()
	stdh.scannerMutex.Lock()
	stdh.scanner = sc
}

// Этот метод блокирует вызывающую горутину пока db не станет доступна для записи
// ======================================================================================
// This method blocks the calling goroutine until the db becomes writable.
func (stdh *StdDbHandler) ChangeDb(db *sql.DB) {
	defer stdh.dbMutex.Unlock()
	stdh.dbMutex.Lock()
	stdh.db = db
}

func GetStdDbHandler(db *sql.DB) DbHandler {
	handler := &StdDbHandler{
		db:      db,
		scanner: sqlreflect.GetScanner(),
	}
	return handler
}

//endregion

//region DB Methods

func GetDb(db *sql.DB, id string) *DB {
	return &DB{
		handler: GetStdDbHandler(db),
		Id:      id,
	}
}

// dest должен быть указателем на slice
// ======================================================================================
// dest should be a pointer to slice
func (db *DB) SelectQuery(query string, queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {
	return db.SelectQueryContext(context.Background(), query, queryConfig, dest, args...)
}

// dest должен быть указателем на slice
// ======================================================================================
// dest should be a pointer to slice
func (db *DB) SelectQueryContext(context context.Context, query string, queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {
	db.handlerMutex.RLock()
	defer db.handlerMutex.RUnlock()

	return db.handler.SelectContext(context, dest, query, queryConfig, args...)
}

// dest должен быть указателем на slice
// ======================================================================================
// dest should be a pointer to slice
func (db *DB) Select(queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {
	return db.SelectContext(context.Background(), queryConfig, dest, args...)
}

// dest должен быть указателем на slice
// ======================================================================================
// dest should be a pointer to slice
func (db *DB) SelectContext(context context.Context, queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {
	db.handlerMutex.RLock()
	defer db.handlerMutex.RUnlock()

	query := ""

	if db.useCachedFuncs.Load() {
		query = sqlstrings.GetSelectQueryCached(queryConfig)
	} else {
		query = sqlstrings.GetSelectQuery(queryConfig)
	}
	return db.handler.SelectContext(context, dest, query, queryConfig, args...)
}

// dest должен быть указателем на структуру
// ======================================================================================
// dest should be a pointer to struct
func (db *DB) Get(queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {
	return db.GetContext(context.Background(), queryConfig, dest, args)
}

// dest должен быть указателем на структуру
// ======================================================================================
// dest should be a pointer to struct
func (db *DB) GetContext(context context.Context, queryConfig sqlstrings.QueryConfig, dest any, args ...any) error {

	if len(queryConfig.ColumnName) == 0 {
		return errors.New("queryConfig parameter ColumnName must be specified")
	}

	tdest := reflect.TypeOf(dest)

	if tdest.Kind() != reflect.Pointer {
		return errors.New("dest must be a pointer")
	}

	tstruct := tdest.Elem()

	if tstruct.Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to the struct")
	}

	slice := reflect.MakeSlice(tstruct, 0, 0)

	query := ""

	if db.useCachedFuncs.Load() {
		query = sqlstrings.GetSelectQueryCached(queryConfig)
	} else {
		query = sqlstrings.GetSelectQuery(queryConfig)
	}

	db.handlerMutex.RLock()
	err := db.handler.SelectContext(context, slice, query, queryConfig, args)
	db.handlerMutex.RUnlock()

	length := slice.Len()
	if length == 0 {
		return errors.New("result set is empty")
	}

	if length > 1 {
		return errors.New("result set have more than 1 element")
	}

	val := reflect.ValueOf(dest)
	if val.CanSet() {
		newVal := reflect.New(tstruct)
		newVal.Elem().Set(slice.Index(0))
		val.Set(newVal)
	}

	return err
}

func (db *DB) Insert(queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {
	return db.InsertContext(context.Background(), queryConfig, args...)
}

func (db *DB) InsertContext(context context.Context, queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {
	db.handlerMutex.RLock()
	defer db.handlerMutex.RUnlock()

	query := ""

	if db.useCachedFuncs.Load() {
		query = sqlstrings.GetInsertQueryCached(queryConfig)
	} else {
		query = sqlstrings.GetInsertQuery(queryConfig)
	}

	return db.handler.InsertContext(context, query, queryConfig, args...)
}

func (db *DB) Update(queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {
	return db.UpdateContext(context.Background(), queryConfig, args)
}

func (db *DB) UpdateContext(context context.Context, queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {

	query := ""

	if db.useCachedFuncs.Load() {
		query = sqlstrings.GetUpdateQueryCached(queryConfig)
	} else {
		query = sqlstrings.GetUpdateQuery(queryConfig)
	}

	db.handlerMutex.RLock()
	res, err := db.handler.ExecContext(context, query, queryConfig, args...)
	db.handlerMutex.RUnlock()

	return res, err
}

func (db *DB) Delete(queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {
	return db.DeleteContext(context.Background(), queryConfig, args)
}

func (db *DB) DeleteContext(context context.Context, queryConfig sqlstrings.QueryConfig, args ...any) (int, error) {

	query := sqlstrings.GetDeleteQuery(queryConfig)

	db.handlerMutex.RLock()
	defer db.handlerMutex.RUnlock()
	res, err := db.handler.ExecContext(context, query, queryConfig, args...)

	return res, err
}

func (db *DB) Exec(query string, args ...any) (int, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *DB) ExecContext(context context.Context, query string, args ...any) (int, error) {
	db.handlerMutex.RLock()
	defer db.handlerMutex.RUnlock()
	res, err := db.handler.ExecContext(context, query, sqlstrings.QueryConfig{}, args...)

	return res, err
}

func (db *DB) UseCachedFuncs(b bool) {
	db.useCachedFuncs.Store(b)
}

//endregion
