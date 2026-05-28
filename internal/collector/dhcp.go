package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectDHCP(ctx context.Context, ch chan<- prometheus.Metric) {
	if !c.target.Features.DHCP {
		return
	}

	scopes, err := c.client.GetDHCPScopes(ctx)
	if err != nil {
		c.logError("failed to get DHCP scopes", err)
		return
	}

	for _, scope := range scopes {
		enabled := float64(0)
		if scope.Enabled {
			enabled = 1
		}
		ch <- prometheus.MustNewConstMetric(c.descDHCPScopeEnabled, prometheus.GaugeValue, enabled, scope.Name)
	}

	leases, err := c.client.GetDHCPLeases(ctx)
	if err != nil {
		c.logError("failed to get DHCP leases", err)
		return
	}

	leaseCounts := make(map[string]int)
	for _, lease := range leases {
		leaseCounts[lease.Scope]++
	}

	for scope, count := range leaseCounts {
		ch <- prometheus.MustNewConstMetric(c.descDHCPLeases, prometheus.GaugeValue, float64(count), scope)
	}
}
