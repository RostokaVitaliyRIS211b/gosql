package gosql

type DbHandler interface {
	Select(query string, args ...any) error
	Insert(query string, args ...any)
	Exec(query string, args ...any) error
}

type DB struct {
	Id                     string
	UseTag                 string
	Handler                *DbHandler
	UseTypeNameAsTableName bool
}
