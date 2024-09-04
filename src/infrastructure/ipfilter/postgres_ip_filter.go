package postgresipfilter

import (
	"database/sql"
	"fmt"
	"net"

	// Import PostgreSQL driver for database/sql.
	_ "github.com/lib/pq"
)

type PostgresqlService struct {
	db *sql.DB
}

func NewPostgresqlService(connString string) *PostgresqlService {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	return &PostgresqlService{db: db}
}

func (s *PostgresqlService) IsIPWhitelisted(ip string) bool {
	return s.isIPListed(ip, true)
}

func (s *PostgresqlService) IsIPBlacklisted(ip string) bool {
	return s.isIPListed(ip, false)
}

func (s *PostgresqlService) IsNetworkWhitelisted(network string) (bool, error) {
	return s.isNetworkListed(network, true)
}

func (s *PostgresqlService) IsNetworkBlacklisted(network string) (bool, error) {
	return s.isNetworkListed(network, false)
}

func (s *PostgresqlService) AddToWhitelist(subnet string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO whitelist (network) VALUES ($1)", ipNet.String())
	return err
}

func (s *PostgresqlService) RemoveFromWhitelist(subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	result, err := s.db.Exec("DELETE FROM whitelist WHERE network = $1", ipNet.String())
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (s *PostgresqlService) AddToBlacklist(subnet string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO blacklist (network) VALUES ($1)", ipNet.String())
	return err
}

func (s *PostgresqlService) RemoveFromBlacklist(subnet string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	result, err := s.db.Exec("DELETE FROM blacklist WHERE network = $1", ipNet.String())
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (s *PostgresqlService) isIPListed(ip string, isWhitelist bool) bool {
	query := "SELECT network FROM blacklist"
	if isWhitelist {
		query = "SELECT network FROM whitelist"
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return false
	}
	defer func() {
		_ = rows.Close()
	}()

	if rows == nil || rows.Err() != nil {
		return false
	}

	for rows.Next() {
		var cidr string
		if err := rows.Scan(&cidr); err != nil {
			return false
		}

		inSubnet, err := isIPInSubnet(ip, cidr)
		if err != nil {
			continue
		}
		if inSubnet {
			return true
		}
	}

	return false
}

func (s *PostgresqlService) isNetworkListed(subnet string, isWhitelist bool) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	query := "SELECT network FROM blacklist WHERE network = $1"
	if isWhitelist {
		query = "SELECT network FROM whitelist WHERE network = $1"
	}

	rows, err := s.db.Query(query, ipNet.String())
	if err != nil {
		return false, err
	}
	defer func() {
		_ = rows.Close()
	}()

	if rows == nil {
		return false, nil
	}
	if rows.Err() != nil {
		return false, rows.Err()
	}

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

func isIPInSubnet(ipStr string, cidr string) (bool, error) {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR: %w", err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %v", ipStr)
	}

	return subnet.Contains(ip), nil
}
