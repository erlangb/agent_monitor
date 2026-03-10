# AgentMonitor build and quality targets.

BINDIR := bin

.PHONY: fmt lint test run-deep run-search build-agent-monitor run-monitor deps mocks vet run-tea run-terminal

# Generate mocks with mockery (reads .mockery.yaml).
mocks:
	mockery

# Format Go code.
fmt:
	gofmt -s -w .
	go mod tidy

# Run linter (golangci-lint if available).
lint:
	@command -v .golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed (optional)"; exit 0)
	golangci-lint run ./...

# Vet Go code.
vet:
	go vet ./...

# Run tests.
test:
	go test -v -race ./...

# Download and tidy Go module dependencies.
deps:
	go mod download
	go mod tidy

# Run the agent directly.
run:
	go run ./cmd/monitor

# Run with the bubbletea TUI runner.
run-tea:
	go run ./cmd/monitor -runner=tea

# Run with the plain terminal runner.
run-terminal:
	go run ./cmd/monitor -runner=terminal

# Build agent monitor binary into bin/.
build-agent-monitor:
	@mkdir -p $(BINDIR)
	go build -o $(BINDIR)/agent-monitor ./cmd/monitor
