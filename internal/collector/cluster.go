package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectCluster(ctx context.Context, ch chan<- prometheus.Metric) {
	if !c.target.Features.Cluster {
		return
	}

	state, err := c.client.GetClusterState(ctx)
	if err != nil {
		c.logError("failed to get cluster state", err)
		return
	}

	if !state.ClusterInitialized {
		return
	}

	ch <- prometheus.MustNewConstMetric(c.descHeartbeatInterval, prometheus.GaugeValue, float64(state.HeartbeatRefreshInterval))
	ch <- prometheus.MustNewConstMetric(c.descClusterHeartbeatRetryInterval, prometheus.GaugeValue, float64(state.HeartbeatRetryInterval))
	ch <- prometheus.MustNewConstMetric(c.descClusterConfigRefreshInterval, prometheus.GaugeValue, float64(state.ConfigRefreshInterval))
	ch <- prometheus.MustNewConstMetric(c.descClusterConfigRetryInterval, prometheus.GaugeValue, float64(state.ConfigRetryInterval))

	if state.ConfigLastSynced != "" {
		if lastSynced, err := time.Parse(time.RFC3339Nano, state.ConfigLastSynced); err == nil {
			ch <- prometheus.MustNewConstMetric(c.descClusterConfigLastSynced, prometheus.GaugeValue, float64(lastSynced.Unix()))
		}
	}

	for _, node := range state.Nodes {
		stateValue := float64(0)
		switch node.State {
		case "Connected":
			stateValue = 1
		case "Self":
			stateValue = 2
		}
		ch <- prometheus.MustNewConstMetric(c.descClusterNodeState, prometheus.GaugeValue, stateValue,
			node.Name, node.Type, node.IPAddress,
		)
	}
}
