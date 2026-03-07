package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type HTTPServerConfig struct {
	Listen       string        `yaml:"listen"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	HTTPServer      HTTPServerConfig `yaml:"http_server"`
	Postgres        PostgresConfig   `yaml:"postgres"`
	GracefulTimeout time.Duration    `yaml:"graceful_timeout"`
}

func ReadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
