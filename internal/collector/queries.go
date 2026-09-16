package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"technitium-dns-exporter/internal/client"
)

// queryCounterKeys lists the counter metrics derived from the dashboard stats
// response. They are accumulated across scrapes into true Prometheus counters,
// see updateQueryCounters.
var queryCounterKeys = []string{
	"total",
	"noerror",
	"servfail",
	"nxdomain",
	"refused",
	"authoritative",
	"recursive",
	"cached",
	"blocked",
	"dropped",
}

// collectDashboardStats emits the query statistics.
//
// The Technitium dashboard API reports totals for a rolling window
// (type=LastHour), so the raw values go up and down. To expose usable
// Prometheus counters the exporter accumulates them: the exact total is taken
// from the per-minute buckets in mainChartData, while the response-type and
// RCODE breakdowns (only available as windowed scalars) are attributed the
// proportional share of that exact delta.
func (c *TechnitiumCollector) collectDashboardStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	stats, err := c.client.GetDashboardStats(ctx)
	if err != nil {
		return err
	}

	c.updateQueryCounters(stats)

	c.counterMu.Lock()
	values := make(map[string]float64, len(queryCounterKeys))
	for _, key := range queryCounterKeys {
		values[key] = c.cumCounters[key]
	}
	c.counterMu.Unlock()

	emitCounter(ch, c.descQueryTotal, values["total"])
	emitCounter(ch, c.descQueryNoError, values["noerror"])
	emitCounter(ch, c.descQueryServFail, values["servfail"])
	emitCounter(ch, c.descQueryNXDomain, values["nxdomain"])
	emitCounter(ch, c.descQueryRefused, values["refused"])
	emitCounter(ch, c.descQueryAuth, values["authoritative"])
	emitCounter(ch, c.descQueryRecursive, values["recursive"])
	emitCounter(ch, c.descQueryCached, values["cached"])
	emitCounter(ch, c.descQueryBlocked, values["blocked"])
	emitCounter(ch, c.descQueryDropped, values["dropped"])

	emitGauge(ch, c.descTotalClients, float64(stats.Stats.TotalClients))
	emitGauge(ch, c.descCachedEntries, float64(stats.Stats.CachedEntries))
	emitGauge(ch, c.descZones, float64(stats.Stats.Zones))
	emitGauge(ch, c.descAllowedZones, float64(stats.Stats.AllowedZones))
	emitGauge(ch, c.descBlockedZones, float64(stats.Stats.BlockedZones))
	emitGauge(ch, c.descAllowListZones, float64(stats.Stats.AllowListZones))
	emitGauge(ch, c.descBlockListZones, float64(stats.Stats.BlockListZones))

	return nil
}

// updateQueryCounters advances the cumulative counter state for one scrape.
func (c *TechnitiumCollector) updateQueryCounters(d *client.DashboardStats) {
	stats := d.Stats
	windowTotal := float64(stats.TotalQueries)

	c.counterMu.Lock()
	defer c.counterMu.Unlock()

	// First scrape after (re)start: seed with the current window total so the
	// counters start at a sensible value instead of zero.
	if !c.countersSeeded {
		c.cumCounters["total"] = windowTotal
		c.cumCounters["noerror"] = float64(stats.TotalNoError)
		c.cumCounters["servfail"] = float64(stats.TotalServerFailure)
		c.cumCounters["nxdomain"] = float64(stats.TotalNxDomain)
		c.cumCounters["refused"] = float64(stats.TotalRefused)
		c.cumCounters["authoritative"] = float64(stats.TotalAuthoritative)
		c.cumCounters["recursive"] = float64(stats.TotalRecursive)
		c.cumCounters["cached"] = float64(stats.TotalCached)
		c.cumCounters["blocked"] = float64(stats.TotalBlocked)
		c.cumCounters["dropped"] = float64(stats.TotalDropped)

		c.lastBucket = latestBucket(d.MainChartData)
		c.countersSeeded = true
		return
	}

	delta := 0.0
	if latest := latestBucket(d.MainChartData); !latest.IsZero() && latest.After(c.lastBucket) {
		delta = sumNewBuckets(d.MainChartData, c.lastBucket)
		c.lastBucket = latest
	}
	if delta < 0 {
		delta = 0
	}

	c.cumCounters["total"] += delta

	// The breakdown metrics are only exposed as windowed scalars, so attribute
	// them the proportional share of the exact total delta.
	share := func(v int64) float64 {
		if windowTotal <= 0 {
			return 0
		}
		return delta * float64(v) / windowTotal
	}
	c.cumCounters["noerror"] += share(stats.TotalNoError)
	c.cumCounters["servfail"] += share(stats.TotalServerFailure)
	c.cumCounters["nxdomain"] += share(stats.TotalNxDomain)
	c.cumCounters["refused"] += share(stats.TotalRefused)
	c.cumCounters["authoritative"] += share(stats.TotalAuthoritative)
	c.cumCounters["recursive"] += share(stats.TotalRecursive)
	c.cumCounters["cached"] += share(stats.TotalCached)
	c.cumCounters["blocked"] += share(stats.TotalBlocked)
	c.cumCounters["dropped"] += share(stats.TotalDropped)
}

// latestBucket returns the newest timestamp in the chart data, or the zero time
// when the data is empty or its labels cannot be parsed.
func latestBucket(data client.ChartData) time.Time {
	var latest time.Time
	for _, label := range data.Labels {
		ts, err := time.Parse(time.RFC3339Nano, label)
		if err != nil {
			continue
		}
		if ts.After(latest) {
			latest = ts
		}
	}
	return latest
}

// sumNewBuckets sums the "Total" dataset for all buckets newer than after.
func sumNewBuckets(data client.ChartData, after time.Time) float64 {
	if len(data.Datasets) == 0 {
		return 0
	}

	dataset := data.Datasets[0]
	for _, ds := range data.Datasets {
		if ds.Label == "Total" {
			dataset = ds
			break
		}
	}

	var sum float64
	for i, label := range data.Labels {
		if i >= len(dataset.Data) {
			break
		}
		ts, err := time.Parse(time.RFC3339Nano, label)
		if err != nil {
			continue
		}
		if ts.After(after) {
			sum += float64(dataset.Data[i])
		}
	}
	return sum
}
