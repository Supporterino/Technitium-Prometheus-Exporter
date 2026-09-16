package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"technitium-dns-exporter/internal/config"
)

func dashboardResponse(labels []string, values []int64, total, noError int64) map[string]interface{} {
	data := make([]int64, len(values))
	copy(data, values)

	return map[string]interface{}{
		"status": "ok",
		"response": map[string]interface{}{
			"stats": map[string]interface{}{
				"totalQueries": total,
				"totalNoError": noError,
			},
			"mainChartData": map[string]interface{}{
				"labels": labels,
				"datasets": []interface{}{
					map[string]interface{}{"label": "Total", "data": data},
				},
			},
		},
	}
}

func gatherValues(t *testing.T, reg *prometheus.Registry, names ...string) map[string]float64 {
	t.Helper()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	wanted := make(map[string]bool, len(names))
	for _, n := range names {
		wanted[n] = true
	}

	values := make(map[string]float64, len(names))
	for _, mf := range mfs {
		if !wanted[mf.GetName()] {
			continue
		}
		if len(mf.Metric) != 1 {
			t.Fatalf("%s: expected 1 metric, got %d", mf.GetName(), len(mf.Metric))
		}
		values[mf.GetName()] = mf.Metric[0].GetCounter().GetValue()
	}
	return values
}

func TestCollectorCumulativeQueryCounters(t *testing.T) {
	labels1 := []string{"2024-01-01T00:00:00Z", "2024-01-01T00:01:00Z", "2024-01-01T00:02:00Z"}
	labels2 := []string{"2024-01-01T00:00:00Z", "2024-01-01T00:01:00Z", "2024-01-01T00:02:00Z", "2024-01-01T00:03:00Z"}

	var call atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/dashboard/stats/get" {
			if call.Add(1) == 1 {
				json.NewEncoder(w).Encode(dashboardResponse(labels1, []int64{10, 20, 30}, 60, 50))
			} else {
				json.NewEncoder(w).Encode(dashboardResponse(labels2, []int64{10, 20, 30, 40}, 100, 80))
			}
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "response": map[string]interface{}{}})
	}))
	defer ts.Close()

	target := config.Target{Name: "test-instance", URL: ts.URL, APIToken: "test-token", Labels: map[string]string{}}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	// First scrape seeds the cumulative counters from the window total.
	values := gatherValues(t, registry, "technitium_dns_queries_total", "technitium_dns_queries_noerror_total")
	if got := values["technitium_dns_queries_total"]; got != 60 {
		t.Errorf("first scrape queries_total = %v, want 60", got)
	}
	if got := values["technitium_dns_queries_noerror_total"]; got != 50 {
		t.Errorf("first scrape queries_noerror_total = %v, want 50", got)
	}

	// Second scrape adds only the new bucket (40) to the exact total and the
	// proportional share of the breakdown: 50 + 40*80/100 = 82.
	values = gatherValues(t, registry, "technitium_dns_queries_total", "technitium_dns_queries_noerror_total")
	if got := values["technitium_dns_queries_total"]; got != 100 {
		t.Errorf("second scrape queries_total = %v, want 100", got)
	}
	if got := values["technitium_dns_queries_noerror_total"]; got != 82 {
		t.Errorf("second scrape queries_noerror_total = %v, want 82", got)
	}

	// Re-scraping the same window must not double count.
	values = gatherValues(t, registry, "technitium_dns_queries_total", "technitium_dns_queries_noerror_total")
	if got := values["technitium_dns_queries_total"]; got != 100 {
		t.Errorf("third scrape queries_total = %v, want 100 (no double count)", got)
	}
	if got := values["technitium_dns_queries_noerror_total"]; got != 82 {
		t.Errorf("third scrape queries_noerror_total = %v, want 82 (no double count)", got)
	}
}

func TestCollectorScrapeSuccessIsOneWhenHealthy(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{Name: "test-instance", URL: ts.URL, APIToken: "test-token", Labels: map[string]string{}}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	mfs, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "technitium_dns_scrape_success" {
			found = true
			if len(mf.Metric) != 1 || mf.Metric[0].GetGauge().GetValue() != 1 {
				t.Errorf("expected scrape_success=1 for healthy target")
			}
		}
	}
	if !found {
		t.Error("technitium_dns_scrape_success not found")
	}
}
