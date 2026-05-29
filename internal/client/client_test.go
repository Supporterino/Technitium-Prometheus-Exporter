package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"technitium-dns-exporter/internal/config"
)

func newClientTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := r.ParseForm(); err != nil {
			http.Error(w, `{"status":"error"}`, http.StatusBadRequest)
			return
		}

		token := r.FormValue("token")
		if token != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"response": map[string]interface{}{
					"errorMessage": "Invalid token",
				},
			})
			return
		}

		path := r.URL.Path
		switch {
		case strings.Contains(path, "/api/dashboard/stats/get"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"stats": map[string]interface{}{
						"totalQueries":      1000,
						"totalNoError":      800,
						"totalServerFailure": 5,
						"totalNxDomain":     100,
						"totalRefused":      10,
						"totalAuthoritative": 200,
						"totalRecursive":    300,
						"totalCached":       400,
						"totalBlocked":      50,
						"totalDropped":      5,
						"totalClients":      15,
						"zones":             10,
						"cachedEntries":     5000,
						"allowedZones":      5,
						"blockedZones":      3,
						"allowListZones":    2,
						"blockListZones":    1000,
					},
				},
			})
		case strings.Contains(path, "/api/zones/list"):
			json.NewEncoder(w).Encode(map[string]interface{}{
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
					},
				},
			})
		case strings.Contains(path, "/api/settings/get"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"version":                   "15.0",
					"uptimestamp":               "2024-01-01T00:00:00Z",
					"enableBlocking":            true,
					"blockListNextUpdatedOn":    "2024-01-15T00:00:00Z",
					"blockListUpdateIntervalHours": 24,
					"blockingType":              "AnyAddress",
					"blockingAnswerTtl":         60,
					"cacheMaximumEntries":       10000,
					"cacheMinimumRecordTtl":     60,
					"cacheMaximumRecordTtl":     86400,
					"cacheNegativeRecordTtl":    300,
					"cacheFailureRecordTtl":     60,
					"cachePrefetchEligibility":  2,
					"cachePrefetchTrigger":      5,
					"saveCache":                true,
					"serveStale":               true,
					"serveStaleTtl":             259200,
					"serveStaleAnswerTtl":       30,
					"serveStaleResetTtl":        60,
					"serveStaleMaxWaitTime":     2000,
					"defaultRecordTtl":          3600,
					"defaultNsRecordTtl":        14400,
					"defaultSoaRecordTtl":       900,
					"dnssecValidation":          true,
					"preferIPv6":                false,
					"ipv6Mode":                  "Enabled",
					"forwarders":                []string{"8.8.8.8"},
					"forwarderProtocol":         "Udp",
				},
			})
		case strings.Contains(path, "/api/admin/cluster/state"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"clusterInitialized":               true,
					"dnsServerDomain":                  "dns.example.com",
					"version":                          "15.0",
					"clusterDomain":                    "cluster.example.com",
					"heartbeatRefreshIntervalSeconds":  30,
					"heartbeatRetryIntervalSeconds":    5,
					"configRefreshIntervalSeconds":     60,
					"configRetryIntervalSeconds":       10,
					"configLastSynced":                 "2024-01-01T00:00:00Z",
					"nodes": []interface{}{
						map[string]interface{}{
							"id":        1,
							"name":      "node1",
							"url":       "https://node1:5380",
							"ipAddress": "10.0.0.1",
							"type":      "primary",
							"state":     "Self",
							"lastSeen":  "2024-01-01T00:00:00Z",
						},
					},
				},
			})
		case strings.Contains(path, "/api/dhcp/scopes/list"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"scopes": []interface{}{
						map[string]interface{}{
							"name":             "LAN",
							"enabled":          true,
							"startingAddress":  "192.168.1.100",
							"endingAddress":    "192.168.1.200",
							"subnetMask":       "255.255.255.0",
							"networkAddress":   "192.168.1.0",
							"broadcastAddress": "192.168.1.255",
						},
					},
				},
			})
		case strings.Contains(path, "/api/dhcp/leases/list"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"response": map[string]interface{}{
					"leases": []interface{}{
						map[string]interface{}{
							"scope":           "LAN",
							"type":            "Dynamic",
							"hardwareAddress": "00:11:22:33:44:55",
							"address":         "192.168.1.100",
							"hostName":        "client1",
							"leaseObtained":   "2024-01-01T00:00:00Z",
							"leaseExpires":    "2024-01-02T00:00:00Z",
						},
						map[string]interface{}{
							"scope":           "LAN",
							"type":            "Reserved",
							"hardwareAddress": "00:11:22:33:44:66",
							"address":         "192.168.1.101",
							"hostName":        "client2",
							"leaseObtained":   "2024-01-01T00:00:00Z",
							"leaseExpires":    "2024-01-02T00:00:00Z",
						},
					},
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "ok",
				"response": map[string]interface{}{},
			})
		}
	}))
}

func TestNew(t *testing.T) {
	target := config.Target{
		Name:           "test",
		URL:            "https://localhost:5380",
		APIToken:       "token",
		TLSSkipVerify:  true,
	}
	client := New(target, 10*time.Second)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "https://localhost:5380" {
		t.Errorf("expected baseURL=https://localhost:5380, got %s", client.baseURL)
	}
	if client.token != "token" {
		t.Errorf("expected token=token, got %s", client.token)
	}
	if client.httpClient == nil {
		t.Fatal("expected non-nil httpClient")
	}
}

func TestDoRequestSuccess(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	raw, err := c.doRequest(context.Background(), "/api/zones/list", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestDoRequestInvalidToken(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "wrong-token",
	}

	_, err := c.doRequest(context.Background(), "/api/zones/list", nil)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestDoRequestBadURL(t *testing.T) {
	c := &APIClient{
		httpClient: &http.Client{},
		baseURL:    "http://[invalid",
		token:      "test-token",
	}

	_, err := c.doRequest(context.Background(), "/api/zones/list", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDoRequestConnectionRefused(t *testing.T) {
	c := &APIClient{
		httpClient: &http.Client{},
		baseURL:    "http://127.0.0.1:19999",
		token:      "test-token",
	}

	_, err := c.doRequest(context.Background(), "/api/zones/list", nil)
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestDoRequestErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"response": map[string]interface{}{
				"errorMessage": "something went wrong",
			},
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.doRequest(context.Background(), "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for non-ok status")
	}
	if !strings.Contains(err.Error(), "API returned status: error") {
		t.Errorf("expected error to contain 'API returned status: error', got %v", err)
	}
}

func TestDoRequestErrorStatusNoMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.doRequest(context.Background(), "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for non-ok status")
	}
	if !strings.Contains(err.Error(), "API returned status: error") {
		t.Errorf("expected generic status error, got %v", err)
	}
}

func TestDoRequestMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.doRequest(context.Background(), "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestGetDashboardStats(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	stats, err := c.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Stats.TotalQueries != 1000 {
		t.Errorf("expected TotalQueries=1000, got %d", stats.Stats.TotalQueries)
	}
}

func TestGetZones(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	zones, err := c.GetZones(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
	if zones[0].Name != "example.com" {
		t.Errorf("expected example.com, got %s", zones[0].Name)
	}
}

func TestGetZonesParseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetZones(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGetSettings(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	settings, err := c.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.Version != "15.0" {
		t.Errorf("expected version=15.0, got %s", settings.Version)
	}
	if len(settings.Forwarders) != 1 {
		t.Errorf("expected 1 forwarder, got %d", len(settings.Forwarders))
	}
}

func TestGetClusterState(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	state, err := c.GetClusterState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.ClusterInitialized {
		t.Error("expected cluster to be initialized")
	}
	if len(state.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(state.Nodes))
	}
	if state.Nodes[0].Name != "node1" {
		t.Errorf("expected node1, got %s", state.Nodes[0].Name)
	}
}

func TestGetDHCPScopes(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	scopes, err := c.GetDHCPScopes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scopes) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(scopes))
	}
	if !scopes[0].Enabled {
		t.Error("expected scope to be enabled")
	}
	if scopes[0].Name != "LAN" {
		t.Errorf("expected LAN, got %s", scopes[0].Name)
	}
}

func TestGetDHCPLeases(t *testing.T) {
	ts := newClientTestServer(t)
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "test-token",
	}

	leases, err := c.GetDHCPLeases(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leases) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(leases))
	}
	if leases[0].Scope != "LAN" {
		t.Errorf("expected LAN, got %s", leases[0].Scope)
	}
}

func TestGetDHCPLeasesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetDHCPLeases(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGetDashboardStatsParseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetDashboardStats(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGetSettingsParseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetSettings(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGetClusterStateParseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetClusterState(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGetDHCPScopesParseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": "not an object",
		})
	}))
	defer ts.Close()

	c := &APIClient{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
		token:      "",
	}

	_, err := c.GetDHCPScopes(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}
