package collector

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"technitium-dns-exporter/internal/config"
)

func BenchmarkCollectorNew(b *testing.B) {
	target := config.Target{
		Name:     "bench-instance",
		URL:      "https://localhost:5380",
		APIToken: "bench-token",
		Labels:   map[string]string{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(target, 30*time.Second, nil)
	}
}

func BenchmarkCollectorCollect(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	}))
	defer ts.Close()

	target := config.Target{
		Name:     "bench-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	logger := slog.New(slog.DiscardHandler)
	c := New(target, 30*time.Second, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 256)
		go func() {
			for range ch {
			}
		}()
		c.Collect(ch)
		close(ch)
	}
}

func BenchmarkCollectorDescribe(b *testing.B) {
	target := config.Target{
		Name:     "bench-instance",
		URL:      "https://localhost:5380",
		APIToken: "bench-token",
		Labels:   map[string]string{},
	}
	c := New(target, 30*time.Second, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan *prometheus.Desc, 256)
		go func() {
			for range ch {
			}
		}()
		c.Describe(ch)
		close(ch)
	}
}
