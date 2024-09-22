package postgresdb

import (
	"database/sql"
	"fmt"
	"regexp"

	// postgres driver.
	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (d *Database) Insert(table string, network string) error {
	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (network) VALUES ($1)", sanitizedTable)
	_, err = d.DB.Exec(query, network)
	if err != nil {
		return fmt.Errorf("failed to insert network: %v", err)
	}

	return nil
}

func (d *Database) Delete(table string, network string) (bool, error) {
	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE network = $1", sanitizedTable)
	result, err := d.DB.Exec(query, network)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (d *Database) GetAll(table string) ([]string, error) {
	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return nil, err
	}

	query := "SELECT network FROM " + sanitizedTable
	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to select networks: %v", err)
	}
	defer rows.Close()

	var networks []string
	for rows.Next() {
		var network string
		if err := rows.Scan(&network); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		networks = append(networks, network)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return networks, nil
}

func (d *Database) GetByValue(table string, network string) (bool, error) {
	sanitizedTable, err := sanitizeTableName(table)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("SELECT network FROM %s WHERE network = $1", sanitizedTable)
	rows, err := d.DB.Query(query, network)
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

func (d *Database) Close() error {
	return d.DB.Close()
}

func sanitizeTableName(table string) (string, error) {
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", table); !matched {
		return "", fmt.Errorf("invalid table name: %s", table)
	}
	return table, nil
}
