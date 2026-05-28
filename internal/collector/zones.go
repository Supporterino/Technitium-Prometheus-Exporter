package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectZones(ctx context.Context, ch chan<- prometheus.Metric) {
	zones, err := c.client.GetZones(ctx)
	if err != nil {
		c.logError("failed to get zones", err)
		return
	}

	for _, zone := range zones {
		zoneName := zone.Name
		if zoneName == "" {
			zoneName = "."
		}

		ch <- prometheus.MustNewConstMetric(c.descZoneInfo, prometheus.GaugeValue, 1,
			zoneName, zone.Type, zone.DNSSECStatus,
		)

		if zone.Expiry != "" {
			expiryTime, err := time.Parse(time.RFC3339Nano, zone.Expiry)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("technitium_dns_zone_expiry_timestamp_seconds",
						"Zone expiry as Unix timestamp.",
						[]string{"zone", "type"}, c.target.Labels,
					),
					prometheus.GaugeValue,
					float64(expiryTime.Unix()),
					zoneName, zone.Type,
				)
			}
		}
	}
}
