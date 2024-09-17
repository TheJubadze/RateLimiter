package postgresdb

import (
	"database/sql"
	"net"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(connString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) InsertNetwork(table, subnet string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	_, err = p.db.Exec("INSERT INTO "+table+" (network) VALUES ($1)", ipNet.String())
	return err
}

func (p *PostgresDB) DeleteNetwork(table, subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	result, err := p.db.Exec("DELETE FROM "+table+" WHERE network = $1", ipNet.String())
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (p *PostgresDB) GetNetworks(table string) (*sql.Rows, error) {
	return p.db.Query("SELECT network FROM " + table)
}

func (p *PostgresDB) IsNetworkExists(table, subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	rows, err := p.db.Query("SELECT network FROM "+table+" WHERE network = $1", ipNet.String())
	if err != nil {
		return false, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	if rows.Next() {
		return true, nil
	}

	return false, nil
}
