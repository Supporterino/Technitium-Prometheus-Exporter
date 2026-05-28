package collector

import (
	"context"

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
