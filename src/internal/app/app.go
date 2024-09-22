package app

import (
	"fmt"
	"os"

	"github.com/TheJubadze/RateLimiter/infrastructure/ipfilter"
	"github.com/TheJubadze/RateLimiter/infrastructure/logger"
	"github.com/TheJubadze/RateLimiter/infrastructure/storage/redis"
	"github.com/TheJubadze/RateLimiter/internal/api"
	"github.com/TheJubadze/RateLimiter/internal/config"
	"github.com/spf13/viper"
)

func StartServer(configFile *string) {
	// Load configuration
	cfg, err := initConfig(*configFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading config: %s\n", err)
		os.Exit(1)
	}

	logrusLogger := logruslogger.NewLogrusLogger(cfg.Logger.Level)

	// Initialize bucket storage (Redis-based)
	bucketStorage := redisstorage.NewRedisBucketStorage(logrusLogger, cfg.Redis.Addr)

	// Initialize whitelist/blacklist service
	ipFilterService, err := ipfilter.NewService(cfg.SQLStorage.DSN)
	if err != nil {
		logrusLogger.Fatalf("Failed to initialize IP filter service: %v", err)
		os.Exit(1)
	}

	// Start the server
	server := api.NewGrpcServer(cfg, logrusLogger, bucketStorage, ipFilterService)
	if err := server.Start(); err != nil {
		logrusLogger.Fatalf("Failed to start server: %v", err)
	}
}

func initConfig(configPath string) (*config.Config, error) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &config.Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
