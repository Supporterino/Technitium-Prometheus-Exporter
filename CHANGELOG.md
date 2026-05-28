# Changelog

## [1.0.6](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.5...v1.0.6) (2026-05-28)


### Bug Fixes

* **ci:** 🐛 Lowercase Helm OCI push owner to fix invalid repository re… ([6d258f9](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/6d258f9dd6c3ec33370edca7a29ac1ab6a4a6cc4))
* **ci:** 🐛 Lowercase Helm OCI push owner to fix invalid repository reference ([136814a](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/136814ada2bb30ee6735aeb67397b36eccce3f27))

## [1.0.5](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.4...v1.0.5) (2026-05-28)


### Bug Fixes

* **build:** 🐛 Fix docker build — remove multi-platform flag, use gore… ([d41f2cd](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/d41f2cdb009065061ad3f092b391c3d4154078aa))
* **build:** 🐛 Fix docker build — remove multi-platform flag, use goreleaser binary ([fd20c1e](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/fd20c1ef24b72f1eb33d3536ef8bd03eeba2c57d))

## [1.0.4](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.3...v1.0.4) (2026-05-28)


### Bug Fixes

* **build:** 🐛 Lowercase docker image tag in goreleaser for OCI registry ([f85c99b](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/f85c99b1fef7805a23eb01a903977b96488abb7e))
* **build:** 🐛 Lowercase docker image tag in goreleaser for OCI registry ([b943930](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/b943930ffdb68c48aa97727332436aa0b5cebf5d))

## [1.0.3](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.2...v1.0.3) (2026-05-28)


### Bug Fixes

* **build:** 🐛 Remove SBOM generation to fix goreleaser syft dependency ([bcb6d7e](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/bcb6d7e4fe0dccfa7f0ff764d281a5ab105f4d51))
* **build:** 🐛 Remove SBOM generation to fix goreleaser syft dependency ([8494004](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/8494004979475d224cd93ec155cd0a09173cb90e))

## [1.0.2](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.1...v1.0.2) (2026-05-28)


### Bug Fixes

* **build:** 🐛 Lower go directive to 1.24.0 for CI compatibility ([093549a](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/093549abd0cf45f0675bff55d4b67a4f1c3f5d62))
* **build:** 🐛 Lower go directive to 1.24.0 for CI compatibility ([207aa73](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/207aa73def1f7f0719f04c0e1b947890d01360fe))

## [1.0.1](https://github.com/Supporterino/Technitium-Prometheus-Exporter/compare/v1.0.0...v1.0.1) (2026-05-28)


### Performance Improvements

* ⚡️ Apply codebase optimisation sweep ([59163d6](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/59163d6363863d409aacfd0f25e1d0505f341249))
* ⚡️ Apply codebase optimisation sweep ([cc58f69](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/cc58f6915ba31f796aba9d1cfcb088a95672cf35))

## 1.0.0 (2026-05-28)


### Features

* ✨ Add 57 new metrics and update Grafana dashboards ([e2dd918](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/e2dd918f04ce449a97d40a794b68c939c5e0f410))
* ✨ Add Prometheus exporter for Technitium DNS Server ([1a3cd9d](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/1a3cd9d1a46e2b2dab97a706ffe865468d34b283))


### Bug Fixes

* ci and dependencies ([cf9bef8](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/cf9bef8222fe6dab3e0827b94c4055a0ab8d0538))
* **ci:** 🐛 Use PAT token for release-please to enable downstream workflow triggers ([888bae6](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/888bae646f7f5ff52101bb70deb63056f42e410f))
* **helm:** 🐛 Add checksum/config annotation to restart pods on ConfigMap change ([283e984](https://github.com/Supporterino/Technitium-Prometheus-Exporter/commit/283e984f1e6b40b8bf5a03d108b331f8c3096f5f))
