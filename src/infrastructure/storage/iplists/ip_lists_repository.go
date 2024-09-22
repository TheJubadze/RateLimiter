package iplistsrepository

import (
	"net"

	"github.com/TheJubadze/RateLimiter/infrastructure/storage/postgres"
	"github.com/TheJubadze/RateLimiter/interfaces/storage/database"
)

type Repository struct {
	db database.Database
}

func NewRepository(connString string) (*Repository, error) {
	db, err := postgresdb.NewDatabase(connString)
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (p *Repository) Close() error {
	return p.db.Close()
}

func (p *Repository) InsertNetwork(table, subnet string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}

	return p.db.Insert(table, ipNet.String())
}

func (p *Repository) DeleteNetwork(table, subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	return p.db.Delete(table, ipNet.String())
}

func (p *Repository) GetNetworks(table string) ([]string, error) {
	return p.db.GetAll(table)
}

func (p *Repository) IsNetworkExists(table, subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	return p.db.GetByValue(table, ipNet.String())
}
