package config

import (
	"log"
	"net"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment      string `env:"APP_ENV" env-default:"local"`
	Logger           LoggerConfig
	GRPC             GRPCServiceConfig   `env-prefix:"GRPC_"`
	GRPCProduct      GRPCProductConfig   `env-prefix:"GRPC_PRODUCT_"`
	GRPCInventory    GRPCInventoryConfig `env-prefix:"GRPC_INVENTORY_"`
	PG               PostgresConfig
	RateLimiter      RateLimiterConfig
	CircuitBreaker   CircuitBreakerConfig
	ProfilingEnabled bool `env:"PROFILING_ENABLED" env-default:"false"`
}

type LoggerConfig struct {
	Level            string   `env:"LOG_LEVEL" env-default:"warn"`
	OutputPaths      []string `env:"LOG_OUTPUT" env-default:"stdout"`
	ErrorOutputPaths []string `env:"LOG_ERROR_OUTPUT" env-default:"stderr"`
	Encoding         string   `env:"LOG_ENCODING" env-default:"console"`
}

// Базовая конфигурация для gRPC-сервисов.
type GRPCServiceConfig struct {
	Host string `env:"HOST" env-default:"localhost"`
	Port string `env:"PORT"`
}

// Addr возвращает корректно сформированный адрес.
func (g GRPCServiceConfig) Addr() string {
	return net.JoinHostPort(g.Host, g.Port)
}

// Если нужно, можно определить отдельные типы для разных сервисов,
// чтобы задать им разные env-теги.
type GRPCProductConfig struct {
	GRPCServiceConfig
}

type GRPCInventoryConfig struct {
	GRPCServiceConfig
}

// TODO maybe порт тестовой бд тоже нужен
type PostgresConfig struct {
	Host     string `env:"PG_HOST"`
	Port     string `env:"PG_PORT"`
	User     string `env:"PG_USER"`
	Password string `env:"PG_PASSWORD"`
	DBName   string `env:"PG_DBNAME"`
	SSLMode  string `env:"PG_SSLMODE" env-default:"disable"`
}

func (p *PostgresConfig) DSN() string {
	return "host=" + p.Host + " port=" + p.Port + " user=" + p.User + " password=" + p.Password + " dbname=" + p.DBName + " sslmode=" + p.SSLMode
}

func (p *PostgresConfig) URL() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DBName + "?sslmode=" + p.SSLMode
}

type RateLimiterConfig struct {
	Limit int `env:"RATE_LIMITER_LIMIT" env-default:"150"`
	Burst int `env:"RATE_LIMITER_BURST" env-default:"150"`
}

type CircuitBreakerConfig struct {
	MaxRequests uint32        `env:"CIRCUIT_BREAKER_MAX_REQUESTS" env-default:"5"`
	Interval    time.Duration `env:"CIRCUIT_BREAKER_INTERVAL" env-default:"60s"`
	Timeout     time.Duration `env:"CIRCUIT_BREAKER_TIMEOUT" env-default:"5s"`
}

// MustNew Reads .env file and returns Config.
func MustNew() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("config: error loading .env file: %v", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	return &cfg
}
