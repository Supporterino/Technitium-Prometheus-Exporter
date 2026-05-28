package collector

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (c *TechnitiumCollector) collectSettingsStats(ctx context.Context, ch chan<- prometheus.Metric) {
	settings, err := c.client.GetSettings(ctx)
	if err != nil {
		c.logError("failed to get settings", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.descCacheMaxEntries, prometheus.GaugeValue, float64(settings.CacheMaximumEntries))
	ch <- prometheus.MustNewConstMetric(c.descCacheSaveEnabled, prometheus.GaugeValue, boolToFloat(settings.SaveCache))
	ch <- prometheus.MustNewConstMetric(c.descCacheServeStaleEnabled, prometheus.GaugeValue, boolToFloat(settings.ServeStale))

	ch <- prometheus.MustNewConstMetric(c.descCacheMinRecordTTL, prometheus.GaugeValue, float64(settings.CacheMinimumRecordTTL))
	ch <- prometheus.MustNewConstMetric(c.descCacheMaxRecordTTL, prometheus.GaugeValue, float64(settings.CacheMaximumRecordTTL))
	ch <- prometheus.MustNewConstMetric(c.descCacheNegativeRecordTTL, prometheus.GaugeValue, float64(settings.CacheNegativeRecordTTL))
	ch <- prometheus.MustNewConstMetric(c.descCacheFailureRecordTTL, prometheus.GaugeValue, float64(settings.CacheFailureRecordTTL))
	ch <- prometheus.MustNewConstMetric(c.descCachePrefetchEligibility, prometheus.GaugeValue, float64(settings.CachePrefetchEligibility))
	ch <- prometheus.MustNewConstMetric(c.descCachePrefetchTrigger, prometheus.GaugeValue, float64(settings.CachePrefetchTrigger))

	ch <- prometheus.MustNewConstMetric(c.descServeStaleConfig, prometheus.GaugeValue, float64(settings.ServeStaleTTL), "stale_ttl")
	ch <- prometheus.MustNewConstMetric(c.descServeStaleConfig, prometheus.GaugeValue, float64(settings.ServeStaleAnswerTTL), "answer_ttl")
	ch <- prometheus.MustNewConstMetric(c.descServeStaleConfig, prometheus.GaugeValue, float64(settings.ServeStaleResetTTL), "reset_ttl")
	ch <- prometheus.MustNewConstMetric(c.descServeStaleConfig, prometheus.GaugeValue, float64(settings.ServeStaleMaxWaitTime), "max_wait")

	ch <- prometheus.MustNewConstMetric(c.descBlockingEnabled, prometheus.GaugeValue, boolToFloat(settings.EnableBlocking))
	ch <- prometheus.MustNewConstMetric(c.descBlockListUpdateInterval, prometheus.GaugeValue, float64(settings.BlockListUpdateIntervalHours))
	ch <- prometheus.MustNewConstMetric(c.descBlockingAnswerTTL, prometheus.GaugeValue, float64(settings.BlockingAnswerTTL))
	ch <- prometheus.MustNewConstMetric(c.descAllowTXTBlockingReport, prometheus.GaugeValue, boolToFloat(settings.AllowTXTBlockingReport))

	if settings.BlockingType != "" {
		blockingTypeValue := 0
		switch strings.ToLower(settings.BlockingType) {
		case "anyaddress":
			blockingTypeValue = 1
		case "nxdomain":
			blockingTypeValue = 2
		case "customaddress":
			blockingTypeValue = 3
		}
		ch <- prometheus.MustNewConstMetric(c.descBlockingType, prometheus.GaugeValue, float64(blockingTypeValue), settings.BlockingType)
	}

	if settings.BlockListNextUpdatedOn != "" {
		if nextUpdate, err := time.Parse(time.RFC3339Nano, settings.BlockListNextUpdatedOn); err == nil {
			ch <- prometheus.MustNewConstMetric(c.descBlockListNextUpdate, prometheus.GaugeValue, float64(nextUpdate.Unix()))
		}
	}

	ch <- prometheus.MustNewConstMetric(c.descForwardersCount, prometheus.GaugeValue, float64(len(settings.Forwarders)))
	for _, fwd := range settings.Forwarders {
		ch <- prometheus.MustNewConstMetric(c.descForwarderInfo, prometheus.GaugeValue, 1, fwd, settings.ForwarderProtocol)
	}

	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverUDPEnabled), "udp")
	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverTCPEnabled), "tcp")
	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverTLSEnabled), "tls")
	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverHTTPSEnabled), "https")
	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverHTTPEnabled), "http")
	ch <- prometheus.MustNewConstMetric(c.descProtocolEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSOverQUICEnabled), "quic")

	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverUDPProxyPort), "udp")
	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverTCPProxyPort), "tcp")
	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverTLSPort), "tls")
	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverHTTPSPort), "https")
	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverHTTPPort), "http")
	ch <- prometheus.MustNewConstMetric(c.descProtocolPort, prometheus.GaugeValue, float64(settings.DNSOverQUICPort), "quic")

	ch <- prometheus.MustNewConstMetric(c.descDefaultTTL, prometheus.GaugeValue, float64(settings.DefaultRecordTTL), "record")
	ch <- prometheus.MustNewConstMetric(c.descDefaultTTL, prometheus.GaugeValue, float64(settings.DefaultNsRecordTTL), "ns")
	ch <- prometheus.MustNewConstMetric(c.descDefaultTTL, prometheus.GaugeValue, float64(settings.DefaultSoaRecordTTL), "soa")

	ch <- prometheus.MustNewConstMetric(c.descDNSSECValidationEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSSECValidation))
	ch <- prometheus.MustNewConstMetric(c.descIPv6PreferEnabled, prometheus.GaugeValue, boolToFloat(settings.PreferIPv6))

	if settings.IPv6Mode != "" {
		ipv6ModeValue := 0
		switch strings.ToLower(settings.IPv6Mode) {
		case "enabled":
			ipv6ModeValue = 1
		case "disabled":
			ipv6ModeValue = 2
		}
		ch <- prometheus.MustNewConstMetric(c.descIPv6Mode, prometheus.GaugeValue, float64(ipv6ModeValue), settings.IPv6Mode)
	}

	ch <- prometheus.MustNewConstMetric(c.descRandomizeNameEnabled, prometheus.GaugeValue, boolToFloat(settings.RandomizeName))
	ch <- prometheus.MustNewConstMetric(c.descQNameMinimizationEnabled, prometheus.GaugeValue, boolToFloat(settings.QNameMinimization))
	ch <- prometheus.MustNewConstMetric(c.descEDNSClientSubnetEnabled, prometheus.GaugeValue, boolToFloat(settings.EDNSClientSubnet))
	ch <- prometheus.MustNewConstMetric(c.descEDNSClientSubnetPrefix, prometheus.GaugeValue, float64(settings.EDNSClientSubnetIPv4Prefix), "ipv4")
	ch <- prometheus.MustNewConstMetric(c.descEDNSClientSubnetPrefix, prometheus.GaugeValue, float64(settings.EDNSClientSubnetIPv6Prefix), "ipv6")

	ch <- prometheus.MustNewConstMetric(c.descUDPPayloadSize, prometheus.GaugeValue, float64(settings.UDPPayloadSize))
	ch <- prometheus.MustNewConstMetric(c.descUDPSocketPoolEnabled, prometheus.GaugeValue, boolToFloat(settings.EnableUDPSocketPool))
	ch <- prometheus.MustNewConstMetric(c.descUDPBufferSizeKB, prometheus.GaugeValue, float64(settings.UDPSendBufferSizeKB), "send")
	ch <- prometheus.MustNewConstMetric(c.descUDPBufferSizeKB, prometheus.GaugeValue, float64(settings.UDPReceiveBufferSizeKB), "recv")

	ch <- prometheus.MustNewConstMetric(c.descClientTimeout, prometheus.GaugeValue, float64(settings.ClientTimeout)/1000)
	ch <- prometheus.MustNewConstMetric(c.descTCPSendTimeout, prometheus.GaugeValue, float64(settings.TCPSendTimeout))
	ch <- prometheus.MustNewConstMetric(c.descTCPReceiveTimeout, prometheus.GaugeValue, float64(settings.TCPReceiveTimeout))
	ch <- prometheus.MustNewConstMetric(c.descListenBacklog, prometheus.GaugeValue, float64(settings.ListenBacklog))
	ch <- prometheus.MustNewConstMetric(c.descMaxConcurrentResolutions, prometheus.GaugeValue, float64(settings.MaxConcurrentResolutions))

	ch <- prometheus.MustNewConstMetric(c.descResolverRetries, prometheus.GaugeValue, float64(settings.ResolverRetries))
	ch <- prometheus.MustNewConstMetric(c.descResolverTimeout, prometheus.GaugeValue, float64(settings.ResolverTimeout))
	ch <- prometheus.MustNewConstMetric(c.descResolverConcurrency, prometheus.GaugeValue, float64(settings.ResolverConcurrency))
	ch <- prometheus.MustNewConstMetric(c.descResolverMaxStackCount, prometheus.GaugeValue, float64(settings.ResolverMaxStackCount))

	ch <- prometheus.MustNewConstMetric(c.descConcurrentForwarding, prometheus.GaugeValue, boolToFloat(settings.ConcurrentForwarding))
	ch <- prometheus.MustNewConstMetric(c.descForwarderRetries, prometheus.GaugeValue, float64(settings.ForwarderRetries))
	ch <- prometheus.MustNewConstMetric(c.descForwarderTimeout, prometheus.GaugeValue, float64(settings.ForwarderTimeout))
	ch <- prometheus.MustNewConstMetric(c.descForwarderConcurrency, prometheus.GaugeValue, float64(settings.ForwarderConcurrency))

	ch <- prometheus.MustNewConstMetric(c.descLogEnabled, prometheus.GaugeValue, boolToFloat(settings.EnableLogging), "error")
	ch <- prometheus.MustNewConstMetric(c.descLogEnabled, prometheus.GaugeValue, boolToFloat(settings.LogQueries), "query")
	ch <- prometheus.MustNewConstMetric(c.descLogUseLocalTime, prometheus.GaugeValue, boolToFloat(settings.UseLocalTime))
	ch <- prometheus.MustNewConstMetric(c.descMaxLogFileDays, prometheus.GaugeValue, float64(settings.MaxLogFileDays))
	ch <- prometheus.MustNewConstMetric(c.descInMemoryStatsEnabled, prometheus.GaugeValue, boolToFloat(settings.EnableInMemoryStats))
	ch <- prometheus.MustNewConstMetric(c.descMaxStatFileDays, prometheus.GaugeValue, float64(settings.MaxStatFileDays))
	ch <- prometheus.MustNewConstMetric(c.descDNSAppsAutoUpdateEnabled, prometheus.GaugeValue, boolToFloat(settings.DNSAppsAutoUpdate))

	ch <- prometheus.MustNewConstMetric(c.descWebServiceHTTPPort, prometheus.GaugeValue, float64(settings.WebServiceHTTPPort))
	ch <- prometheus.MustNewConstMetric(c.descWebServiceTLSEnabled, prometheus.GaugeValue, boolToFloat(settings.WebServiceTLSEnabled))
	ch <- prometheus.MustNewConstMetric(c.descWebServiceTLSPort, prometheus.GaugeValue, float64(settings.WebServiceTLSPort))

	ch <- prometheus.MustNewConstMetric(c.descQPMLimitSampleMinutes, prometheus.GaugeValue, float64(settings.QPMLimitSampleMinutes))
	ch <- prometheus.MustNewConstMetric(c.descQPMLimitUDPTruncationPct, prometheus.GaugeValue, float64(settings.QPMLimitUDPTruncationPct))

	ch <- prometheus.MustNewConstMetric(c.descVersionInfo, prometheus.GaugeValue, 1, settings.Version)
	if settings.Uptimestamp != "" {
		if uptime, err := time.Parse(time.RFC3339Nano, settings.Uptimestamp); err == nil {
			ch <- prometheus.MustNewConstMetric(c.descUptimeSeconds, prometheus.GaugeValue, time.Since(uptime).Seconds())
		}
	}
}