package config

type loggerConfig struct {
	Level string `mapstructure:"level"`
}

type grpcServerConfig struct {
	Port string `mapstructure:"port"`
}

type sqlStorageConfig struct {
	DSN           string `mapstructure:"dsn"`
	MigrationsDir string `mapstructure:"migrations_dir"`
}

type redisConfig struct {
	Addr string `mapstructure:"addr"`
}

type loginLimitsConfig struct {
	LeakRate int `mapstructure:"leak_rate"`
	Login    int `mapstructure:"login_capacity"`
	Password int `mapstructure:"password_capacity"`
	IP       int `mapstructure:"ip_capacity"`
}

type Config struct {
	Logger      loggerConfig      `mapstructure:"logger"`
	GrpcServer  grpcServerConfig  `mapstructure:"grpc_server"`
	SQLStorage  sqlStorageConfig  `mapstructure:"sql_storage"`
	Redis       redisConfig       `mapstructure:"redis"`
	LoginLimits loginLimitsConfig `mapstructure:"leaky_bucket"`
}

func CreateTestConfig(leakRate int, loginCapacity int, passwordCapacity int, ipCapacity int) *Config {
	return &Config{
		Logger: loggerConfig{
			Level: "info",
		},
		GrpcServer: grpcServerConfig{
			Port: ":50051",
		},
		SQLStorage: sqlStorageConfig{
			DSN:           "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable",
			MigrationsDir: "file://migrations",
		},
		Redis: redisConfig{
			Addr: "localhost:6379",
		},
		LoginLimits: loginLimitsConfig{
			LeakRate: leakRate,
			Login:    loginCapacity,
			Password: passwordCapacity,
			IP:       ipCapacity,
		},
	}
}
