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

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type RedisConfig struct {
	Host             string        `yaml:"host"`
	Port             int           `yaml:"port"`
	Password         string        `yaml:"password"`
	DB               int           `yaml:"db"`
	MaxIdle          int           `yaml:"max_idle"`
	MaxActive        int           `yaml:"max_active"`
	IdleTimeout      time.Duration `yaml:"idle_timeout"`
	Wait             bool          `yaml:"wait"`
	ConnectTimeout   time.Duration `yaml:"connect_timeout"`
	ReadTimeout      time.Duration `yaml:"read_timeout"`
	WriteTimeout     time.Duration `yaml:"write_timeout"`
	PingAfterIdleFor time.Duration `yaml:"ping_after_idle_for"`
}

type SessionConfig struct {
	TTL          time.Duration `yaml:"ttl"`
	CookieName   string        `yaml:"cookie_name"`
	CookiePath   string        `yaml:"cookie_path"`
	CookieSecure bool          `yaml:"cookie_secure"`
}

type S3Config struct {
	Endpoint       string `yaml:"endpoint"`
	Region         string `yaml:"region"`
	Bucket         string `yaml:"bucket"`
	AccessKey      string `yaml:"access_key"`
	SecretKey      string `yaml:"secret_key"`
	PublicBaseURL  string `yaml:"public_base_url"`
	ForcePathStyle bool   `yaml:"force_path_style"`
}

type AuthServiceConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type VKIDConfig struct {
	ClientID           int64  `yaml:"client_id"`
	RedirectURI        string `yaml:"redirect_uri"`
	AuthDomain         string `yaml:"auth_domain"`
	Scope              string `yaml:"scope"`
	DefaultRedirectURL string `yaml:"default_redirect_url"`
	ErrorRedirectURL   string `yaml:"error_redirect_url"`
}

type ProfileServiceConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type Config struct {
	HTTPServer      HTTPServerConfig     `yaml:"http_server"`
	Postgres        PostgresConfig       `yaml:"postgres"`
	Redis           RedisConfig          `yaml:"redis"`
	Session         SessionConfig        `yaml:"session"`
	AuthService     AuthServiceConfig    `yaml:"auth_service"`
	ProfileService  ProfileServiceConfig `yaml:"profile_service"`
	VKID            VKIDConfig           `yaml:"vkid"`
	S3              S3Config             `yaml:"s3"`
	CORS            CORSConfig           `yaml:"cors"`
	GracefulTimeout time.Duration        `yaml:"graceful_timeout"`
}

func ReadConfig(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expanded := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
