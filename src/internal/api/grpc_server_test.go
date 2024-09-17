package api_test

import (
	"context"
	"testing"

	"github.com/TheJubadze/RateLimiter/infrastructure/logger"
	"github.com/TheJubadze/RateLimiter/interfaces/ipfilter"
	"github.com/TheJubadze/RateLimiter/interfaces/storage/bucket"
	"github.com/TheJubadze/RateLimiter/internal/api"
	"github.com/TheJubadze/RateLimiter/internal/config"
	"github.com/TheJubadze/RateLimiter/proto/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthorize(t *testing.T) {
	mockIPFilterService := new(ipfilter.MockIPFilterService)
	mockBucketStorage := new(bucket.MockBucketStorage)
	resetMocks := func() {
		mockIPFilterService.ExpectedCalls = nil
		mockBucketStorage.ExpectedCalls = nil
	}

	cfg := config.CreateTestConfig(1, 5, 5, 5)
	log := logruslogger.NewLogrusLogger("info")

	server := api.NewGrpcServer(cfg, log, mockBucketStorage, mockIPFilterService)

	tests := []struct {
		name       string
		req        *pb.AuthorizeRequest
		setupMocks func()
		expected   *pb.AuthorizeResponse
		expectErr  bool
	}{
		{
			name: "IP Whitelisted",
			req:  &pb.AuthorizeRequest{Ip: "192.168.1.1"},
			setupMocks: func() {
				resetMocks()
				mockIPFilterService.On("IsIPWhitelisted", "192.168.1.1").Return(true)
			},
			expected: &pb.AuthorizeResponse{
				Authorized: true,
				Message:    "Authorized: IP is whitelisted",
			},
			expectErr: false,
		},
		{
			name: "IP Blacklisted",
			req:  &pb.AuthorizeRequest{Ip: "192.168.1.1"},
			setupMocks: func() {
				resetMocks()
				mockIPFilterService.On("IsIPWhitelisted", "192.168.1.1").Return(false)
				mockIPFilterService.On("IsIPBlacklisted", "192.168.1.1").Return(true)
			},
			expected: &pb.AuthorizeResponse{
				Authorized: false,
				Message:    "Unauthorized: IP is blacklisted",
			},
			expectErr: false,
		},
		{
			name: "Rate Limit Exceeded",
			req:  &pb.AuthorizeRequest{Ip: "192.168.1.1", Login: "user"},
			setupMocks: func() {
				resetMocks()
				mockIPFilterService.On("IsIPWhitelisted", "192.168.1.1").Return(false)
				mockIPFilterService.On("IsIPBlacklisted", "192.168.1.1").Return(false)
				mockBucketStorage.On("CheckRateLimit", mock.Anything, "user", 5, mock.Anything).Return(false, nil)
			},
			expected: &pb.AuthorizeResponse{
				Authorized: false,
				Message:    "Login rate limit exceeded",
			},
			expectErr: false,
		},
		{
			name: "Authorized",
			req:  &pb.AuthorizeRequest{Ip: "192.168.1.1", Login: "user"},
			setupMocks: func() {
				resetMocks()
				mockIPFilterService.On("IsIPWhitelisted", "192.168.1.1").Return(false)
				mockIPFilterService.On("IsIPBlacklisted", "192.168.1.1").Return(false)
				mockBucketStorage.On("CheckRateLimit", mock.Anything, "user", 5, mock.Anything).Return(true, nil)
				mockBucketStorage.On("CheckRateLimit", mock.Anything, "192.168.1.1", 5, mock.Anything).Return(true, nil)
			},
			expected: &pb.AuthorizeResponse{
				Authorized: true,
				Message:    "Authorized",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.Authorize(context.Background(), tt.req)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockIPFilterService.AssertExpectations(t)
			mockBucketStorage.AssertExpectations(t)
		})
	}
}

func TestResetBucket(t *testing.T) {
	mockBucketStorage := new(bucket.MockBucketStorage)
	cfg := &config.Config{}
	log := logruslogger.NewLogrusLogger("info")
	mockIPFilterService := new(ipfilter.MockIPFilterService)

	server := api.NewGrpcServer(cfg, log, mockBucketStorage, mockIPFilterService)

	tests := []struct {
		name       string
		req        *pb.ResetBucketRequest
		setupMocks func()
		expected   *pb.ResetBucketResponse
		expectErr  bool
	}{
		{
			name: "Reset IP Bucket",
			req:  &pb.ResetBucketRequest{Ip: "192.168.1.1"},
			setupMocks: func() {
				mockBucketStorage.On("ResetBucket", mock.Anything, "192.168.1.1").Return(nil)
			},
			expected: &pb.ResetBucketResponse{
				Message: "Bucket reset",
			},
			expectErr: false,
		},
		{
			name: "Reset Login Bucket",
			req:  &pb.ResetBucketRequest{Login: "user"},
			setupMocks: func() {
				mockBucketStorage.On("ResetBucket", mock.Anything, "user").Return(nil)
			},
			expected: &pb.ResetBucketResponse{
				Message: "Bucket reset",
			},
			expectErr: false,
		},
		{
			name:       "No IP or Login Provided",
			req:        &pb.ResetBucketRequest{},
			setupMocks: func() {},
			expected:   nil,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.ResetBucket(context.Background(), tt.req)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockBucketStorage.AssertExpectations(t)
		})
	}
}

func TestAddToWhitelist(t *testing.T) {
	mockIPFilterService := new(ipfilter.MockIPFilterService)
	mockIPFilterService.On("IsNetworkWhitelisted", "192.168.1.1/24").Return(false, nil)
	mockIPFilterService.On("IsNetworkBlacklisted", "192.168.1.1/24").Return(false, nil)
	mockIPFilterService.On("AddToWhitelist", "192.168.1.1/24").Return(nil)

	cfg := &config.Config{}
	log := logruslogger.NewLogrusLogger("info")
	bucketStorage := new(bucket.MockBucketStorage)

	server := api.NewGrpcServer(cfg, log, bucketStorage, mockIPFilterService)

	req := &pb.AddToWhitelistRequest{Ip: "192.168.1.1/24"}
	resp, err := server.AddToWhitelist(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Added 192.168.1.1/24 to the whitelist", resp.Message)
	mockIPFilterService.AssertExpectations(t)
}

func TestAddToBlacklist(t *testing.T) {
	mockIPFilterService := new(ipfilter.MockIPFilterService)
	mockIPFilterService.On("IsNetworkWhitelisted", "192.168.1.1/24").Return(false, nil)
	mockIPFilterService.On("IsNetworkBlacklisted", "192.168.1.1/24").Return(false, nil)
	mockIPFilterService.On("AddToBlacklist", "192.168.1.1/24").Return(nil)

	cfg := &config.Config{}
	log := logruslogger.NewLogrusLogger("info")
	bucketStorage := new(bucket.MockBucketStorage)

	server := api.NewGrpcServer(cfg, log, bucketStorage, mockIPFilterService)

	req := &pb.AddToBlacklistRequest{Ip: "192.168.1.1/24"}
	resp, err := server.AddToBlacklist(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Added 192.168.1.1/24 to the blacklist", resp.Message)
	mockIPFilterService.AssertExpectations(t)
}

func TestRemoveFromWhitelist(t *testing.T) {
	mockIPFilterService := new(ipfilter.MockIPFilterService)
	mockIPFilterService.On("RemoveFromWhitelist", "192.168.1.1/24").Return(true, nil)

	cfg := &config.Config{}
	log := logruslogger.NewLogrusLogger("info")
	bucketStorage := new(bucket.MockBucketStorage)

	server := api.NewGrpcServer(cfg, log, bucketStorage, mockIPFilterService)

	req := &pb.RemoveFromWhitelistRequest{Ip: "192.168.1.1/24"}
	resp, err := server.RemoveFromWhitelist(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Removed 192.168.1.1/24 from the whitelist", resp.Message)
	mockIPFilterService.AssertExpectations(t)
}

func TestRemoveFromBlacklist(t *testing.T) {
	mockIPFilterService := new(ipfilter.MockIPFilterService)
	mockIPFilterService.On("RemoveFromBlacklist", "192.168.1.1/24").Return(true, nil)

	cfg := &config.Config{}
	log := logruslogger.NewLogrusLogger("info")
	bucketStorage := new(bucket.MockBucketStorage)

	server := api.NewGrpcServer(cfg, log, bucketStorage, mockIPFilterService)

	req := &pb.RemoveFromBlacklistRequest{Ip: "192.168.1.1/24"}
	resp, err := server.RemoveFromBlacklist(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Removed 192.168.1.1/24 from the blacklist", resp.Message)
	mockIPFilterService.AssertExpectations(t)
}
