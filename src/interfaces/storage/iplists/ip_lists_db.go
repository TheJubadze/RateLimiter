package iplists

import "database/sql"

type DB interface {
	InsertNetwork(table, subnet string) error
	DeleteNetwork(table, subnet string) (bool, error)
	GetNetworks(table string) (*sql.Rows, error)
	IsNetworkExists(table, subnet string) (bool, error)
	Close() error
}
