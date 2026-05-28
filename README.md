# Technitium DNS Prometheus Exporter

Prometheus exporter for [Technitium DNS Server](https://technitium.com/dns/). Collects metrics from one or more Technitium DNS instances and exposes them for Prometheus scraping.

## Metrics Reference

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `technitium_dns_scrape_success` | Gauge | instance, job | Whether the scrape was successful (1) or failed (0) |
| `technitium_dns_scrape_duration_seconds` | Gauge | instance, job | Duration of the scrape |
| `technitium_dns_queries_total` | Counter | instance, job | Total DNS queries |
| `technitium_dns_queries_noerror_total` | Counter | instance, job | Queries with NOERROR |
| `technitium_dns_queries_servfail_total` | Counter | instance, job | Queries with SERVFAIL |
| `technitium_dns_queries_nxdomain_total` | Counter | instance, job | Queries with NXDOMAIN |
| `technitium_dns_queries_refused_total` | Counter | instance, job | Queries with REFUSED |
| `technitium_dns_queries_authoritative_total` | Counter | instance, job | Authoritative queries |
| `technitium_dns_queries_recursive_total` | Counter | instance, job | Recursive queries |
| `technitium_dns_queries_cached_total` | Counter | instance, job | Cached queries |
| `technitium_dns_queries_blocked_total` | Counter | instance, job | Blocked queries |
| `technitium_dns_queries_dropped_total` | Counter | instance, job | Dropped queries |
| `technitium_dns_clients_active` | Gauge | instance, job | Active clients |
| `technitium_dns_cache_entries` | Gauge | instance, job | Cache entry count |
| `technitium_dns_cache_max_entries` | Gauge | instance, job | Max cache entries |
| `technitium_dns_cache_save_enabled` | Gauge | instance, job | Save cache enabled |
| `technitium_dns_cache_serve_stale_enabled` | Gauge | instance, job | Serve-stale enabled |
| `technitium_dns_zones_count` | Gauge | instance, job | Zone count |
| `technitium_dns_zone_info` | Gauge | instance, job, zone, type, dnssec_status | Zone details |
| `technitium_dns_zone_expiry_timestamp_seconds` | Gauge | instance, job, zone, type | Zone expiry |
| `technitium_dns_allowed_zones_count` | Gauge | instance, job | Allowed zones |
| `technitium_dns_blocked_zones_count` | Gauge | instance, job | Blocked zones |
| `technitium_dns_allowlist_zones_count` | Gauge | instance, job | Allow list zones |
| `technitium_dns_blocklist_zones_count` | Gauge | instance, job | Blocklist zones |
| `technitium_dns_blocking_enabled` | Gauge | instance, job | Blocking enabled |
| `technitium_dns_blocklist_update_interval_hours` | Gauge | instance, job | Blocklist update interval |
| `technitium_dns_blocklist_next_update_timestamp_seconds` | Gauge | instance, job | Next blocklist update |
| `technitium_dns_forwarders_count` | Gauge | instance, job | Forwarder count |
| `technitium_dns_forwarder_info` | Gauge | instance, job, address, protocol | Forwarder details |
| `technitium_dns_dhcp_leases_count` | Gauge | instance, job, scope | DHCP leases per scope |
| `technitium_dns_dhcp_scope_enabled` | Gauge | instance, job, scope | DHCP scope enabled |
| `technitium_dns_cluster_node_state` | Gauge | instance, job, node, node_type, ip_address | Cluster node state |
| `technitium_dns_cluster_heartbeat_interval_seconds` | Gauge | instance, job | Heartbeat interval |

## Quick Start (Local)

```bash
make compose-up
```

This starts a full stack:
- Technitium DNS server on port 5380
- Exporter on port 9119
- Prometheus on port 9090
- Grafana on port 3000 (admin/admin)

Open http://localhost:3000 to view dashboards.

To stop:
```bash
make compose-down
```

## Configuration

See `config/config.yaml` for an example configuration.

```yaml
exporter:
  listen_address: ":9119"
  metrics_path: "/metrics"
  scrape_timeout: 30s
  log_level: "info"

targets:
  - name: "dns-primary"
    url: "http://dns1.example.com:5380"
    api_token: "your-api-token"
    labels:
      environment: "production"
    features:
      dhcp: true
      cluster: false
```

Environment variables can be interpolated using `${VAR_NAME}` syntax. The config path can be set via `--config` flag or `EXPORTER_CONFIG` environment variable.

## Deployment (Helm)

```bash
helm install technitium-dns-exporter deploy/helm/technitium-dns-exporter \
  --set config.targets[0].name=dns-primary \
  --set config.targets[0].url=http://dns1.example.com:5380 \
  --set config.targets[0].api_token=your-token
```

### With ServiceMonitor (Prometheus Operator)

```bash
helm install technitium-dns-exporter deploy/helm/technitium-dns-exporter \
  --set serviceMonitor.enabled=true \
  --set config.targets[0].name=dns-primary \
  --set config.targets[0].url=http://dns1.example.com:5380 \
  --set config.targets[0].api_token=your-token
```

## Building

```bash
make build       # Compile locally
make test        # Run tests with race detection
make lint        # Run golangci-lint
make docker-build # Build multi-arch Docker image
make run-local   # Run with local config
```

## Contributing

1. Ensure `make build` and `make test` pass
2. Run `make lint` to check for issues
3. Open a PR with a clear description
