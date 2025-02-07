package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	GRPC           GRPCConfig
	PG             PostgresConfig
	Prometheus     PrometheusConfig
	Pprof          PprofConfig
	RateLimiter    RateLimiterConfig
	CircuitBreaker CircuitBreakerConfig
}

type GRPCConfig struct {
	Host string `env:"GRPC_HOST" env_default:"localhost"`
	Port string `env:"GRPC_PORT" required:"true"`
}

func (g *GRPCConfig) Addr() string {
	return g.Host + ":" + g.Port
}

type PostgresConfig struct {
	Host     string `env:"PG_HOST"`
	Port     string `env:"PG_PORT"`
	User     string `env:"PG_USER"`
	Password string `env:"PG_PASSWORD"`
	DBName   string `env:"PG_DBNAME"`
	SSLMode  string `env:"PG_SSLMODE"`
}

func (p *PostgresConfig) DSN() string {
	return "host=" + p.Host + " port=" + p.Port + " user=" + p.User + " password=" + p.Password + " dbname=" + p.DBName + " sslmode=" + p.SSLMode
}

type PrometheusConfig struct {
	Port string `env:"PROMETHEUS_PORT" env_default:"9090"`
}

type PprofConfig struct {
	Port string `env:"PPROF_PORT" env_default:"6060"`
}

type RateLimiterConfig struct {
	Limit int `env:"RATE_LIMITER_LIMIT" env_default:"150"`
	Burst int `env:"RATE_LIMITER_BURST" env_default:"150"`
}

type CircuitBreakerConfig struct {
	MaxRequests uint32        `env:"CIRCUIT_BREAKER_MAX_REQUESTS" env_default:"5"`
	Interval    time.Duration `env:"CIRCUIT_BREAKER_INTERVAL" env_default:"60s"`
	Timeout     time.Duration `env:"CIRCUIT_BREAKER_TIMEOUT" env_default:"5s"`
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
