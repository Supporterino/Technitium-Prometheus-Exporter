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

	emitGauge(ch, c.descCacheMaxEntries, float64(settings.CacheMaximumEntries))
	emitGauge(ch, c.descCacheSaveEnabled, boolToFloat(settings.SaveCache))
	emitGauge(ch, c.descCacheServeStaleEnabled, boolToFloat(settings.ServeStale))

	emitGauge(ch, c.descCacheMinRecordTTL, float64(settings.CacheMinimumRecordTTL))
	emitGauge(ch, c.descCacheMaxRecordTTL, float64(settings.CacheMaximumRecordTTL))
	emitGauge(ch, c.descCacheNegativeRecordTTL, float64(settings.CacheNegativeRecordTTL))
	emitGauge(ch, c.descCacheFailureRecordTTL, float64(settings.CacheFailureRecordTTL))
	emitGauge(ch, c.descCachePrefetchEligibility, float64(settings.CachePrefetchEligibility))
	emitGauge(ch, c.descCachePrefetchTrigger, float64(settings.CachePrefetchTrigger))

	emitGauge(ch, c.descServeStaleConfig, float64(settings.ServeStaleTTL), "stale_ttl")
	emitGauge(ch, c.descServeStaleConfig, float64(settings.ServeStaleAnswerTTL), "answer_ttl")
	emitGauge(ch, c.descServeStaleConfig, float64(settings.ServeStaleResetTTL), "reset_ttl")
	emitGauge(ch, c.descServeStaleConfig, float64(settings.ServeStaleMaxWaitTime), "max_wait")

	emitGauge(ch, c.descBlockingEnabled, boolToFloat(settings.EnableBlocking))
	emitGauge(ch, c.descBlockListUpdateInterval, float64(settings.BlockListUpdateIntervalHours))
	emitGauge(ch, c.descBlockingAnswerTTL, float64(settings.BlockingAnswerTTL))
	emitGauge(ch, c.descAllowTXTBlockingReport, boolToFloat(settings.AllowTXTBlockingReport))

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
		emitGauge(ch, c.descBlockingType, float64(blockingTypeValue), settings.BlockingType)
	}

	if settings.BlockListNextUpdatedOn != "" {
		if nextUpdate, err := time.Parse(time.RFC3339Nano, settings.BlockListNextUpdatedOn); err == nil {
			emitGauge(ch, c.descBlockListNextUpdate, float64(nextUpdate.Unix()))
		}
	}

	emitGauge(ch, c.descForwardersCount, float64(len(settings.Forwarders)))
	for _, fwd := range settings.Forwarders {
		emitGauge(ch, c.descForwarderInfo, 1, fwd, settings.ForwarderProtocol)
	}

	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverUDPEnabled), "udp")
	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverTCPEnabled), "tcp")
	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverTLSEnabled), "tls")
	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverHTTPSEnabled), "https")
	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverHTTPEnabled), "http")
	emitGauge(ch, c.descProtocolEnabled, boolToFloat(settings.DNSOverQUICEnabled), "quic")

	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverUDPProxyPort), "udp")
	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverTCPProxyPort), "tcp")
	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverTLSPort), "tls")
	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverHTTPSPort), "https")
	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverHTTPPort), "http")
	emitGauge(ch, c.descProtocolPort, float64(settings.DNSOverQUICPort), "quic")

	emitGauge(ch, c.descDefaultTTL, float64(settings.DefaultRecordTTL), "record")
	emitGauge(ch, c.descDefaultTTL, float64(settings.DefaultNsRecordTTL), "ns")
	emitGauge(ch, c.descDefaultTTL, float64(settings.DefaultSoaRecordTTL), "soa")

	emitGauge(ch, c.descDNSSECValidationEnabled, boolToFloat(settings.DNSSECValidation))
	emitGauge(ch, c.descIPv6PreferEnabled, boolToFloat(settings.PreferIPv6))

	if settings.IPv6Mode != "" {
		ipv6ModeValue := 0
		switch strings.ToLower(settings.IPv6Mode) {
		case "enabled":
			ipv6ModeValue = 1
		case "disabled":
			ipv6ModeValue = 2
		}
		emitGauge(ch, c.descIPv6Mode, float64(ipv6ModeValue), settings.IPv6Mode)
	}

	emitGauge(ch, c.descRandomizeNameEnabled, boolToFloat(settings.RandomizeName))
	emitGauge(ch, c.descQNameMinimizationEnabled, boolToFloat(settings.QNameMinimization))
	emitGauge(ch, c.descEDNSClientSubnetEnabled, boolToFloat(settings.EDNSClientSubnet))
	emitGauge(ch, c.descEDNSClientSubnetPrefix, float64(settings.EDNSClientSubnetIPv4Prefix), "ipv4")
	emitGauge(ch, c.descEDNSClientSubnetPrefix, float64(settings.EDNSClientSubnetIPv6Prefix), "ipv6")

	emitGauge(ch, c.descUDPPayloadSize, float64(settings.UDPPayloadSize))
	emitGauge(ch, c.descUDPSocketPoolEnabled, boolToFloat(settings.EnableUDPSocketPool))
	emitGauge(ch, c.descUDPBufferSizeKB, float64(settings.UDPSendBufferSizeKB), "send")
	emitGauge(ch, c.descUDPBufferSizeKB, float64(settings.UDPReceiveBufferSizeKB), "recv")

	emitGauge(ch, c.descClientTimeout, float64(settings.ClientTimeout)/1000)
	emitGauge(ch, c.descTCPSendTimeout, float64(settings.TCPSendTimeout))
	emitGauge(ch, c.descTCPReceiveTimeout, float64(settings.TCPReceiveTimeout))
	emitGauge(ch, c.descListenBacklog, float64(settings.ListenBacklog))
	emitGauge(ch, c.descMaxConcurrentResolutions, float64(settings.MaxConcurrentResolutions))

	emitGauge(ch, c.descResolverRetries, float64(settings.ResolverRetries))
	emitGauge(ch, c.descResolverTimeout, float64(settings.ResolverTimeout))
	emitGauge(ch, c.descResolverConcurrency, float64(settings.ResolverConcurrency))
	emitGauge(ch, c.descResolverMaxStackCount, float64(settings.ResolverMaxStackCount))

	emitGauge(ch, c.descConcurrentForwarding, boolToFloat(settings.ConcurrentForwarding))
	emitGauge(ch, c.descForwarderRetries, float64(settings.ForwarderRetries))
	emitGauge(ch, c.descForwarderTimeout, float64(settings.ForwarderTimeout))
	emitGauge(ch, c.descForwarderConcurrency, float64(settings.ForwarderConcurrency))

	emitGauge(ch, c.descLogEnabled, boolToFloat(settings.EnableLogging), "error")
	emitGauge(ch, c.descLogEnabled, boolToFloat(settings.LogQueries), "query")
	emitGauge(ch, c.descLogUseLocalTime, boolToFloat(settings.UseLocalTime))
	emitGauge(ch, c.descMaxLogFileDays, float64(settings.MaxLogFileDays))
	emitGauge(ch, c.descInMemoryStatsEnabled, boolToFloat(settings.EnableInMemoryStats))
	emitGauge(ch, c.descMaxStatFileDays, float64(settings.MaxStatFileDays))
	emitGauge(ch, c.descDNSAppsAutoUpdateEnabled, boolToFloat(settings.DNSAppsAutoUpdate))

	emitGauge(ch, c.descWebServiceHTTPPort, float64(settings.WebServiceHTTPPort))
	emitGauge(ch, c.descWebServiceTLSEnabled, boolToFloat(settings.WebServiceTLSEnabled))
	emitGauge(ch, c.descWebServiceTLSPort, float64(settings.WebServiceTLSPort))

	emitGauge(ch, c.descQPMLimitSampleMinutes, float64(settings.QPMLimitSampleMinutes))
	emitGauge(ch, c.descQPMLimitUDPTruncationPct, float64(settings.QPMLimitUDPTruncationPct))

	emitGauge(ch, c.descVersionInfo, 1, settings.Version)
	if settings.Uptimestamp != "" {
		if uptime, err := time.Parse(time.RFC3339Nano, settings.Uptimestamp); err == nil {
			emitGauge(ch, c.descUptimeSeconds, time.Since(uptime).Seconds())
		}
	}
}
