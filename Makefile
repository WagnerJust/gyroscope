BIN := bin/gyroscope
LDFLAGS := -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) \
           -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) \
           -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build test test-race vet
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/gyroscope
test:
	go test ./...
test-race:
	go test -race ./...
vet:
	go vet ./...
