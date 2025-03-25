package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment      string `env:"APP_ENV" env-default:"local"`
	Logger           LoggerConfig
	GRPC             GRPCConfig
	PG               PostgresConfig
	RateLimiter      RateLimiterConfig
	CircuitBreaker   CircuitBreakerConfig
	Kafka            KafkaConfig
	ProfilingEnabled bool `env:"PROFILING_ENABLED" env-default:"false"`
}

type LoggerConfig struct {
	Level               string   `env:"LOG_LEVEL" env-default:"warn"`
	Encoding            string   `env:"LOG_ENCODING" env-default:"console"`
	OutputPaths         []string `env:"LOG_OUTPUT" env-default:"stdout"`
	ErrorOutputPaths    []string `env:"LOG_ERROR_OUTPUT" env-default:"stderr"`
	OutputFilePath      string   `env:"LOG_OUTPUT_FILE" env-default:"logs/inventory.log"`
	ErrorOutputFilePath string   `env:"LOG_ERROR_OUTPUT_FILE" env-default:"logs/inventory_error.log"`
}

type GRPCConfig struct {
	Host string `env:"GRPC_HOST" env-default:"localhost"`
	Port string `env:"GRPC_PORT"`
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

type KafkaConfig struct {
	// List of brokers to connect to.
	Brokers []string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	// The group id to use when consuming messages.
	GroupID string `env:"KAFKA_GROUP_ID" env-default:"inventory-service"`
	// Topics to consume messages from.
	Topics []string `env:"KAFKA_TOPICS" env-default:"inventory-events"`
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
