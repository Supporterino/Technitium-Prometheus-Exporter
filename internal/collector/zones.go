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

		ch <- prometheus.MustNewConstMetric(c.descZoneDisabled, prometheus.GaugeValue, boolToFloat(zone.Disabled),
			zoneName, zone.Type,
		)
		ch <- prometheus.MustNewConstMetric(c.descZoneExpired, prometheus.GaugeValue, boolToFloat(zone.IsExpired),
			zoneName, zone.Type,
		)
		ch <- prometheus.MustNewConstMetric(c.descZoneSyncFailed, prometheus.GaugeValue, boolToFloat(zone.SyncFailed),
			zoneName, zone.Type,
		)
		ch <- prometheus.MustNewConstMetric(c.descZoneNotifyFailed, prometheus.GaugeValue, boolToFloat(zone.NotifyFailed),
			zoneName, zone.Type,
		)
		ch <- prometheus.MustNewConstMetric(c.descZoneInternal, prometheus.GaugeValue, boolToFloat(zone.Internal),
			zoneName, zone.Type,
		)
		ch <- prometheus.MustNewConstMetric(c.descZoneSOASerial, prometheus.GaugeValue, float64(zone.SOASerial),
			zoneName, zone.Type,
		)

		if zone.Expiry != "" {
			if expiryTime, err := time.Parse(time.RFC3339Nano, zone.Expiry); err == nil {
				ch <- prometheus.MustNewConstMetric(c.descZoneExpiryTimestamp, prometheus.GaugeValue,
					float64(expiryTime.Unix()),
					zoneName, zone.Type,
				)
			}
		}
	}
}
