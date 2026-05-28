package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectCacheStats(ctx context.Context, ch chan<- prometheus.Metric) {
	settings, err := c.client.GetSettings(ctx)
	if err != nil {
		c.logError("failed to get settings for cache metrics", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_cache_max_entries", "Maximum cache entries configured.",
			nil, c.target.Labels),
		prometheus.GaugeValue, float64(settings.CacheMaximumEntries),
	)

	saveCache := float64(0)
	if settings.SaveCache {
		saveCache = 1
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_cache_save_enabled", "Whether saving cache to disk is enabled.",
			nil, c.target.Labels),
		prometheus.GaugeValue, saveCache,
	)

	serveStale := float64(0)
	if settings.ServeStale {
		serveStale = 1
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_cache_serve_stale_enabled", "Whether serve-stale is enabled.",
			nil, c.target.Labels),
		prometheus.GaugeValue, serveStale,
	)
}
