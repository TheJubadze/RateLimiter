package ipfilter

import (
	"fmt"
	"net"

	"github.com/TheJubadze/RateLimiter/infrastructure/storage/iplists"
	"github.com/TheJubadze/RateLimiter/interfaces/storage/iplists"
)

type Service struct {
	repository iplists.Repository
}

func NewService(connString string) (*Service, error) {
	repo, err := iplistsrepository.NewRepository(connString)
	if err != nil {
		return nil, err
	}
	return &Service{repository: repo}, nil
}

func (s *Service) Close() error {
	return s.repository.Close()
}

func (s *Service) IsIPWhitelisted(ip string) bool {
	return s.isIPListed(ip, true)
}

func (s *Service) IsIPBlacklisted(ip string) bool {
	return s.isIPListed(ip, false)
}

func (s *Service) IsNetworkWhitelisted(network string) (bool, error) {
	return s.repository.IsNetworkExists("whitelist", network)
}

func (s *Service) IsNetworkBlacklisted(network string) (bool, error) {
	return s.repository.IsNetworkExists("blacklist", network)
}

func (s *Service) AddToWhitelist(subnet string) error {
	return s.repository.InsertNetwork("whitelist", subnet)
}

func (s *Service) RemoveFromWhitelist(subnet string) (bool, error) {
	return s.repository.DeleteNetwork("whitelist", subnet)
}

func (s *Service) AddToBlacklist(subnet string) error {
	return s.repository.InsertNetwork("blacklist", subnet)
}

func (s *Service) RemoveFromBlacklist(subnet string) (bool, error) {
	return s.repository.DeleteNetwork("blacklist", subnet)
}

func (s *Service) isIPListed(ip string, isWhitelist bool) bool {
	table := "blacklist"
	if isWhitelist {
		table = "whitelist"
	}

	rows, err := s.repository.GetNetworks(table)
	if err != nil {
		return false
	}

	for _, cidr := range rows {
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
