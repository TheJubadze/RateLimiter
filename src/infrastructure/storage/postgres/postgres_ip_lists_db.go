package postgresdb

import (
	"database/sql"
	"fmt"
	"net"
	"regexp"

	// postgres driver.
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

	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return err
	}

	// #nosec G201 - sanitized table name is safe
	query := fmt.Sprintf("INSERT INTO %s (network) VALUES ($1)", sanitizedTable)
	_, err = p.db.Exec(query, ipNet.String())

	return err
}

func (p *PostgresDB) DeleteNetwork(table, subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return false, err
	}

	// #nosec G201 - sanitized table name is safe
	query := fmt.Sprintf("DELETE FROM %s WHERE network = $1", sanitizedTable)
	result, err := p.db.Exec(query, ipNet.String())
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

	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return false, err
	}

	// #nosec G201 - sanitized table name is safe
	query := fmt.Sprintf("SELECT network FROM %s WHERE network = $1", sanitizedTable)
	rows, err := p.db.Query(query, ipNet.String())
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return false, err
	}

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

func sanitizeTableName(table string) (string, error) {
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", table); !matched {
		return "", fmt.Errorf("invalid table name: %s", table)
	}
	return table, nil
}
