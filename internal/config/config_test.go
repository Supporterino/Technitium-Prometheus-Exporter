package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadValidConfig(t *testing.T) {
	configYAML := `
exporter:
  listen_address: ":9119"
  metrics_path: "/metrics"
  scrape_timeout: 30s
  log_level: "info"

targets:
  - name: "dns-primary"
    url: "http://localhost:5380"
    api_token: "test-token"
    request_timeout: 10s
    labels:
      env: "prod"
    features:
      dhcp: true
      cluster: false
`
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Exporter.ListenAddress != ":9119" {
		t.Errorf("expected listen_address :9119, got %q", cfg.Exporter.ListenAddress)
	}
	if cfg.Exporter.ScrapeTimeout != 30*time.Second {
		t.Errorf("expected scrape_timeout 30s, got %v", cfg.Exporter.ScrapeTimeout)
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "dns-primary" {
		t.Errorf("expected target name dns-primary, got %q", cfg.Targets[0].Name)
	}
	if cfg.Targets[0].RequestTimeout != 10*time.Second {
		t.Errorf("expected request_timeout 10s, got %v", cfg.Targets[0].RequestTimeout)
	}
	if cfg.Targets[0].Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got %q", cfg.Targets[0].Labels["env"])
	}
	if !cfg.Targets[0].Features.DHCP {
		t.Error("expected DHCP feature enabled")
	}
	if cfg.Targets[0].Features.Cluster {
		t.Error("expected Cluster feature disabled")
	}
}

func TestValidateDefaults(t *testing.T) {
	configYAML := `
targets:
  - name: "dns"
    url: "http://localhost:5380"
    api_token: "token"
`
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Exporter.ListenAddress != ":9119" {
		t.Errorf("expected default listen_address :9119, got %q", cfg.Exporter.ListenAddress)
	}
	if cfg.Exporter.MetricsPath != "/metrics" {
		t.Errorf("expected default metrics_path /metrics, got %q", cfg.Exporter.MetricsPath)
	}
	if cfg.Exporter.LogLevel != "info" {
		t.Errorf("expected default log_level info, got %q", cfg.Exporter.LogLevel)
	}
	if cfg.Targets[0].RequestTimeout != 5*time.Second {
		t.Errorf("expected default request_timeout 5s, got %v", cfg.Targets[0].RequestTimeout)
	}
}

func TestValidateMissingTargetFields(t *testing.T) {
	tests := []struct {
		name   string
		config string
		errMsg string
	}{
		{
			name: "missing name",
			config: `
targets:
  - url: "http://localhost:5380"
    api_token: "token"
`,
			errMsg: "name is required",
		},
		{
			name: "missing url",
			config: `
targets:
  - name: "dns"
    api_token: "token"
`,
			errMsg: "url is required",
		},
		{
			name: "missing api_token",
			config: `
targets:
  - name: "dns"
    url: "http://localhost:5380"
`,
			errMsg: "api_token is required",
		},
		{
			name: "no targets",
			config: ``,
			errMsg: "at least one target",
		},
		{
			name:   "invalid log level",
			config: "exporter:\n  log_level: invalid\ntargets:\n  - name: dns\n    url: http://localhost:5380\n    api_token: token\n",
			errMsg: "invalid log_level",
		},
		{
			name: "reserved label name",
			config: `
targets:
  - name: dns
    url: http://localhost:5380
    api_token: token
    labels:
      node: primary
`,
			errMsg: "conflicts with a reserved Prometheus variable label name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.config), 0644); err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			_, err := Load(tmpFile)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestEnvVarInterpolation(t *testing.T) {
	os.Setenv("TEST_TOKEN", "my-secret-token")
	defer os.Unsetenv("TEST_TOKEN")

	configYAML := `
targets:
  - name: "dns"
    url: "http://localhost:5380"
    api_token: "${TEST_TOKEN}"
`
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Targets[0].APIToken != "my-secret-token" {
		t.Errorf("expected api_token 'my-secret-token', got %q", cfg.Targets[0].APIToken)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
