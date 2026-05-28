package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Exporter ExporterConfig `yaml:"exporter"`
	Targets  []Target       `yaml:"targets"`
}

type ExporterConfig struct {
	ListenAddress string        `yaml:"listen_address"`
	MetricsPath   string        `yaml:"metrics_path"`
	ScrapeTimeout time.Duration `yaml:"scrape_timeout"`
	LogLevel      string        `yaml:"log_level"`
}

type Target struct {
	Name           string            `yaml:"name"`
	URL            string            `yaml:"url"`
	APIToken       string            `yaml:"api_token"`
	TLSSkipVerify  bool              `yaml:"tls_skip_verify"`
	Labels         map[string]string `yaml:"labels"`
	Features       FeatureFlags      `yaml:"features"`
}

type FeatureFlags struct {
	DHCP    bool `yaml:"dhcp"`
	Cluster bool `yaml:"cluster"`
}

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func interpolateEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := envVarPattern.ReplaceAllString(match, "$1")
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return match
	})
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	dataStr := interpolateEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(dataStr), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Exporter.ListenAddress == "" {
		c.Exporter.ListenAddress = ":9119"
	}
	if c.Exporter.MetricsPath == "" {
		c.Exporter.MetricsPath = "/metrics"
	}
	if c.Exporter.ScrapeTimeout == 0 {
		c.Exporter.ScrapeTimeout = 30 * time.Second
	}
	if c.Exporter.LogLevel == "" {
		c.Exporter.LogLevel = "info"
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Exporter.LogLevel] {
		return fmt.Errorf("invalid log_level: %q (must be debug, info, warn, or error)", c.Exporter.LogLevel)
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target must be configured")
	}

	for i, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("target[%d]: name is required", i)
		}
		if t.URL == "" {
			return fmt.Errorf("target[%d] (%s): url is required", i, t.Name)
		}
		if t.APIToken == "" {
			return fmt.Errorf("target[%d] (%s): api_token is required", i, t.Name)
		}
		if t.Labels == nil {
			t.Labels = make(map[string]string)
		}
		c.Targets[i] = t
	}

	return nil
}
