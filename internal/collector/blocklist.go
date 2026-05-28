package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectBlocklistStats(ctx context.Context, ch chan<- prometheus.Metric) {
	settings, err := c.client.GetSettings(ctx)
	if err != nil {
		c.logError("failed to get settings for blocklist metrics", err)
		return
	}

	blockingEnabled := float64(0)
	if settings.EnableBlocking {
		blockingEnabled = 1
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_blocking_enabled", "Whether blocking is enabled.",
			nil, c.target.Labels),
		prometheus.GaugeValue, blockingEnabled,
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_blocklist_update_interval_hours", "Blocklist update interval in hours.",
			nil, c.target.Labels),
		prometheus.GaugeValue, float64(settings.BlockListUpdateIntervalHours),
	)

	if settings.BlockListNextUpdatedOn != "" {
		nextUpdate, err := time.Parse(time.RFC3339Nano, settings.BlockListNextUpdatedOn)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("technitium_dns_blocklist_next_update_timestamp_seconds",
					"Next blocklist update as Unix timestamp.",
					nil, c.target.Labels),
				prometheus.GaugeValue, float64(nextUpdate.Unix()),
			)
		}
	}
}
