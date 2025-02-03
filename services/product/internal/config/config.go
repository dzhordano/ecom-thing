package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	GRPC GRPCConfig
	PG   PostgresConfig
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

// MustNew Reads .env file and returns Config.
func MustNew() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	return &cfg
}
