package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"technitium-dns-exporter/internal/config"
)

type APIClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func New(target config.Target) *APIClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: target.TLSSkipVerify,
		},
	}
	return &APIClient{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
		baseURL: target.URL,
		token:   target.APIToken,
	}
}

type apiResponse struct {
	Status   string          `json:"status"`
	Response json.RawMessage `json:"response"`
}

func (c *APIClient) doRequest(ctx context.Context, apiPath string, params url.Values) (json.RawMessage, error) {
	u, err := url.Parse(c.baseURL + apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("token", c.token)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "ok" {
		var errResp struct {
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.ErrorMessage != "" {
			return nil, fmt.Errorf("API error: %s", errResp.ErrorMessage)
		}
		return nil, fmt.Errorf("API returned status: %s", apiResp.Status)
	}

	return apiResp.Response, nil
}

func (c *APIClient) doPostRequest(ctx context.Context, apiPath string, formData url.Values) (json.RawMessage, error) {
	u, err := url.Parse(c.baseURL + apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if formData == nil {
		formData = url.Values{}
	}
	formData.Set("token", c.token)

	body := bytes.NewBufferString(formData.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "ok" {
		var errResp struct {
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.ErrorMessage != "" {
			return nil, fmt.Errorf("API error: %s", errResp.ErrorMessage)
		}
		return nil, fmt.Errorf("API returned status: %s", apiResp.Status)
	}

	return apiResp.Response, nil
}

type DashboardStats struct {
	Stats                   StatsData            `json:"stats"`
	TopClients              []TopClient          `json:"topClients"`
	TopDomains              []TopDomain          `json:"topDomains"`
	TopBlockedDomains       []TopBlockedDomain   `json:"topBlockedDomains"`
	QueryTypeChartData      ChartData            `json:"queryTypeChartData"`
	ProtocolTypeChartData   ChartData            `json:"protocolTypeChartData"`
}

// GetDashboardStats calls /api/dashboard/stats/get?type=LastHour&utc=true
func (c *APIClient) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	params := url.Values{}
	params.Set("type", "LastHour")
	params.Set("utc", "true")

	resp, err := c.doRequest(ctx, "/api/dashboard/stats/get", params)
	if err != nil {
		return nil, err
	}

	var stats DashboardStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse dashboard stats: %w", err)
	}
	return &stats, nil
}

// StatsData represents the stats object from /api/dashboard/stats/get
type StatsData struct {
	TotalQueries        int64 `json:"totalQueries"`
	TotalNoError        int64 `json:"totalNoError"`
	TotalServerFailure  int64 `json:"totalServerFailure"`
	TotalNxDomain       int64 `json:"totalNxDomain"`
	TotalRefused        int64 `json:"totalRefused"`
	TotalAuthoritative  int64 `json:"totalAuthoritative"`
	TotalRecursive      int64 `json:"totalRecursive"`
	TotalCached         int64 `json:"totalCached"`
	TotalBlocked        int64 `json:"totalBlocked"`
	TotalDropped        int64 `json:"totalDropped"`
	TotalClients        int64 `json:"totalClients"`
	Zones               int64 `json:"zones"`
	CachedEntries       int64 `json:"cachedEntries"`
	AllowedZones        int64 `json:"allowedZones"`
	BlockedZones        int64 `json:"blockedZones"`
	AllowListZones      int64 `json:"allowListZones"`
	BlockListZones      int64 `json:"blockListZones"`
}

type ChartData struct {
	Labels   []string    `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

type ChartDataset struct {
	Data            []int64  `json:"data"`
	Label           string   `json:"label"`
	BackgroundColor []string `json:"backgroundColor,omitempty"`
}

type TopClient struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Hits        int64  `json:"hits"`
	RateLimited bool   `json:"rateLimited"`
}

type TopDomain struct {
	Name string `json:"name"`
	Hits int64  `json:"hits"`
}

type TopBlockedDomain struct {
	Name string `json:"name"`
	Hits int64  `json:"hits"`
}

// Zone represents a zone from /api/zones/list
type Zone struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DNSSECStatus string `json:"dnssecStatus"`
	SOASerial    int64  `json:"soaSerial"`
	Expiry       string `json:"expiry"`
	IsExpired    bool   `json:"isExpired"`
	SyncFailed   bool   `json:"syncFailed"`
	LastModified string `json:"lastModified"`
	Disabled     bool   `json:"disabled"`
	Internal     bool   `json:"internal"`
}

// GetZones calls /api/zones/list
func (c *APIClient) GetZones(ctx context.Context) ([]Zone, error) {
	resp, err := c.doRequest(ctx, "/api/zones/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Zones []Zone `json:"zones"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse zones: %w", err)
	}
	return result.Zones, nil
}

// DNSSettings represents settings from /api/settings/get
type DNSSettings struct {
	Version                      string              `json:"version"`
	Uptimestamp                  string              `json:"uptimestamp"`
	ClusterInitialized           bool                `json:"clusterInitialized"`
	DNSOverUDPEnabled            bool                `json:"enableDnsOverUdpProxy"`
	DNSOverTCPEnabled            bool                `json:"enableDnsOverTcpProxy"`
	DNSOverTLSEnabled            bool                `json:"enableDnsOverTls"`
	DNSOverHTTPSEnabled          bool                `json:"enableDnsOverHttps"`
	DNSOverHTTPEnabled           bool                `json:"enableDnsOverHttp"`
	DNSOverQUICEnabled           bool                `json:"enableDnsOverQuic"`
	DNSOverHTTP3Enabled          bool                `json:"enableDnsOverHttp3"`
	Recursion                    string              `json:"recursion"`
	EnableBlocking               bool                `json:"enableBlocking"`
	BlockListNextUpdatedOn       string              `json:"blockListNextUpdatedOn"`
	BlockListUpdateIntervalHours int64               `json:"blockListUpdateIntervalHours"`
	CacheMaximumEntries          int64               `json:"cacheMaximumEntries"`
	Forwarders                   []string            `json:"forwarders"`
	ForwarderProtocol            string              `json:"forwarderProtocol"`
	SaveCache                    bool                `json:"saveCache"`
	ServeStale                   bool                `json:"serveStale"`
	DHCPServerEnabled            bool                `json:"dhcpServerEnabled"`
	QPMPrefixLimitsIPv4          []QPMPrefixLimit    `json:"qpmPrefixLimitsIPv4"`
	QPMPrefixLimitsIPv6          []QPMPrefixLimit    `json:"qpmPrefixLimitsIPv6"`
}

type QPMPrefixLimit struct {
	Prefix   int64 `json:"prefix"`
	UDPLimit int64 `json:"udpLimit"`
	TCPLimit int64 `json:"tcpLimit"`
}

// GetSettings calls /api/settings/get
func (c *APIClient) GetSettings(ctx context.Context) (*DNSSettings, error) {
	resp, err := c.doRequest(ctx, "/api/settings/get", nil)
	if err != nil {
		return nil, err
	}

	var settings DNSSettings
	if err := json.Unmarshal(resp, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}
	return &settings, nil
}

// ClusterState represents cluster state from /api/admin/cluster/state
type ClusterState struct {
	ClusterInitialized             bool          `json:"clusterInitialized"`
	DNSDomain                      string        `json:"dnsServerDomain"`
	Version                        string        `json:"version"`
	ClusterDomain                  string        `json:"clusterDomain"`
	HeartbeatRefreshInterval       int64         `json:"heartbeatRefreshIntervalSeconds"`
	HeartbeatRetryInterval         int64         `json:"heartbeatRetryIntervalSeconds"`
	ConfigRefreshInterval          int64         `json:"configRefreshIntervalSeconds"`
	ConfigRetryInterval            int64         `json:"configRetryIntervalSeconds"`
	ConfigLastSynced               string        `json:"configLastSynced"`
	Nodes                          []ClusterNode `json:"nodes"`
}

type ClusterNode struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	IPAddress string `json:"ipAddress"`
	Type      string `json:"type"`
	State     string `json:"state"`
	LastSeen  string `json:"lastSeen"`
}

// GetClusterState calls /api/admin/cluster/state
func (c *APIClient) GetClusterState(ctx context.Context) (*ClusterState, error) {
	resp, err := c.doRequest(ctx, "/api/admin/cluster/state", nil)
	if err != nil {
		return nil, err
	}

	var state ClusterState
	if err := json.Unmarshal(resp, &state); err != nil {
		return nil, fmt.Errorf("failed to parse cluster state: %w", err)
	}
	return &state, nil
}

// Lease represents a DHCP lease from /api/dhcp/leases/list
type Lease struct {
	Scope           string `json:"scope"`
	Type            string `json:"type"`
	HardwareAddress string `json:"hardwareAddress"`
	ClientIdentifier string `json:"clientIdentifier"`
	Address         string `json:"address"`
	HostName        string `json:"hostName"`
	LeaseObtained   string `json:"leaseObtained"`
	LeaseExpires    string `json:"leaseExpires"`
}

// GetDHCPScopes calls /api/dhcp/scopes/list
func (c *APIClient) GetDHCPScopes(ctx context.Context) ([]DHCPScope, error) {
	resp, err := c.doRequest(ctx, "/api/dhcp/scopes/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Scopes []DHCPScope `json:"scopes"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DHCP scopes: %w", err)
	}
	return result.Scopes, nil
}

type DHCPScope struct {
	Name             string `json:"name"`
	Enabled          bool   `json:"enabled"`
	StartingAddress  string `json:"startingAddress"`
	EndingAddress    string `json:"endingAddress"`
	SubnetMask       string `json:"subnetMask"`
	NetworkAddress   string `json:"networkAddress"`
	BroadcastAddress string `json:"broadcastAddress"`
}

// GetDHCPLeases calls /api/dhcp/leases/list
func (c *APIClient) GetDHCPLeases(ctx context.Context) ([]Lease, error) {
	resp, err := c.doRequest(ctx, "/api/dhcp/leases/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Leases []Lease `json:"leases"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DHCP leases: %w", err)
	}
	return result.Leases, nil
}
