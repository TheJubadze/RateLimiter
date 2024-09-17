package ipfilter

import (
	"github.com/stretchr/testify/mock"
)

type MockIPFilterService struct {
	mock.Mock
}

func (m *MockIPFilterService) IsIPWhitelisted(ip string) bool {
	args := m.Called(ip)
	return args.Bool(0)
}

func (m *MockIPFilterService) IsIPBlacklisted(ip string) bool {
	args := m.Called(ip)
	return args.Bool(0)
}

func (m *MockIPFilterService) IsNetworkWhitelisted(network string) (bool, error) {
	args := m.Called(network)
	return args.Bool(0), args.Error(1)
}

func (m *MockIPFilterService) IsNetworkBlacklisted(network string) (bool, error) {
	args := m.Called(network)
	return args.Bool(0), args.Error(1)
}

func (m *MockIPFilterService) AddToWhitelist(subnet string) error {
	args := m.Called(subnet)
	return args.Error(0)
}

func (m *MockIPFilterService) RemoveFromWhitelist(subnet string) (bool, error) {
	args := m.Called(subnet)
	return args.Bool(0), args.Error(1)
}

func (m *MockIPFilterService) AddToBlacklist(subnet string) error {
	args := m.Called(subnet)
	return args.Error(0)
}

func (m *MockIPFilterService) RemoveFromBlacklist(subnet string) (bool, error) {
	args := m.Called(subnet)
	return args.Bool(0), args.Error(1)
}
