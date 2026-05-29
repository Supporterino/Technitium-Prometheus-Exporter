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
	client  *client.APIClient
	target  config.Target
	logger  *slog.Logger
	timeout time.Duration

	descs         []*prometheus.Desc
	subCollectors []func(context.Context, chan<- prometheus.Metric)

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

	descTotalClients    *prometheus.Desc
	descCachedEntries   *prometheus.Desc
	descZones           *prometheus.Desc
	descAllowedZones    *prometheus.Desc
	descBlockedZones    *prometheus.Desc
	descAllowListZones  *prometheus.Desc
	descBlockListZones  *prometheus.Desc

	descZoneInfo            *prometheus.Desc
	descZoneExpiryTimestamp *prometheus.Desc
	descZoneDisabled        *prometheus.Desc
	descZoneExpired         *prometheus.Desc
	descZoneSyncFailed      *prometheus.Desc
	descZoneNotifyFailed    *prometheus.Desc
	descZoneInternal        *prometheus.Desc
	descZoneSOASerial       *prometheus.Desc

	descDHCPLeases         *prometheus.Desc
	descDHCPLeasesByType   *prometheus.Desc
	descDHCPScopeEnabled   *prometheus.Desc

	descClusterNodeState               *prometheus.Desc
	descHeartbeatInterval              *prometheus.Desc
	descClusterHeartbeatRetryInterval  *prometheus.Desc
	descClusterConfigRefreshInterval   *prometheus.Desc
	descClusterConfigRetryInterval     *prometheus.Desc
	descClusterConfigLastSynced        *prometheus.Desc

	descCacheMaxEntries           *prometheus.Desc
	descCacheSaveEnabled          *prometheus.Desc
	descCacheServeStaleEnabled    *prometheus.Desc
	descCacheMinRecordTTL         *prometheus.Desc
	descCacheMaxRecordTTL         *prometheus.Desc
	descCacheNegativeRecordTTL    *prometheus.Desc
	descCacheFailureRecordTTL     *prometheus.Desc
	descCachePrefetchEligibility  *prometheus.Desc
	descCachePrefetchTrigger      *prometheus.Desc
	descServeStaleConfig          *prometheus.Desc
	descBlockingEnabled           *prometheus.Desc
	descBlockListUpdateInterval   *prometheus.Desc
	descBlockListNextUpdate       *prometheus.Desc
	descBlockingType              *prometheus.Desc
	descBlockingAnswerTTL         *prometheus.Desc
	descAllowTXTBlockingReport    *prometheus.Desc
	descForwardersCount           *prometheus.Desc
	descForwarderInfo             *prometheus.Desc
	descProtocolEnabled           *prometheus.Desc
	descProtocolPort              *prometheus.Desc
	descDefaultTTL                *prometheus.Desc
	descDNSSECValidationEnabled   *prometheus.Desc
	descIPv6PreferEnabled         *prometheus.Desc
	descIPv6Mode                  *prometheus.Desc
	descRandomizeNameEnabled      *prometheus.Desc
	descQNameMinimizationEnabled  *prometheus.Desc
	descEDNSClientSubnetEnabled   *prometheus.Desc
	descEDNSClientSubnetPrefix    *prometheus.Desc
	descUDPPayloadSize            *prometheus.Desc
	descUDPSocketPoolEnabled      *prometheus.Desc
	descUDPBufferSizeKB           *prometheus.Desc
	descClientTimeout             *prometheus.Desc
	descTCPSendTimeout            *prometheus.Desc
	descTCPReceiveTimeout         *prometheus.Desc
	descListenBacklog             *prometheus.Desc
	descMaxConcurrentResolutions  *prometheus.Desc
	descResolverRetries           *prometheus.Desc
	descResolverTimeout           *prometheus.Desc
	descResolverConcurrency       *prometheus.Desc
	descResolverMaxStackCount     *prometheus.Desc
	descConcurrentForwarding      *prometheus.Desc
	descForwarderRetries          *prometheus.Desc
	descForwarderTimeout          *prometheus.Desc
	descForwarderConcurrency      *prometheus.Desc
	descLogEnabled                *prometheus.Desc
	descLogUseLocalTime           *prometheus.Desc
	descMaxLogFileDays            *prometheus.Desc
	descInMemoryStatsEnabled      *prometheus.Desc
	descMaxStatFileDays           *prometheus.Desc
	descDNSAppsAutoUpdateEnabled  *prometheus.Desc
	descWebServiceHTTPPort        *prometheus.Desc
	descWebServiceTLSEnabled      *prometheus.Desc
	descWebServiceTLSPort         *prometheus.Desc
	descQPMLimitSampleMinutes     *prometheus.Desc
	descQPMLimitUDPTruncationPct  *prometheus.Desc
	descUptimeSeconds             *prometheus.Desc
	descVersionInfo               *prometheus.Desc
}

func New(target config.Target, timeout time.Duration, logger *slog.Logger) *TechnitiumCollector {
	apiClient := client.New(target)
	labels := target.Labels
	labels["instance"] = target.Name

	c := &TechnitiumCollector{
		client:  apiClient,
		target:  target,
		logger:  logger,
		timeout: timeout,

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
		descZoneExpiryTimestamp: prometheus.NewDesc(
			"technitium_dns_zone_expiry_timestamp_seconds",
			"Zone expiry as Unix timestamp.",
			[]string{"zone", "type"}, labels,
		),
		descZoneDisabled: prometheus.NewDesc(
			"technitium_dns_zone_disabled",
			"Whether the zone is disabled.",
			[]string{"zone", "type"}, labels,
		),
		descZoneExpired: prometheus.NewDesc(
			"technitium_dns_zone_expired",
			"Whether the zone has expired.",
			[]string{"zone", "type"}, labels,
		),
		descZoneSyncFailed: prometheus.NewDesc(
			"technitium_dns_zone_sync_failed",
			"Whether zone sync has failed.",
			[]string{"zone", "type"}, labels,
		),
		descZoneNotifyFailed: prometheus.NewDesc(
			"technitium_dns_zone_notify_failed",
			"Whether zone notify has failed.",
			[]string{"zone", "type"}, labels,
		),
		descZoneInternal: prometheus.NewDesc(
			"technitium_dns_zone_internal",
			"Whether the zone is internal.",
			[]string{"zone", "type"}, labels,
		),
		descZoneSOASerial: prometheus.NewDesc(
			"technitium_dns_zone_soa_serial",
			"SOA serial number of the zone.",
			[]string{"zone", "type"}, labels,
		),

		descDHCPLeases: prometheus.NewDesc(
			"technitium_dns_dhcp_leases_count",
			"Number of DHCP leases per scope.",
			[]string{"scope"}, labels,
		),
		descDHCPLeasesByType: prometheus.NewDesc(
			"technitium_dns_dhcp_leases_by_type_count",
			"Number of DHCP leases per scope and type.",
			[]string{"scope", "type"}, labels,
		),
		descDHCPScopeEnabled: prometheus.NewDesc(
			"technitium_dns_dhcp_scope_enabled",
			"Whether a DHCP scope is enabled.",
			[]string{"scope"}, labels,
		),

		descClusterNodeState: prometheus.NewDesc(
			"technitium_dns_cluster_node_state",
			"State of a cluster node.",
			[]string{"cluster_node", "cluster_node_type", "cluster_ip_address"}, labels,
		),
		descHeartbeatInterval: prometheus.NewDesc(
			"technitium_dns_cluster_heartbeat_interval_seconds",
			"Cluster heartbeat refresh interval in seconds.",
			nil, labels,
		),
		descClusterHeartbeatRetryInterval: prometheus.NewDesc(
			"technitium_dns_cluster_heartbeat_retry_interval_seconds",
			"Cluster heartbeat retry interval in seconds.",
			nil, labels,
		),
		descClusterConfigRefreshInterval: prometheus.NewDesc(
			"technitium_dns_cluster_config_refresh_interval_seconds",
			"Cluster config refresh interval in seconds.",
			nil, labels,
		),
		descClusterConfigRetryInterval: prometheus.NewDesc(
			"technitium_dns_cluster_config_retry_interval_seconds",
			"Cluster config retry interval in seconds.",
			nil, labels,
		),
		descClusterConfigLastSynced: prometheus.NewDesc(
			"technitium_dns_cluster_config_last_synced_timestamp_seconds",
			"Cluster config last synced Unix timestamp.",
			nil, labels,
		),

		descCacheMaxEntries: prometheus.NewDesc(
			"technitium_dns_cache_max_entries",
			"Maximum cache entries configured.",
			nil, labels,
		),
		descCacheSaveEnabled: prometheus.NewDesc(
			"technitium_dns_cache_save_enabled",
			"Whether saving cache to disk is enabled.",
			nil, labels,
		),
		descCacheServeStaleEnabled: prometheus.NewDesc(
			"technitium_dns_cache_serve_stale_enabled",
			"Whether serve-stale is enabled.",
			nil, labels,
		),
		descCacheMinRecordTTL: prometheus.NewDesc(
			"technitium_dns_cache_min_record_ttl_seconds",
			"Cache minimum record TTL in seconds.",
			nil, labels,
		),
		descCacheMaxRecordTTL: prometheus.NewDesc(
			"technitium_dns_cache_max_record_ttl_seconds",
			"Cache maximum record TTL in seconds.",
			nil, labels,
		),
		descCacheNegativeRecordTTL: prometheus.NewDesc(
			"technitium_dns_cache_negative_record_ttl_seconds",
			"Cache negative record TTL in seconds.",
			nil, labels,
		),
		descCacheFailureRecordTTL: prometheus.NewDesc(
			"technitium_dns_cache_failure_record_ttl_seconds",
			"Cache failure record TTL in seconds.",
			nil, labels,
		),
		descCachePrefetchEligibility: prometheus.NewDesc(
			"technitium_dns_cache_prefetch_eligibility",
			"Cache prefetch eligibility threshold.",
			nil, labels,
		),
		descCachePrefetchTrigger: prometheus.NewDesc(
			"technitium_dns_cache_prefetch_trigger_seconds",
			"Cache prefetch trigger in seconds.",
			nil, labels,
		),
		descServeStaleConfig: prometheus.NewDesc(
			"technitium_dns_cache_serve_stale_ttl_seconds",
			"Serve-stale TTL configuration in seconds.",
			[]string{"type"}, labels,
		),
		descBlockingEnabled: prometheus.NewDesc(
			"technitium_dns_blocking_enabled",
			"Whether blocking is enabled.",
			nil, labels,
		),
		descBlockListUpdateInterval: prometheus.NewDesc(
			"technitium_dns_blocklist_update_interval_hours",
			"Blocklist update interval in hours.",
			nil, labels,
		),
		descBlockListNextUpdate: prometheus.NewDesc(
			"technitium_dns_blocklist_next_update_timestamp_seconds",
			"Next blocklist update as Unix timestamp.",
			nil, labels,
		),
		descBlockingType: prometheus.NewDesc(
			"technitium_dns_blocking_type",
			"Blocking type mode.",
			[]string{"type"}, labels,
		),
		descBlockingAnswerTTL: prometheus.NewDesc(
			"technitium_dns_blocking_answer_ttl_seconds",
			"Blocking answer TTL in seconds.",
			nil, labels,
		),
		descAllowTXTBlockingReport: prometheus.NewDesc(
			"technitium_dns_allow_txt_blocking_report_enabled",
			"Whether TXT blocking report is allowed.",
			nil, labels,
		),
		descForwardersCount: prometheus.NewDesc(
			"technitium_dns_forwarders_count",
			"Number of configured forwarders.",
			nil, labels,
		),
		descForwarderInfo: prometheus.NewDesc(
			"technitium_dns_forwarder_info",
			"Forwarder address info.",
			[]string{"address", "protocol"}, labels,
		),
		descProtocolEnabled: prometheus.NewDesc(
			"technitium_dns_protocol_enabled",
			"Whether a DNS protocol is enabled.",
			[]string{"protocol"}, labels,
		),
		descProtocolPort: prometheus.NewDesc(
			"technitium_dns_protocol_port",
			"DNS protocol port number.",
			[]string{"protocol"}, labels,
		),
		descDefaultTTL: prometheus.NewDesc(
			"technitium_dns_default_ttl_seconds",
			"Default TTL in seconds.",
			[]string{"type"}, labels,
		),
		descDNSSECValidationEnabled: prometheus.NewDesc(
			"technitium_dns_dnssec_validation_enabled",
			"Whether DNSSEC validation is enabled.",
			nil, labels,
		),
		descIPv6PreferEnabled: prometheus.NewDesc(
			"technitium_dns_ipv6_prefer_enabled",
			"Whether IPv6 is preferred.",
			nil, labels,
		),
		descIPv6Mode: prometheus.NewDesc(
			"technitium_dns_ipv6_mode",
			"IPv6 mode configuration.",
			[]string{"mode"}, labels,
		),
		descRandomizeNameEnabled: prometheus.NewDesc(
			"technitium_dns_randomize_name_enabled",
			"Whether name randomization is enabled.",
			nil, labels,
		),
		descQNameMinimizationEnabled: prometheus.NewDesc(
			"technitium_dns_qname_minimization_enabled",
			"Whether QNAME minimization is enabled.",
			nil, labels,
		),
		descEDNSClientSubnetEnabled: prometheus.NewDesc(
			"technitium_dns_edns_client_subnet_enabled",
			"Whether EDNS Client Subnet is enabled.",
			nil, labels,
		),
		descEDNSClientSubnetPrefix: prometheus.NewDesc(
			"technitium_dns_edns_client_subnet_prefix_length",
			"EDNS Client Subnet prefix length.",
			[]string{"type"}, labels,
		),
		descUDPPayloadSize: prometheus.NewDesc(
			"technitium_dns_udp_payload_size",
			"UDP payload size.",
			nil, labels,
		),
		descUDPSocketPoolEnabled: prometheus.NewDesc(
			"technitium_dns_udp_socket_pool_enabled",
			"Whether UDP socket pool is enabled.",
			nil, labels,
		),
		descUDPBufferSizeKB: prometheus.NewDesc(
			"technitium_dns_udp_buffer_size_kb",
			"UDP buffer size in KB.",
			[]string{"type"}, labels,
		),
		descClientTimeout: prometheus.NewDesc(
			"technitium_dns_client_timeout_seconds",
			"Client timeout in seconds.",
			nil, labels,
		),
		descTCPSendTimeout: prometheus.NewDesc(
			"technitium_dns_tcp_send_timeout_seconds",
			"TCP send timeout in seconds.",
			nil, labels,
		),
		descTCPReceiveTimeout: prometheus.NewDesc(
			"technitium_dns_tcp_receive_timeout_seconds",
			"TCP receive timeout in seconds.",
			nil, labels,
		),
		descListenBacklog: prometheus.NewDesc(
			"technitium_dns_listen_backlog",
			"Listen backlog size.",
			nil, labels,
		),
		descMaxConcurrentResolutions: prometheus.NewDesc(
			"technitium_dns_max_concurrent_resolutions_per_core",
			"Maximum concurrent resolutions per CPU core.",
			nil, labels,
		),
		descResolverRetries: prometheus.NewDesc(
			"technitium_dns_resolver_retries",
			"Resolver retry count.",
			nil, labels,
		),
		descResolverTimeout: prometheus.NewDesc(
			"technitium_dns_resolver_timeout_seconds",
			"Resolver timeout in seconds.",
			nil, labels,
		),
		descResolverConcurrency: prometheus.NewDesc(
			"technitium_dns_resolver_concurrency",
			"Resolver concurrency.",
			nil, labels,
		),
		descResolverMaxStackCount: prometheus.NewDesc(
			"technitium_dns_resolver_max_stack_count",
			"Resolver maximum stack count.",
			nil, labels,
		),
		descConcurrentForwarding: prometheus.NewDesc(
			"technitium_dns_concurrent_forwarding_enabled",
			"Whether concurrent forwarding is enabled.",
			nil, labels,
		),
		descForwarderRetries: prometheus.NewDesc(
			"technitium_dns_forwarder_retries",
			"Forwarder retry count.",
			nil, labels,
		),
		descForwarderTimeout: prometheus.NewDesc(
			"technitium_dns_forwarder_timeout_seconds",
			"Forwarder timeout in seconds.",
			nil, labels,
		),
		descForwarderConcurrency: prometheus.NewDesc(
			"technitium_dns_forwarder_concurrency",
			"Forwarder concurrency.",
			nil, labels,
		),
		descLogEnabled: prometheus.NewDesc(
			"technitium_dns_log_enabled",
			"Whether logging is enabled.",
			[]string{"type"}, labels,
		),
		descLogUseLocalTime: prometheus.NewDesc(
			"technitium_dns_log_use_local_time_enabled",
			"Whether local time is used in logs.",
			nil, labels,
		),
		descMaxLogFileDays: prometheus.NewDesc(
			"technitium_dns_log_retention_days",
			"Maximum log file retention days.",
			nil, labels,
		),
		descInMemoryStatsEnabled: prometheus.NewDesc(
			"technitium_dns_in_memory_stats_enabled",
			"Whether in-memory stats are enabled.",
			nil, labels,
		),
		descMaxStatFileDays: prometheus.NewDesc(
			"technitium_dns_stats_retention_days",
			"Maximum stats file retention days.",
			nil, labels,
		),
		descDNSAppsAutoUpdateEnabled: prometheus.NewDesc(
			"technitium_dns_apps_auto_update_enabled",
			"Whether DNS apps auto update is enabled.",
			nil, labels,
		),
		descWebServiceHTTPPort: prometheus.NewDesc(
			"technitium_dns_web_service_http_port",
			"Web service HTTP port.",
			nil, labels,
		),
		descWebServiceTLSEnabled: prometheus.NewDesc(
			"technitium_dns_web_service_tls_enabled",
			"Whether web service TLS is enabled.",
			nil, labels,
		),
		descWebServiceTLSPort: prometheus.NewDesc(
			"technitium_dns_web_service_tls_port",
			"Web service TLS port.",
			nil, labels,
		),
		descQPMLimitSampleMinutes: prometheus.NewDesc(
			"technitium_dns_qpm_limit_sample_minutes",
			"QPM limit sample window in minutes.",
			nil, labels,
		),
		descQPMLimitUDPTruncationPct: prometheus.NewDesc(
			"technitium_dns_qpm_limit_udp_truncation_percent",
			"QPM limit UDP truncation percentage.",
			nil, labels,
		),
		descUptimeSeconds: prometheus.NewDesc(
			"technitium_dns_uptime_seconds",
			"DNS server uptime in seconds.",
			nil, labels,
		),
		descVersionInfo: prometheus.NewDesc(
			"technitium_dns_version_info",
			"DNS server version.",
			[]string{"version"}, labels,
		),
	}

	c.subCollectors = []func(context.Context, chan<- prometheus.Metric){
		c.collectDashboardStats,
		c.collectZones,
		c.collectSettingsStats,
		c.collectDHCP,
		c.collectCluster,
	}

	c.descs = []*prometheus.Desc{
		c.descScrapeSuccess,
		c.descScrapeDuration,
		c.descQueryTotal,
		c.descQueryNoError,
		c.descQueryServFail,
		c.descQueryNXDomain,
		c.descQueryRefused,
		c.descQueryAuth,
		c.descQueryRecursive,
		c.descQueryCached,
		c.descQueryBlocked,
		c.descQueryDropped,
		c.descTotalClients,
		c.descCachedEntries,
		c.descZones,
		c.descAllowedZones,
		c.descBlockedZones,
		c.descAllowListZones,
		c.descBlockListZones,
		c.descZoneInfo,
		c.descZoneExpiryTimestamp,
		c.descZoneDisabled,
		c.descZoneExpired,
		c.descZoneSyncFailed,
		c.descZoneNotifyFailed,
		c.descZoneInternal,
		c.descZoneSOASerial,
		c.descDHCPLeases,
		c.descDHCPLeasesByType,
		c.descDHCPScopeEnabled,
		c.descClusterNodeState,
		c.descHeartbeatInterval,
		c.descClusterHeartbeatRetryInterval,
		c.descClusterConfigRefreshInterval,
		c.descClusterConfigRetryInterval,
		c.descClusterConfigLastSynced,
		c.descCacheMaxEntries,
		c.descCacheSaveEnabled,
		c.descCacheServeStaleEnabled,
		c.descCacheMinRecordTTL,
		c.descCacheMaxRecordTTL,
		c.descCacheNegativeRecordTTL,
		c.descCacheFailureRecordTTL,
		c.descCachePrefetchEligibility,
		c.descCachePrefetchTrigger,
		c.descServeStaleConfig,
		c.descBlockingEnabled,
		c.descBlockListUpdateInterval,
		c.descBlockListNextUpdate,
		c.descBlockingType,
		c.descBlockingAnswerTTL,
		c.descAllowTXTBlockingReport,
		c.descForwardersCount,
		c.descForwarderInfo,
		c.descProtocolEnabled,
		c.descProtocolPort,
		c.descDefaultTTL,
		c.descDNSSECValidationEnabled,
		c.descIPv6PreferEnabled,
		c.descIPv6Mode,
		c.descRandomizeNameEnabled,
		c.descQNameMinimizationEnabled,
		c.descEDNSClientSubnetEnabled,
		c.descEDNSClientSubnetPrefix,
		c.descUDPPayloadSize,
		c.descUDPSocketPoolEnabled,
		c.descUDPBufferSizeKB,
		c.descClientTimeout,
		c.descTCPSendTimeout,
		c.descTCPReceiveTimeout,
		c.descListenBacklog,
		c.descMaxConcurrentResolutions,
		c.descResolverRetries,
		c.descResolverTimeout,
		c.descResolverConcurrency,
		c.descResolverMaxStackCount,
		c.descConcurrentForwarding,
		c.descForwarderRetries,
		c.descForwarderTimeout,
		c.descForwarderConcurrency,
		c.descLogEnabled,
		c.descLogUseLocalTime,
		c.descMaxLogFileDays,
		c.descInMemoryStatsEnabled,
		c.descMaxStatFileDays,
		c.descDNSAppsAutoUpdateEnabled,
		c.descWebServiceHTTPPort,
		c.descWebServiceTLSEnabled,
		c.descWebServiceTLSPort,
		c.descQPMLimitSampleMinutes,
		c.descQPMLimitUDPTruncationPct,
		c.descUptimeSeconds,
		c.descVersionInfo,
	}

	return c
}

func (c *TechnitiumCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descs {
		ch <- d
	}
}

func (c *TechnitiumCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	start := time.Now()
	scrapeSuccess := float64(1)

	var wg sync.WaitGroup

	for _, collect := range c.subCollectors {
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

func emitGauge(ch chan<- prometheus.Metric, desc *prometheus.Desc, val float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, labels...)
}

func emitCounter(ch chan<- prometheus.Metric, desc *prometheus.Desc, val float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, val, labels...)
}
