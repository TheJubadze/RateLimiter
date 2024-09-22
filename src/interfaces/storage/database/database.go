package database

type Database interface {
	Insert(table string, value string) error
	Delete(table string, value string) (bool, error)
	GetAll(table string) ([]string, error)
	GetByValue(table string, value string) (bool, error)
	Close() error
}
