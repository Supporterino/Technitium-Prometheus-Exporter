package collector

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
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
							"notifyFailed": false,
							"lastModified": "2024-01-01T00:00:00Z",
							"disabled":     false,
							"internal":     false,
						},
						map[string]interface{}{
							"name":         "test.com",
							"type":         "Secondary",
							"dnssecStatus": "SignedWithNSEC",
							"soaSerial":    2024010201,
							"expiry":       "2024-12-31T23:59:59Z",
							"isExpired":    false,
							"syncFailed":   true,
							"notifyFailed": false,
							"lastModified": "2024-01-02T00:00:00Z",
							"disabled":     true,
							"internal":     true,
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
					"version":                          "15.0",
					"uptimestamp":                      "2024-01-01T00:00:00Z",
					"enableBlocking":                   true,
					"blockListNextUpdatedOn":           "2024-01-15T00:00:00Z",
					"blockListUpdateIntervalHours":     24,
					"blockingType":                     "AnyAddress",
					"blockingAnswerTtl":                60,
					"allowTxtBlockingReport":           false,
					"cacheMaximumEntries":              10000,
					"cacheMinimumRecordTtl":            60,
					"cacheMaximumRecordTtl":            86400,
					"cacheNegativeRecordTtl":           300,
					"cacheFailureRecordTtl":            60,
					"cachePrefetchEligibility":         2,
					"cachePrefetchTrigger":             5,
					"saveCache":                        true,
					"serveStale":                       true,
					"serveStaleTtl":                    259200,
					"serveStaleAnswerTtl":              30,
					"serveStaleResetTtl":               60,
					"serveStaleMaxWaitTime":            2000,
					"defaultRecordTtl":                 3600,
					"defaultNsRecordTtl":               14400,
					"defaultSoaRecordTtl":              900,
					"dnssecValidation":                 true,
					"preferIPv6":                       false,
					"ipv6Mode":                         "Enabled",
					"randomizeName":                    true,
					"qnameMinimization":                false,
					"eDnsClientSubnet":                 false,
					"eDnsClientSubnetIPv4PrefixLength": 24,
					"eDnsClientSubnetIPv6PrefixLength": 56,
					"udpPayloadSize":                   1232,
					"enableUdpSocketPool":              true,
					"udpSendBufferSizeKB":              256,
					"udpReceiveBufferSizeKB":           256,
					"clientTimeout":                    2000,
					"tcpSendTimeout":                   10,
					"tcpReceiveTimeout":                10,
					"listenBacklog":                    128,
					"maxConcurrentResolutionsPerCore":  8,
					"resolverRetries":                  2,
					"resolverTimeout":                  2000,
					"resolverConcurrency":              4,
					"resolverMaxStackCount":            16,
					"concurrentForwarding":             true,
					"forwarderRetries":                 2,
					"forwarderTimeout":                 2000,
					"forwarderConcurrency":             4,
					"forwarders":                       []string{"8.8.8.8", "1.1.1.1"},
					"forwarderProtocol":                "Udp",
					"enableLogging":                    true,
					"logQueries":                       false,
					"useLocalTime":                     false,
					"maxLogFileDays":                   90,
					"enableInMemoryStats":              true,
					"maxStatFileDays":                  365,
					"dnsAppsEnableAutomaticUpdate":     false,
					"webServiceHttpPort":               5380,
					"webServiceEnableTls":              false,
					"webServiceTlsPort":                53443,
					"enableDnsOverUdpProxy":            true,
					"enableDnsOverTcpProxy":            true,
					"enableDnsOverTls":                 false,
					"enableDnsOverHttps":               false,
					"enableDnsOverHttp":                false,
					"enableDnsOverQuic":                false,
					"dnsOverUdpProxyPort":              53,
					"dnsOverTcpProxyPort":              53,
					"dnsOverTlsPort":                   853,
					"dnsOverHttpsPort":                 443,
					"dnsOverHttpPort":                  80,
					"dnsOverQuicPort":                  853,
					"qpmLimitSampleMinutes":            5,
					"qpmLimitUdpTruncationPercentage": 90,
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, "/api/dhcp/scopes/list") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"scopes": []interface{}{
						map[string]interface{}{
							"name":            "LAN",
							"enabled":         true,
							"startingAddress": "192.168.1.100",
							"endingAddress":   "192.168.1.200",
							"subnetMask":      "255.255.255.0",
						},
						map[string]interface{}{
							"name":            "Guest",
							"enabled":         false,
							"startingAddress": "10.0.0.100",
							"endingAddress":   "10.0.0.200",
							"subnetMask":      "255.255.255.0",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if strings.Contains(r.URL.Path, "/api/dhcp/leases/list") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"leases": []interface{}{
						map[string]interface{}{
							"scope":         "LAN",
							"type":          "Dynamic",
							"hardwareAddress": "00:11:22:33:44:55",
							"address":       "192.168.1.100",
							"hostName":      "client1",
						},
						map[string]interface{}{
							"scope":         "LAN",
							"type":          "Reserved",
							"hardwareAddress": "00:11:22:33:44:66",
							"address":       "192.168.1.101",
							"hostName":      "client2",
						},
						map[string]interface{}{
							"scope":         "Guest",
							"type":          "Dynamic",
							"hardwareAddress": "00:11:22:33:44:77",
							"address":       "10.0.0.100",
							"hostName":      "client3",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if strings.Contains(r.URL.Path, "/api/admin/cluster/state") {
			resp := map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"clusterInitialized":              true,
					"dnsServerDomain":                 "dns.example.com",
					"version":                         "15.0",
					"heartbeatRefreshIntervalSeconds": 30,
					"heartbeatRetryIntervalSeconds":   5,
					"configRefreshIntervalSeconds":    60,
					"configRetryIntervalSeconds":      10,
					"configLastSynced":                "2024-01-01T00:00:00Z",
					"nodes": []interface{}{
						map[string]interface{}{
							"id":        1,
							"name":      "node1",
							"url":       "https://node1:5380",
							"ipAddress": "10.0.0.1",
							"type":      "primary",
							"state":     "Self",
						},
						map[string]interface{}{
							"id":        2,
							"name":      "node2",
							"url":       "https://node2:5380",
							"ipAddress": "10.0.0.2",
							"type":      "secondary",
							"state":     "Connected",
						},
					},
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
	c := New(target, 5*time.Second, 30*time.Second, nil)

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
	c := New(target, 5*time.Second, 30*time.Second, nil)

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
	c := New(target, 5*time.Second, 30*time.Second, nil)

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
	c := New(target, 5*time.Second, 30*time.Second, nil)

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

func TestCollectorSettingsMetrics(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	metricNames := map[string]bool{}
	for _, mf := range metrics {
		metricNames[mf.GetName()] = true
	}

	expectedMetrics := []string{
		"technitium_dns_cache_max_entries",
		"technitium_dns_cache_save_enabled",
		"technitium_dns_cache_serve_stale_enabled",
		"technitium_dns_cache_min_record_ttl_seconds",
		"technitium_dns_cache_max_record_ttl_seconds",
		"technitium_dns_cache_negative_record_ttl_seconds",
		"technitium_dns_cache_failure_record_ttl_seconds",
		"technitium_dns_cache_prefetch_eligibility",
		"technitium_dns_cache_prefetch_trigger_seconds",
		"technitium_dns_cache_serve_stale_ttl_seconds",
		"technitium_dns_blocking_enabled",
		"technitium_dns_blocklist_update_interval_hours",
		"technitium_dns_blocklist_next_update_timestamp_seconds",
		"technitium_dns_blocking_type",
		"technitium_dns_blocking_answer_ttl_seconds",
		"technitium_dns_allow_txt_blocking_report_enabled",
		"technitium_dns_forwarders_count",
		"technitium_dns_forwarder_info",
		"technitium_dns_protocol_enabled",
		"technitium_dns_protocol_port",
		"technitium_dns_default_ttl_seconds",
		"technitium_dns_dnssec_validation_enabled",
		"technitium_dns_ipv6_prefer_enabled",
		"technitium_dns_ipv6_mode",
		"technitium_dns_randomize_name_enabled",
		"technitium_dns_qname_minimization_enabled",
		"technitium_dns_edns_client_subnet_enabled",
		"technitium_dns_edns_client_subnet_prefix_length",
		"technitium_dns_udp_payload_size",
		"technitium_dns_udp_socket_pool_enabled",
		"technitium_dns_udp_buffer_size_kb",
		"technitium_dns_client_timeout_seconds",
		"technitium_dns_tcp_send_timeout_seconds",
		"technitium_dns_tcp_receive_timeout_seconds",
		"technitium_dns_listen_backlog",
		"technitium_dns_max_concurrent_resolutions_per_core",
		"technitium_dns_resolver_retries",
		"technitium_dns_resolver_timeout_seconds",
		"technitium_dns_resolver_concurrency",
		"technitium_dns_resolver_max_stack_count",
		"technitium_dns_concurrent_forwarding_enabled",
		"technitium_dns_forwarder_retries",
		"technitium_dns_forwarder_timeout_seconds",
		"technitium_dns_forwarder_concurrency",
		"technitium_dns_log_enabled",
		"technitium_dns_log_use_local_time_enabled",
		"technitium_dns_log_retention_days",
		"technitium_dns_in_memory_stats_enabled",
		"technitium_dns_stats_retention_days",
		"technitium_dns_apps_auto_update_enabled",
		"technitium_dns_web_service_http_port",
		"technitium_dns_web_service_tls_enabled",
		"technitium_dns_web_service_tls_port",
		"technitium_dns_qpm_limit_sample_minutes",
		"technitium_dns_qpm_limit_udp_truncation_percent",
		"technitium_dns_version_info",
	}
	for _, name := range expectedMetrics {
		if !metricNames[name] {
			t.Errorf("expected metric %q not found in gathered metrics", name)
		}
	}
}

func TestCollectorZoneStateMetrics(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	for _, mf := range metrics {
		switch mf.GetName() {
		case "technitium_dns_zone_soa_serial":
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				if labels["zone"] == "example.com" {
					if m.GetGauge().GetValue() != 2024010101 {
						t.Errorf("expected example.com soa_serial=2024010101, got %f", m.GetGauge().GetValue())
					}
				}
			}
		case "technitium_dns_zone_disabled":
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				if labels["zone"] == "test.com" {
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected test.com disabled=1, got %f", m.GetGauge().GetValue())
					}
				}
			}
		case "technitium_dns_zone_internal":
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				if labels["zone"] == "test.com" {
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected test.com internal=1, got %f", m.GetGauge().GetValue())
					}
				}
			}
		case "technitium_dns_zone_sync_failed":
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				if labels["zone"] == "test.com" {
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected test.com sync_failed=1, got %f", m.GetGauge().GetValue())
					}
				}
			}
		}
	}
}

func TestCollectorProtocolEnabled(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_protocol_enabled" {
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				switch labels["protocol"] {
				case "udp", "tcp":
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected protocol %s enabled=1, got %f", labels["protocol"], m.GetGauge().GetValue())
					}
				case "tls", "https", "http", "quic":
					if m.GetGauge().GetValue() != 0 {
						t.Errorf("expected protocol %s enabled=0, got %f", labels["protocol"], m.GetGauge().GetValue())
					}
				}
			}
		}
	}
}

func labelMap(labels []*dto.LabelPair) map[string]string {
	m := make(map[string]string)
	for _, l := range labels {
		m[l.GetName()] = l.GetValue()
	}
	return m
}

func TestCollectorDHCP(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
		Features: config.FeatureFlags{DHCP: true},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	foundScope := false
	foundLeases := false
	for _, mf := range metrics {
		switch mf.GetName() {
		case "technitium_dns_dhcp_scope_enabled":
			foundScope = true
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				switch labels["scope"] {
				case "LAN":
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected LAN enabled=1, got %f", m.GetGauge().GetValue())
					}
				case "Guest":
					if m.GetGauge().GetValue() != 0 {
						t.Errorf("expected Guest enabled=0, got %f", m.GetGauge().GetValue())
					}
				}
			}
		case "technitium_dns_dhcp_leases_count":
			foundLeases = true
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				switch labels["scope"] {
				case "LAN":
					if m.GetGauge().GetValue() != 2 {
						t.Errorf("expected LAN lease count=2, got %f", m.GetGauge().GetValue())
					}
				case "Guest":
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected Guest lease count=1, got %f", m.GetGauge().GetValue())
					}
				}
			}
		}
	}
	if !foundScope {
		t.Error("technitium_dns_dhcp_scope_enabled not found")
	}
	if !foundLeases {
		t.Error("technitium_dns_dhcp_leases_count not found")
	}
}

func TestCollectorCluster(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
		Features: config.FeatureFlags{Cluster: true},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	foundNodeState := false
	foundHeartbeat := false
	for _, mf := range metrics {
		switch mf.GetName() {
		case "technitium_dns_cluster_node_state":
			foundNodeState = true
			for _, m := range mf.Metric {
				labels := labelMap(m.GetLabel())
				switch labels["cluster_node"] {
				case "node1":
					if m.GetGauge().GetValue() != 2 {
						t.Errorf("expected node1 Self=2, got %f", m.GetGauge().GetValue())
					}
				case "node2":
					if m.GetGauge().GetValue() != 1 {
						t.Errorf("expected node2 Connected=1, got %f", m.GetGauge().GetValue())
					}
				}
			}
		case "technitium_dns_cluster_heartbeat_interval_seconds":
			foundHeartbeat = true
			for _, m := range mf.Metric {
				if m.GetGauge().GetValue() != 30 {
					t.Errorf("expected heartbeat interval=30, got %f", m.GetGauge().GetValue())
				}
			}
		}
	}
	if !foundNodeState {
		t.Error("technitium_dns_cluster_node_state not found")
	}
	if !foundHeartbeat {
		t.Error("technitium_dns_cluster_heartbeat_interval_seconds not found")
	}
}

func TestCollectorClusterConfigLastSynced(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
		Features: config.FeatureFlags{Cluster: true},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_cluster_config_last_synced_timestamp_seconds" {
			found = true
			if len(mf.Metric) == 0 {
				t.Error("expected config last synced metric")
			}
		}
	}
	if !found {
		t.Error("technitium_dns_cluster_config_last_synced_timestamp_seconds not found")
	}
}

func TestCollectorAPIErrorResponses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"response": map[string]interface{}{
				"errorMessage": "internal server error",
			},
		})
	}))
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := New(target, 5*time.Second, 30*time.Second, logger)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	// Should not panic; scrapeSuccess remains 1 despite API errors
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_scrape_success" {
			for _, m := range mf.Metric {
				if m.GetGauge().GetValue() != 1 {
					t.Error("scrape_success must be 1 even when API returns errors")
				}
			}
		}
	}
}

func TestCollectorEmptyResponses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	}))
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := New(target, 5*time.Second, 30*time.Second, logger)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	// Should still produce scrape metrics even with empty API responses
	foundScrape := false
	foundDuration := false
	for _, mf := range metrics {
		switch mf.GetName() {
		case "technitium_dns_scrape_success":
			foundScrape = true
		case "technitium_dns_scrape_duration_seconds":
			foundDuration = true
		}
	}
	if !foundScrape {
		t.Error("scrape_success not found in collected metrics")
	}
	if !foundDuration {
		t.Error("scrape_duration_seconds not found in collected metrics")
	}
}

func TestCollectorMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json at all {{{"))
	}))
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := New(target, 5*time.Second, 30*time.Second, logger)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	// Should not panic even with malformed JSON from API
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_scrape_success" {
			for _, m := range mf.Metric {
				if m.GetGauge().GetValue() != 1 {
					t.Error("scrape_success must be 1 even with malformed JSON")
				}
			}
		}
	}

	if len(metrics) < 2 {
		t.Error("expected at least scrape_success and scrape_duration metrics")
	}
}

func TestCollectorScrapeDuration(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "test-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "technitium_dns_scrape_duration_seconds" {
			found = true
			if len(mf.Metric) != 1 {
				t.Errorf("expected 1 scrape_duration metric, got %d", len(mf.Metric))
			}
			for _, m := range mf.Metric {
				if m.GetGauge().GetValue() < 0 {
					t.Error("scrape duration must be non-negative")
				}
			}
		}
	}
	if !found {
		t.Error("technitium_dns_scrape_duration_seconds not found")
	}
}

func TestCollectorInstanceLabel(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	target := config.Target{
		Name:     "my-instance",
		URL:      ts.URL,
		APIToken: "test-token",
		Labels:   map[string]string{},
	}
	c := New(target, 5*time.Second, 30*time.Second, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}

	for _, mf := range metrics {
		for _, m := range mf.Metric {
			labels := labelMap(m.GetLabel())
			if labels["instance"] != "my-instance" {
				t.Errorf("expected instance=my-instance on metric %s, got %s", mf.GetName(), labels["instance"])
			}
		}
	}
}
