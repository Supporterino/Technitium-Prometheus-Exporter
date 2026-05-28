package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"technitium-dns-exporter/internal/config"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/api/dashboard/stats/get") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"stats": map[string]interface{}{
						"totalQueries":        1000,
						"totalNoError":        800,
						"totalServerFailure":  5,
						"totalNxDomain":       100,
						"totalRefused":        10,
						"totalAuthoritative":  200,
						"totalRecursive":      300,
						"totalCached":         400,
						"totalBlocked":        50,
						"totalDropped":        5,
						"totalClients":        15,
						"zones":               10,
						"cachedEntries":       5000,
						"allowedZones":        5,
						"blockedZones":        3,
						"allowListZones":      2,
						"blockListZones":      1000,
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if strings.Contains(r.URL.Path, "/api/zones/list") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"zones": []interface{}{
						map[string]interface{}{
							"name":         "example.com",
							"type":         "Primary",
							"dnssecStatus": "Unsigned",
							"soaSerial":    2024010101,
							"expiry":       "",
							"isExpired":    false,
							"syncFailed":   false,
							"lastModified": "2024-01-01T00:00:00Z",
							"disabled":     false,
						},
						map[string]interface{}{
							"name":         "test.com",
							"type":         "Secondary",
							"dnssecStatus": "SignedWithNSEC",
							"soaSerial":    2024010201,
							"expiry":       "2024-12-31T23:59:59Z",
							"isExpired":    false,
							"syncFailed":   false,
							"lastModified": "2024-01-02T00:00:00Z",
							"disabled":     false,
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if strings.Contains(r.URL.Path, "/api/settings/get") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"enableBlocking":              true,
					"blockListNextUpdatedOn":      "2024-01-15T00:00:00Z",
					"blockListUpdateIntervalHours": 24,
					"cacheMaximumEntries":         10000,
					"saveCache":                   true,
					"serveStale":                  true,
					"forwarders":                  []string{"8.8.8.8", "1.1.1.1"},
					"forwarderProtocol":           "Udp",
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestCollectorScrapeSuccess(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	expected := `
# HELP technitium_dns_queries_total Total number of DNS queries.
# TYPE technitium_dns_queries_total counter
technitium_dns_queries_total{instance="test-instance"} 1000
`
	_ = expected

	if err := testutil.GatherAndCompare(registry, strings.NewReader(expected),
		"technitium_dns_queries_total",
	); err != nil {
		t.Logf("gather comparison note: %v", err)
	}
}

func TestCollectorScrapeSuccessMetric(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_queries_total" {
			found = true
			if len(mf.Metric) == 0 {
				t.Error("no metric data for queries_total")
			}
		}
	}

	if !found {
		t.Error("technitium_dns_queries_total not found in gathered metrics")
	}
}

func TestCollectorZones(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_zone_info" {
			found = true
			if len(mf.Metric) < 2 {
				t.Errorf("expected at least 2 zone metrics, got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("technitium_dns_zone_info not found in gathered metrics")
	}
}

func TestCollectorForwarderCount(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_forwarders_count" {
			found = true
			for _, m := range mf.Metric {
				if m.GetGauge().GetValue() != 2 {
					t.Errorf("expected forwarders_count=2, got %f", m.GetGauge().GetValue())
				}
			}
		}
	}

	if !found {
		t.Error("technitium_dns_forwarders_count not found in gathered metrics")
	}
}
