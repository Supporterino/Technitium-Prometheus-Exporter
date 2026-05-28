.PHONY: build test lint docker-build helm-lint helm-template compose-up compose-down run-local

build:
	go build ./...

test:
	go test ./... -race -coverprofile=coverage.out

lint:
	golangci-lint run

docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 -f deploy/Dockerfile -t technitium-dns-exporter:latest .

helm-lint:
	helm lint deploy/helm/technitium-dns-exporter

helm-template:
	helm template technitium-dns-exporter deploy/helm/technitium-dns-exporter

compose-up:
	docker compose -f deploy/docker-compose.yml up -d

compose-down:
	docker compose -f deploy/docker-compose.yml down -v

run-local:
	go run ./cmd/exporter --config config/config.yaml
