# AGENTS.md — Technitium DNS Prometheus Exporter

## Commands

```bash
make build        # go build ./...
make test         # go test ./... -race -coverprofile=coverage.out
make lint         # golangci-lint run
make run-local    # go run ./cmd/exporter --config config/config.yaml
make docker-build # multi-arch (linux/amd64,linux/arm64) via buildx
make helm-lint    # lint the Helm chart
make helm-template # render Helm chart to stdout for review
make compose-up   # full local stack: Technitium DNS + exporter + Prometheus + Grafana
make compose-down # tear down local stack
```

- Go module path: `technitium-dns-exporter` (matches directory name).
- Linting requires `golangci-lint` installed locally.
- Docker builds need `docker buildx`. Build context is repo root, Dockerfile at `deploy/Dockerfile`.
- Coverage output: `coverage.out` at repo root (gitignored).

## Architecture

- **Single binary**: `cmd/exporter/main.go` — entrypoint. Loads YAML config, registers collectors, serves `/metrics`.
- **Multi-target**: each `targets[]` entry gets its own `prometheus.Registry` and a dedicated `TechnitiumCollector` instance (see `internal/collector/collector.go:63-64`). Registry per target, fan-out on scrape.
- **Sub-collection**: within a single `Collect()` call, sub-collectors (dashboard, zones, cache, blocklist, forwarders, DHCP, cluster) run concurrently via `sync.WaitGroup` goroutines — NOT `errgroup`. Each sub-collector makes its own API calls to the target.
- **API client**: `internal/client/client.go` — sends token as query param or form field. All responses wrapped in `{"status":"ok","response":...}`. `doPostRequest` exists but is unused.

## Config

- Path: `--config` flag, then `$EXPORTER_CONFIG` env var, then `./config/config.yaml` default.
- Supports `${VAR_NAME}` env-var interpolation in any string field.
- SIGHUP triggers config reload (hot-reload). Old config kept on reload failure.
- Log level is determined by `$EXPORTER_LOG_LEVEL` env var at startup, not by the config's `log_level` field. The config field only changes what `setLogLevel` prints, not the handler's actual level.

## Testing

- Tests use `httptest.NewServer` to mock the Technitium API; fixtures inlined in test files.
- Prometheus metric assertions use `testutil.GatherAndCompare` from `prometheus/client_golang`.
- Config tests are table-driven with temp files created via `t.TempDir()`.
- Run a single test: `go test -run TestName ./internal/...`

## Collector conventions

- Metric descriptors are defined in the `New()` constructor and cached on the struct. **Exception**: `cache_max_entries`, `cache_save_enabled`, `cache_serve_stale_enabled` create `prometheus.NewDesc` inline inside `collectCacheStats` — avoid adding more inline descriptors; follow the constructor pattern.
- `instance` label auto-set to `target.Name` in the constructor (not from config labels).
- `scrapeSuccess` is hardcoded to `1` regardless of sub-collector failures — partial failures log errors but don't flip the success gauge.

## Deploy

- Helm chart: `deploy/helm/technitium-dns-exporter/` (Chart.yaml, values.yaml, templates/).
- Grafana dashboards: `dashboards/overview.json` etc. — provisioned via sidecar-compatible JSON.
- Container runs as non-root in `gcr.io/distroless/static:nonroot`.
- `deploy/docker-compose.yml` includes Technitium DNS, exporter, Prometheus, and Grafana on ports 5380, 9119, 9090, 3000 respectively. Grafana credentials: admin/admin.

## Repo conventions

- `.opencode/`, `opencode.json`, `plan.md`, `.ctx/` are all gitignored. Do not rely on these being present in a fresh clone.
- No CI workflows in repo. All verification is local (`make build && make test && make lint`).
- The `opencode.json` instructions block references `docs/guidelines.md` and `docs/security.md` — these files do not exist.
