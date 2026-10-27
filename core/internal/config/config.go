// internal/config/config.go
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	HTTP   HTTPConfig
	GRPC   GRPCConfig
	DB     DBConfig
	NATS   NATSConfig
}

type HTTPConfig struct {
	Port         string        `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
}

type GRPCConfig struct {
	RustParserAddr string `env:"RUST_PARSER_ADDR" envDefault:"localhost:50051"`
	PythonAIAddr   string `env:"PYTHON_AI_ADDR" envDefault:"localhost:50052"`
}

type DBConfig struct {
	PostgresURL string `env:"POSTGRES_URL" envDefault:"postgres://chronos:chronos_password@localhost:5432/chronos_db"`
	Neo4jURL    string `env:"NEO4J_URL" envDefault:"bolt://localhost:7687"`
}

type NATSConfig struct {
	URL string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
}

// Load đọc config và validate
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	// Thêm logic validate nếu cần (vd: check xem URL có bị rỗng không)
	return cfg, nil
}
