package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"technitium-dns-exporter/internal/client"
	"technitium-dns-exporter/internal/config"
)

type TechnitiumCollector struct {
	client *client.APIClient
	target config.Target
	logger *slog.Logger
	timeout time.Duration

	descScrapeSuccess  *prometheus.Desc
	descScrapeDuration *prometheus.Desc

	descQueryTotal     *prometheus.Desc
	descQueryNoError   *prometheus.Desc
	descQueryServFail  *prometheus.Desc
	descQueryNXDomain  *prometheus.Desc
	descQueryRefused   *prometheus.Desc
	descQueryAuth      *prometheus.Desc
	descQueryRecursive *prometheus.Desc
	descQueryCached    *prometheus.Desc
	descQueryBlocked   *prometheus.Desc
	descQueryDropped   *prometheus.Desc

	descTotalClients *prometheus.Desc
	descCachedEntries *prometheus.Desc
	descZones         *prometheus.Desc
	descAllowedZones  *prometheus.Desc
	descBlockedZones  *prometheus.Desc
	descAllowListZones *prometheus.Desc
	descBlockListZones *prometheus.Desc

	descZoneInfo *prometheus.Desc

	descDHCPLeases      *prometheus.Desc
	descDHCPScopeEnabled *prometheus.Desc

	descClusterNodeState *prometheus.Desc
	descHeartbeatInterval *prometheus.Desc
}

func New(target config.Target, timeout time.Duration, logger *slog.Logger) *TechnitiumCollector {
	apiClient := client.New(target)
	labels := target.Labels
	labels["instance"] = target.Name

	return &TechnitiumCollector{
		client:   apiClient,
		target:   target,
		logger:   logger,
		timeout:  timeout,

		descScrapeSuccess: prometheus.NewDesc(
			"technitium_dns_scrape_success",
			"Whether the scrape against the target was successful.",
			nil, labels,
		),
		descScrapeDuration: prometheus.NewDesc(
			"technitium_dns_scrape_duration_seconds",
			"Duration of the scrape against the target.",
			nil, labels,
		),

		descQueryTotal: prometheus.NewDesc(
			"technitium_dns_queries_total",
			"Total number of DNS queries.",
			nil, labels,
		),
		descQueryNoError: prometheus.NewDesc(
			"technitium_dns_queries_noerror_total",
			"Total number of DNS queries with NOERROR response.",
			nil, labels,
		),
		descQueryServFail: prometheus.NewDesc(
			"technitium_dns_queries_servfail_total",
			"Total number of DNS queries with SERVFAIL response.",
			nil, labels,
		),
		descQueryNXDomain: prometheus.NewDesc(
			"technitium_dns_queries_nxdomain_total",
			"Total number of DNS queries with NXDOMAIN response.",
			nil, labels,
		),
		descQueryRefused: prometheus.NewDesc(
			"technitium_dns_queries_refused_total",
			"Total number of DNS queries with REFUSED response.",
			nil, labels,
		),
		descQueryAuth: prometheus.NewDesc(
			"technitium_dns_queries_authoritative_total",
			"Total number of authoritative DNS queries.",
			nil, labels,
		),
		descQueryRecursive: prometheus.NewDesc(
			"technitium_dns_queries_recursive_total",
			"Total number of recursive DNS queries.",
			nil, labels,
		),
		descQueryCached: prometheus.NewDesc(
			"technitium_dns_queries_cached_total",
			"Total number of cached DNS queries.",
			nil, labels,
		),
		descQueryBlocked: prometheus.NewDesc(
			"technitium_dns_queries_blocked_total",
			"Total number of blocked DNS queries.",
			nil, labels,
		),
		descQueryDropped: prometheus.NewDesc(
			"technitium_dns_queries_dropped_total",
			"Total number of dropped DNS queries.",
			nil, labels,
		),

		descTotalClients: prometheus.NewDesc(
			"technitium_dns_clients_active",
			"Number of active clients.",
			nil, labels,
		),
		descCachedEntries: prometheus.NewDesc(
			"technitium_dns_cache_entries",
			"Number of cache entries.",
			nil, labels,
		),
		descZones: prometheus.NewDesc(
			"technitium_dns_zones_count",
			"Number of authoritative zones.",
			nil, labels,
		),
		descAllowedZones: prometheus.NewDesc(
			"technitium_dns_allowed_zones_count",
			"Number of allowed zones.",
			nil, labels,
		),
		descBlockedZones: prometheus.NewDesc(
			"technitium_dns_blocked_zones_count",
			"Number of blocked zones.",
			nil, labels,
		),
		descAllowListZones: prometheus.NewDesc(
			"technitium_dns_allowlist_zones_count",
			"Number of allow list zones.",
			nil, labels,
		),
		descBlockListZones: prometheus.NewDesc(
			"technitium_dns_blocklist_zones_count",
			"Number of block list zones.",
			nil, labels,
		),

		descZoneInfo: prometheus.NewDesc(
			"technitium_dns_zone_info",
			"Information about a DNS zone.",
			[]string{"zone", "type", "dnssec_status"}, labels,
		),

		descDHCPLeases: prometheus.NewDesc(
			"technitium_dns_dhcp_leases_count",
			"Number of DHCP leases per scope.",
			[]string{"scope"}, labels,
		),
		descDHCPScopeEnabled: prometheus.NewDesc(
			"technitium_dns_dhcp_scope_enabled",
			"Whether a DHCP scope is enabled.",
			[]string{"scope"}, labels,
		),

		descClusterNodeState: prometheus.NewDesc(
			"technitium_dns_cluster_node_state",
			"State of a cluster node.",
			[]string{"node", "node_type", "ip_address"}, labels,
		),
		descHeartbeatInterval: prometheus.NewDesc(
			"technitium_dns_cluster_heartbeat_interval_seconds",
			"Cluster heartbeat refresh interval in seconds.",
			nil, labels,
		),
	}
}

func (c *TechnitiumCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.descScrapeSuccess
	ch <- c.descScrapeDuration
	ch <- c.descQueryTotal
	ch <- c.descQueryNoError
	ch <- c.descQueryServFail
	ch <- c.descQueryNXDomain
	ch <- c.descQueryRefused
	ch <- c.descQueryAuth
	ch <- c.descQueryRecursive
	ch <- c.descQueryCached
	ch <- c.descQueryBlocked
	ch <- c.descQueryDropped
	ch <- c.descTotalClients
	ch <- c.descCachedEntries
	ch <- c.descZones
	ch <- c.descAllowedZones
	ch <- c.descBlockedZones
	ch <- c.descAllowListZones
	ch <- c.descBlockListZones
	ch <- c.descZoneInfo
	ch <- c.descDHCPLeases
	ch <- c.descDHCPScopeEnabled
	ch <- c.descClusterNodeState
	ch <- c.descHeartbeatInterval
}

func (c *TechnitiumCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	start := time.Now()
	scrapeSuccess := float64(1)

	var wg sync.WaitGroup

	collectors := []func(context.Context, chan<- prometheus.Metric){
		c.collectDashboardStats,
		c.collectZones,
		c.collectCacheStats,
		c.collectBlocklistStats,
		c.collectForwarderStats,
		c.collectDHCP,
		c.collectCluster,
	}

	for _, collect := range collectors {
		wg.Add(1)
		go func(collectFn func(context.Context, chan<- prometheus.Metric)) {
			defer wg.Done()
			collectFn(ctx, ch)
		}(collect)
	}

	wg.Wait()

	duration := time.Since(start).Seconds()

	ch <- prometheus.MustNewConstMetric(c.descScrapeSuccess, prometheus.GaugeValue, scrapeSuccess)
	ch <- prometheus.MustNewConstMetric(c.descScrapeDuration, prometheus.GaugeValue, duration)
}

func (c *TechnitiumCollector) logError(msg string, err error) {
	c.logger.Error(msg, "target", c.target.Name, "error", err)
}

func (c *TechnitiumCollector) logDebug(msg string, args ...any) {
	allArgs := append([]any{"target", c.target.Name}, args...)
	c.logger.Debug(msg, allArgs...)
}
