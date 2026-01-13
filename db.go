package gosql

type DbHandler interface {
	Select(dest any, args ...any) error
	Exec(qury string, args []any) error
}

type DB struct {
	Id                     string
	Handler                *DbHandler
	UseTypeNameAsTableName bool
	UsedTag                string
}
