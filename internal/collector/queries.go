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

	emitCounter(ch, c.descQueryTotal, float64(stats.Stats.TotalQueries))
	emitCounter(ch, c.descQueryNoError, float64(stats.Stats.TotalNoError))
	emitCounter(ch, c.descQueryServFail, float64(stats.Stats.TotalServerFailure))
	emitCounter(ch, c.descQueryNXDomain, float64(stats.Stats.TotalNxDomain))
	emitCounter(ch, c.descQueryRefused, float64(stats.Stats.TotalRefused))
	emitCounter(ch, c.descQueryAuth, float64(stats.Stats.TotalAuthoritative))
	emitCounter(ch, c.descQueryRecursive, float64(stats.Stats.TotalRecursive))
	emitCounter(ch, c.descQueryCached, float64(stats.Stats.TotalCached))
	emitCounter(ch, c.descQueryBlocked, float64(stats.Stats.TotalBlocked))
	emitCounter(ch, c.descQueryDropped, float64(stats.Stats.TotalDropped))

	emitGauge(ch, c.descTotalClients, float64(stats.Stats.TotalClients))
	emitGauge(ch, c.descCachedEntries, float64(stats.Stats.CachedEntries))
	emitGauge(ch, c.descZones, float64(stats.Stats.Zones))
	emitGauge(ch, c.descAllowedZones, float64(stats.Stats.AllowedZones))
	emitGauge(ch, c.descBlockedZones, float64(stats.Stats.BlockedZones))
	emitGauge(ch, c.descAllowListZones, float64(stats.Stats.AllowListZones))
	emitGauge(ch, c.descBlockListZones, float64(stats.Stats.BlockListZones))
}
