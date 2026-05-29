package config

import (
	"os"
	"path/filepath"
	"strings"
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

type ProfileServiceConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type ObservabilityConfig struct {
	MetricsAddr string `yaml:"metrics_addr"`
}

type KafkaConfig struct {
	Brokers       string `yaml:"brokers"`
	AdEventsTopic string `yaml:"ad_events_topic"`
	ConsumerGroup string `yaml:"consumer_group"`
}

func (c KafkaConfig) BrokerList() []string {
	parts := strings.Split(c.Brokers, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			out = append(out, broker)
		}
	}
	return out
}

type ClickHouseConfig struct {
	Addr     string `yaml:"addr"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type YookassaConfig struct {
	ShopID     string `yaml:"shop_id"`
	SecretKey  string `yaml:"secret_key"`
	ReturnURL  string `yaml:"return_url"`
	WebhookURL string `yaml:"webhook_url"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type BalanceAutomationConfig struct {
	AutopayInterval            time.Duration `yaml:"autopay_interval"`
	NotificationInterval       time.Duration `yaml:"notification_interval"`
	SubscriptionExpiryInterval time.Duration `yaml:"subscription_expiry_interval"`
}

type RegistrationVerificationConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

type PasswordResetConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

type OpenAIConfig struct {
	APIKey       string        `yaml:"api_key"`
	BaseURL      string        `yaml:"base_url"`
	TextModel    string        `yaml:"text_model"`
	Timeout      time.Duration `yaml:"timeout"`
	Organization string        `yaml:"organization"`
	Project      string        `yaml:"project"`
}

type NanoBananaConfig struct {
	APIKey       string        `yaml:"api_key"`
	BaseURL      string        `yaml:"base_url"`
	CallbackURL  string        `yaml:"callback_url"`
	Timeout      time.Duration `yaml:"timeout"`
	PollInterval time.Duration `yaml:"poll_interval"`
	PollTimeout  time.Duration `yaml:"poll_timeout"`
}

type Config struct {
	HTTPServer               HTTPServerConfig               `yaml:"http_server"`
	Postgres                 PostgresConfig                 `yaml:"postgres"`
	Redis                    RedisConfig                    `yaml:"redis"`
	Session                  SessionConfig                  `yaml:"session"`
	AuthService              AuthServiceConfig              `yaml:"auth_service"`
	ProfileService           ProfileServiceConfig           `yaml:"profile_service"`
	S3                       S3Config                       `yaml:"s3"`
	CORS                     CORSConfig                     `yaml:"cors"`
	Observability            ObservabilityConfig            `yaml:"observability"`
	Kafka                    KafkaConfig                    `yaml:"kafka"`
	ClickHouse               ClickHouseConfig               `yaml:"clickhouse"`
	Yookassa                 YookassaConfig                 `yaml:"yookassa"`
	SMTP                     SMTPConfig                     `yaml:"smtp"`
	BalanceAutomation        BalanceAutomationConfig        `yaml:"balance_automation"`
	RegistrationVerification RegistrationVerificationConfig `yaml:"registration_verification"`
	PasswordReset            PasswordResetConfig            `yaml:"password_reset"`
	OpenAI                   OpenAIConfig                   `yaml:"openai"`
	NanoBanana               NanoBananaConfig               `yaml:"nanobanana"`
	GracefulTimeout          time.Duration                  `yaml:"graceful_timeout"`
}

func ReadConfig(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	expanded := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
