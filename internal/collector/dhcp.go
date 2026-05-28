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
		emitGauge(ch, c.descDHCPScopeEnabled, boolToFloat(scope.Enabled), scope.Name)
	}

	leases, err := c.client.GetDHCPLeases(ctx)
	if err != nil {
		c.logError("failed to get DHCP leases", err)
		return
	}

	leaseCounts := make(map[string]int)
	leaseByType := make(map[string]map[string]int)
	for _, lease := range leases {
		leaseCounts[lease.Scope]++
		if leaseByType[lease.Scope] == nil {
			leaseByType[lease.Scope] = make(map[string]int)
		}
		leaseByType[lease.Scope][lease.Type]++
	}

	for scope, count := range leaseCounts {
		emitGauge(ch, c.descDHCPLeases, float64(count), scope)
	}

	for scope, types := range leaseByType {
		for leaseType, count := range types {
			emitGauge(ch, c.descDHCPLeasesByType, float64(count), scope, leaseType)
		}
	}
}
