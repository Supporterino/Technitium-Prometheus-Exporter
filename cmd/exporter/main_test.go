package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"technitium-dns-exporter/internal/collector"
	"technitium-dns-exporter/internal/config"
)

func newMockAPI(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","response":{}}`))
	}))
}

func testExporterHandler(t *testing.T, cfg *config.Config, logger *slog.Logger) http.Handler {
	t.Helper()

	mu := &sync.Mutex{}
	var handler http.Handler = promhttp.HandlerFor(prometheus.NewRegistry(), promhttp.HandlerOpts{})

	mu.Lock()
	var registries []prometheus.Gatherer
	for _, target := range cfg.Targets {
		reg := prometheus.NewRegistry()
		c := collector.New(target, cfg.Exporter.ScrapeTimeout, logger)
		reg.MustRegister(c)
		registries = append(registries, reg)
	}
	handler = promhttp.HandlerFor(prometheus.Gatherers(registries), promhttp.HandlerOpts{})
	mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Exporter.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		h := handler
		mu.Unlock()
		h.ServeHTTP(w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`<html>
<head><title>Technitium DNS Exporter</title></head>
<body>
<h1>Technitium DNS Exporter</h1>
<p><a href="` + cfg.Exporter.MetricsPath + `">Metrics</a></p>
</body>
</html>`))
	})

	return mux
}

func TestExporterHealthz(t *testing.T) {
	api := newMockAPI(t)
	defer api.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		Exporter: config.ExporterConfig{
			ListenAddress: ":0",
			MetricsPath:   "/metrics",
			ScrapeTimeout: 30 * time.Second,
			LogLevel:      "info",
		},
		Targets: []config.Target{
			{
				Name:     "test",
				URL:      api.URL,
				APIToken: "token",
				Labels:   map[string]string{},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	handler := testExporterHandler(t, cfg, logger)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected 'ok', got %q", string(body))
	}
}

func TestExporterMetrics(t *testing.T) {
	api := newMockAPI(t)
	defer api.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		Exporter: config.ExporterConfig{
			ListenAddress: ":0",
			MetricsPath:   "/metrics",
			ScrapeTimeout: 30 * time.Second,
			LogLevel:      "info",
		},
		Targets: []config.Target{
			{
				Name:     "test",
				URL:      api.URL,
				APIToken: "token",
				Labels:   map[string]string{},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	handler := testExporterHandler(t, cfg, logger)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty metrics body")
	}
}

func TestExporterIndex(t *testing.T) {
	api := newMockAPI(t)
	defer api.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		Exporter: config.ExporterConfig{
			ListenAddress: ":0",
			MetricsPath:   "/metrics",
			ScrapeTimeout: 30 * time.Second,
			LogLevel:      "info",
		},
		Targets: []config.Target{
			{
				Name:     "test",
				URL:      api.URL,
				APIToken: "token",
				Labels:   map[string]string{},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	handler := testExporterHandler(t, cfg, logger)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Technitium DNS Exporter") {
		t.Error("expected index page content")
	}
}

func TestExporterNotFound(t *testing.T) {
	api := newMockAPI(t)
	defer api.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		Exporter: config.ExporterConfig{
			ListenAddress: ":0",
			MetricsPath:   "/metrics",
			ScrapeTimeout: 30 * time.Second,
			LogLevel:      "info",
		},
		Targets: []config.Target{
			{
				Name:     "test",
				URL:      api.URL,
				APIToken: "token",
				Labels:   map[string]string{},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	handler := testExporterHandler(t, cfg, logger)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("GET /nonexistent failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"default (unknown)", "trace", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leveler := &atomicLevel{}
			setLogLevel(leveler, tt.level)
			if leveler.Level() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, leveler.Level())
			}
		})
	}
}
