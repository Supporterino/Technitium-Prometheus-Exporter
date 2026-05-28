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

		emitGauge(ch, c.descZoneInfo, 1,
			zoneName, zone.Type, zone.DNSSECStatus,
		)

		emitGauge(ch, c.descZoneDisabled, boolToFloat(zone.Disabled),
			zoneName, zone.Type,
		)
		emitGauge(ch, c.descZoneExpired, boolToFloat(zone.IsExpired),
			zoneName, zone.Type,
		)
		emitGauge(ch, c.descZoneSyncFailed, boolToFloat(zone.SyncFailed),
			zoneName, zone.Type,
		)
		emitGauge(ch, c.descZoneNotifyFailed, boolToFloat(zone.NotifyFailed),
			zoneName, zone.Type,
		)
		emitGauge(ch, c.descZoneInternal, boolToFloat(zone.Internal),
			zoneName, zone.Type,
		)
		emitGauge(ch, c.descZoneSOASerial, float64(zone.SOASerial),
			zoneName, zone.Type,
		)

		if zone.Expiry != "" {
			if expiryTime, err := time.Parse(time.RFC3339Nano, zone.Expiry); err == nil {
				emitGauge(ch, c.descZoneExpiryTimestamp, float64(expiryTime.Unix()),
					zoneName, zone.Type,
				)
			}
		}
	}
}
