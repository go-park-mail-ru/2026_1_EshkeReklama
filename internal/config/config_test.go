package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfig_ExpandsEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	_ = os.Setenv("LISTEN_ADDR", ":9999")

	data := []byte(`
http_server:
  listen: ${LISTEN_ADDR}
  read_timeout: 1s
  write_timeout: 2s
session:
  ttl: 10m
  cookie_name: session_id
  cookie_path: /
  cookie_secure: false
cors:
  allowed_origins: ["http://localhost:3000"]
graceful_timeout: 3s
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if cfg.HTTPServer.Listen != ":9999" {
		t.Fatalf("expected listen :9999 got %q", cfg.HTTPServer.Listen)
	}
}

func TestKafkaConfigBrokerList(t *testing.T) {
	cfg := KafkaConfig{Brokers: " kafka:9092, localhost:9094 , ,127.0.0.1:9092 "}
	got := cfg.BrokerList()
	if len(got) != 3 || got[0] != "kafka:9092" || got[1] != "localhost:9094" || got[2] != "127.0.0.1:9092" {
		t.Fatalf("unexpected broker list: %#v", got)
	}
}
