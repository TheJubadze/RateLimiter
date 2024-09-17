package postgresipfilter

import (
	"fmt"
	"net"

	"github.com/TheJubadze/RateLimiter/infrastructure/storage/postgres"
	"github.com/TheJubadze/RateLimiter/interfaces/storage/iplists"
)

type PostgresqlService struct {
	db iplists.DB
}

// NewPostgresqlService initializes a new PostgresqlService
func NewPostgresqlService(connString string) (*PostgresqlService, error) {
	db, err := postgresdb.NewPostgresDB(connString)
	if err != nil {
		return nil, err
	}
	return &PostgresqlService{db: db}, nil
}

// Close closes the service connection
func (s *PostgresqlService) Close() error {
	return s.db.Close()
}

func (s *PostgresqlService) IsIPWhitelisted(ip string) bool {
	return s.isIPListed(ip, true)
}

func (s *PostgresqlService) IsIPBlacklisted(ip string) bool {
	return s.isIPListed(ip, false)
}

func (s *PostgresqlService) IsNetworkWhitelisted(network string) (bool, error) {
	return s.db.IsNetworkExists("whitelist", network)
}

func (s *PostgresqlService) IsNetworkBlacklisted(network string) (bool, error) {
	return s.db.IsNetworkExists("blacklist", network)
}

func (s *PostgresqlService) AddToWhitelist(subnet string) error {
	return s.db.InsertNetwork("whitelist", subnet)
}

func (s *PostgresqlService) RemoveFromWhitelist(subnet string) (bool, error) {
	return s.db.DeleteNetwork("whitelist", subnet)
}

func (s *PostgresqlService) AddToBlacklist(subnet string) error {
	return s.db.InsertNetwork("blacklist", subnet)
}

func (s *PostgresqlService) RemoveFromBlacklist(subnet string) (bool, error) {
	return s.db.DeleteNetwork("blacklist", subnet)
}

func (s *PostgresqlService) isIPListed(ip string, isWhitelist bool) bool {
	table := "blacklist"
	if isWhitelist {
		table = "whitelist"
	}

	rows, err := s.db.GetNetworks(table)
	if err != nil {
		return false
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var cidr string
		if err := rows.Scan(&cidr); err != nil {
			continue
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
