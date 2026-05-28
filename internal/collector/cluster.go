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

	emitGauge(ch, c.descHeartbeatInterval, float64(state.HeartbeatRefreshInterval))
	emitGauge(ch, c.descClusterHeartbeatRetryInterval, float64(state.HeartbeatRetryInterval))
	emitGauge(ch, c.descClusterConfigRefreshInterval, float64(state.ConfigRefreshInterval))
	emitGauge(ch, c.descClusterConfigRetryInterval, float64(state.ConfigRetryInterval))

	if state.ConfigLastSynced != "" {
		if lastSynced, err := time.Parse(time.RFC3339Nano, state.ConfigLastSynced); err == nil {
			emitGauge(ch, c.descClusterConfigLastSynced, float64(lastSynced.Unix()))
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
		emitGauge(ch, c.descClusterNodeState, stateValue,
			node.Name, node.Type, node.IPAddress,
		)
	}
}
