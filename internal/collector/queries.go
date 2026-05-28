package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectDashboardStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetDashboardStats(ctx)
	if err != nil {
		c.logError("failed to get dashboard stats", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.descQueryTotal, prometheus.CounterValue, float64(stats.Stats.TotalQueries))
	ch <- prometheus.MustNewConstMetric(c.descQueryNoError, prometheus.CounterValue, float64(stats.Stats.TotalNoError))
	ch <- prometheus.MustNewConstMetric(c.descQueryServFail, prometheus.CounterValue, float64(stats.Stats.TotalServerFailure))
	ch <- prometheus.MustNewConstMetric(c.descQueryNXDomain, prometheus.CounterValue, float64(stats.Stats.TotalNxDomain))
	ch <- prometheus.MustNewConstMetric(c.descQueryRefused, prometheus.CounterValue, float64(stats.Stats.TotalRefused))
	ch <- prometheus.MustNewConstMetric(c.descQueryAuth, prometheus.CounterValue, float64(stats.Stats.TotalAuthoritative))
	ch <- prometheus.MustNewConstMetric(c.descQueryRecursive, prometheus.CounterValue, float64(stats.Stats.TotalRecursive))
	ch <- prometheus.MustNewConstMetric(c.descQueryCached, prometheus.CounterValue, float64(stats.Stats.TotalCached))
	ch <- prometheus.MustNewConstMetric(c.descQueryBlocked, prometheus.CounterValue, float64(stats.Stats.TotalBlocked))
	ch <- prometheus.MustNewConstMetric(c.descQueryDropped, prometheus.CounterValue, float64(stats.Stats.TotalDropped))

	ch <- prometheus.MustNewConstMetric(c.descTotalClients, prometheus.GaugeValue, float64(stats.Stats.TotalClients))
	ch <- prometheus.MustNewConstMetric(c.descCachedEntries, prometheus.GaugeValue, float64(stats.Stats.CachedEntries))
	ch <- prometheus.MustNewConstMetric(c.descZones, prometheus.GaugeValue, float64(stats.Stats.Zones))
	ch <- prometheus.MustNewConstMetric(c.descAllowedZones, prometheus.GaugeValue, float64(stats.Stats.AllowedZones))
	ch <- prometheus.MustNewConstMetric(c.descBlockedZones, prometheus.GaugeValue, float64(stats.Stats.BlockedZones))
	ch <- prometheus.MustNewConstMetric(c.descAllowListZones, prometheus.GaugeValue, float64(stats.Stats.AllowListZones))
	ch <- prometheus.MustNewConstMetric(c.descBlockListZones, prometheus.GaugeValue, float64(stats.Stats.BlockListZones))
}
