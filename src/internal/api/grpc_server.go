package api

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/TheJubadze/RateLimiter/internal/config"
	"github.com/TheJubadze/RateLimiter/pkg/ipfilter"
	"github.com/TheJubadze/RateLimiter/pkg/logger"
	"github.com/TheJubadze/RateLimiter/pkg/storage"
	"github.com/TheJubadze/RateLimiter/proto/pb"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	pb.UnimplementedRateLimiterServer
	config          *config.Config
	logger          logger.Logger
	bucketStorage   storage.BucketStorage
	ipFilterService ipfilter.Service
}

func NewGrpcServer(cfg *config.Config, logger logger.Logger, bucketStorage storage.BucketStorage, ipFilterService ipfilter.Service) *GrpcServer {
	return &GrpcServer{
		config:          cfg,
		logger:          logger,
		bucketStorage:   bucketStorage,
		ipFilterService: ipFilterService,
	}
}

// Start starts the gRPC server.
func (s *GrpcServer) Start() error {
	lis, err := net.Listen("tcp", `:`+s.config.GrpcServer.Port)
	if err != nil {
		s.logger.Fatalf("Failed to listen: %v", err)
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRateLimiterServer(grpcServer, s)

	s.logger.Printf("Starting gRPC server on port %s", s.config.GrpcServer.Port)
	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Fatalf("Failed to serve: %v", err)
		return err
	}

	return nil
}

// Authorize implements the Authorize gRPC method.
func (s *GrpcServer) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	s.logger.Printf("Authorize request: %v", req)
	limitExceededResponse := &pb.AuthorizeResponse{
		Authorized: false,
	}

	if s.ipFilterService.IsIPWhitelisted(req.Ip) {
		return &pb.AuthorizeResponse{
			Authorized: true,
			Message:    "Authorized: IP is whitelisted",
		}, nil
	}

	if s.ipFilterService.IsIPBlacklisted(req.Ip) {
		return &pb.AuthorizeResponse{
			Authorized: false,
			Message:    "Unauthorized: IP is blacklisted",
		}, nil
	}

	leakRate := time.Duration(s.config.LoginLimits.LeakRate) * time.Second

	login := req.GetLogin()
	if login != "" {
		success, err := s.bucketStorage.CheckRateLimit(ctx, req.Login, s.config.LoginLimits.Login, leakRate)
		if err != nil {
			return nil, err
		}
		if !success {
			limitExceededResponse.Message = "Login rate limit exceeded"
			return limitExceededResponse, nil
		}
	}

	password := req.GetPassword()
	if password != "" {
		success, err := s.bucketStorage.CheckRateLimit(ctx, req.Password, s.config.LoginLimits.Password, leakRate)
		if err != nil {
			return nil, err
		}
		if !success {
			limitExceededResponse.Message = "Password rate limit exceeded"
			return limitExceededResponse, nil
		}
	}

	ip := req.GetIp()
	if ip != "" {
		success, err := s.bucketStorage.CheckRateLimit(ctx, req.Ip, s.config.LoginLimits.IP, leakRate)
		if err != nil {
			return nil, err
		}
		if !success {
			limitExceededResponse.Message = "IP rate limit exceeded"
			return limitExceededResponse, nil
		}
	}

	return &pb.AuthorizeResponse{
		Authorized: true,
		Message:    "Authorized",
	}, nil
}

// ResetBucket implements the ResetBucket gRPC method.
func (s *GrpcServer) ResetBucket(ctx context.Context, req *pb.ResetBucketRequest) (*pb.ResetBucketResponse, error) {
	if req == nil || (req.Ip == "" && req.Login == "") {
		return nil, fmt.Errorf("IP or login must be provided")
	}

	if req.Ip != "" {
		err := s.bucketStorage.ResetBucket(ctx, req.Ip)
		if err != nil {
			return nil, err
		}
		s.logger.Printf("Bucket reset for IP: %s", req.Ip)
	}

	if req.Login != "" {
		err := s.bucketStorage.ResetBucket(ctx, req.Login)
		if err != nil {
			return nil, err
		}
		s.logger.Printf("Bucket reset for login: %s", req.Login)
	}

	return &pb.ResetBucketResponse{
		Message: "Bucket reset",
	}, nil
}

// AddToWhitelist implements the AddToWhitelist gRPC method.
func (s *GrpcServer) AddToWhitelist(_ context.Context, req *pb.AddToWhitelistRequest) (*pb.AddToWhitelistResponse, error) {
	s.logger.Printf("Adding %s to the whitelist", req.Ip)

	isListed, err := isNetworkListed(s, req.Ip)
	if err != nil {
		return &pb.AddToWhitelistResponse{
			Message: err.Error(),
		}, err
	}
	if isListed == 1 {
		return &pb.AddToWhitelistResponse{
			Message: "IP is already whitelisted",
		}, nil
	}
	if isListed == 2 {
		return &pb.AddToWhitelistResponse{
			Message: "IP is already blacklisted",
		}, nil
	}

	err = s.ipFilterService.AddToWhitelist(req.Ip)
	if err != nil {
		return nil, err
	}

	return &pb.AddToWhitelistResponse{
		Message: fmt.Sprintf("Added %s to the whitelist", req.Ip),
	}, nil
}

// AddToBlacklist implements the AddToBlacklist gRPC method.
func (s *GrpcServer) AddToBlacklist(_ context.Context, req *pb.AddToBlacklistRequest) (*pb.AddToBlacklistResponse, error) {
	s.logger.Printf("Adding %s to the blacklist", req.Ip)

	isListed, err := isNetworkListed(s, req.Ip)
	if err != nil {
		return &pb.AddToBlacklistResponse{
			Message: err.Error(),
		}, err
	}
	if isListed == 1 {
		return &pb.AddToBlacklistResponse{
			Message: "IP is already whitelisted",
		}, nil
	}
	if isListed == 2 {
		return &pb.AddToBlacklistResponse{
			Message: "IP is already blacklisted",
		}, nil
	}

	err = s.ipFilterService.AddToBlacklist(req.Ip)
	if err != nil {
		return nil, err
	}

	return &pb.AddToBlacklistResponse{
		Message: fmt.Sprintf("Added %s to the blacklist", req.Ip),
	}, nil
}

// RemoveFromWhitelist implements the RemoveFromWhitelist gRPC method.
func (s *GrpcServer) RemoveFromWhitelist(_ context.Context, req *pb.RemoveFromWhitelistRequest) (*pb.RemoveFromWhitelistResponse, error) {
	s.logger.Printf("Removing %s from the whitelist", req.Ip)

	result, err := s.ipFilterService.RemoveFromWhitelist(req.Ip)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Removed %s from the whitelist", req.Ip)
	if !result {
		message = fmt.Sprintf("%s not found in the whitelist", req.Ip)
	}

	return &pb.RemoveFromWhitelistResponse{
		Message: message,
	}, nil
}

// RemoveFromBlacklist implements the RemoveFromBlacklist gRPC method.
func (s *GrpcServer) RemoveFromBlacklist(_ context.Context, req *pb.RemoveFromBlacklistRequest) (*pb.RemoveFromBlacklistResponse, error) {
	s.logger.Printf("Removing %s from the blacklist", req.Ip)

	result, err := s.ipFilterService.RemoveFromBlacklist(req.Ip)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Removed %s from the blacklist", req.Ip)
	if !result {
		message = fmt.Sprintf("%s not found in the blacklist", req.Ip)
	}

	return &pb.RemoveFromBlacklistResponse{
		Message: message,
	}, nil
}

// isNetworkListed checks if the IP is already listed in the whitelist or blacklist
// Returns:
// -1 - error occurred
// 0 - IP is not listed
// 1 - IP is whitelisted
// 2 - IP is blacklisted.
func isNetworkListed(s *GrpcServer, ip string) (int, error) {
	isInList, err := s.ipFilterService.IsNetworkWhitelisted(ip)
	if err != nil {
		return -1, err
	}
	if isInList {
		return 1, nil
	}

	isInList, err = s.ipFilterService.IsNetworkBlacklisted(ip)
	if err != nil {
		return -1, err
	}
	if isInList {
		return 2, nil
	}

	return 0, nil
}
