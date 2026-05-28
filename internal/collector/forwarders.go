package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func (c *TechnitiumCollector) collectForwarderStats(ctx context.Context, ch chan<- prometheus.Metric) {
	settings, err := c.client.GetSettings(ctx)
	if err != nil {
		c.logError("failed to get settings for forwarder metrics", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("technitium_dns_forwarders_count", "Number of configured forwarders.",
			nil, c.target.Labels),
		prometheus.GaugeValue, float64(len(settings.Forwarders)),
	)

	for _, fwd := range settings.Forwarders {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("technitium_dns_forwarder_info", "Forwarder address info.",
				[]string{"address", "protocol"}, c.target.Labels),
			prometheus.GaugeValue, 1,
			fwd, settings.ForwarderProtocol,
		)
	}
}
