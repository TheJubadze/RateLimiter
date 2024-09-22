package iplists

type Repository interface {
	InsertNetwork(table, subnet string) error
	DeleteNetwork(table, subnet string) (bool, error)
	GetNetworks(table string) ([]string, error)
	IsNetworkExists(table, subnet string) (bool, error)
	Close() error
}
