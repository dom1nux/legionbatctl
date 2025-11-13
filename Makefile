# LegionBatCTL Go Version Makefile

# Variables
BINARY_NAME := legionbatctl
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/dom1nux/legionbatctl/pkg/version.Version=$(VERSION) -X github.com/dom1nux/legionbatctl/pkg/version.Commit=$(COMMIT)"

# Default target
.PHONY: all
all: build

# Build binary
.PHONY: build
build:
	@echo "Building LegionBatCTL $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/legionbatctl

# Install binary and system files
.PHONY: install
install: build
	@echo "Installing LegionBatCTL..."
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) /usr/bin/$(BINARY_NAME)
	install -Dm644 systemd/legionbatctl.service /etc/systemd/system/legionbatctl.service
	install -Dm644 systemd/legionbatctl.timer /etc/systemd/system/legionbatctl.timer
	install -Dm644 man/legionbatctl.1 /usr/share/man/man1/legionbatctl.1
	install -Dm644 docs/legionbatctl.conf /etc/legionbatctl.conf
	systemctl daemon-reload
	@echo "Installation complete. Use 'sudo systemctl enable --now legionbatctl.timer' to enable auto mode."

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "Uninstalling LegionBatCTL..."
	systemctl stop legionbatctl.timer 2>/dev/null || true
	systemctl disable legionbatctl.timer 2>/dev/null || true
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f /etc/systemd/system/legionbatctl.service
	rm -f /etc/systemd/system/legionbatctl.timer
	rm -f /usr/share/man/man1/legionbatctl.1
	rm -f /etc/legionbatctl.conf
	systemctl daemon-reload
	@echo "Uninstallation complete."

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with coverage
.PHONY: test-coverage
test-coverage: test
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Run security check
.PHONY: security
security:
	@echo "Running security check..."
	gosec ./...

# Run all checks
.PHONY: check
check: fmt lint test security

# Build for different architectures
.PHONY: build-all
build-all:
	@echo "Building for multiple architectures..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/legionbatctl

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/legionbatctl

# Development mode - build and run immediately
.PHONY: dev
dev: build
	sudo $(BUILD_DIR)/$(BINARY_NAME) status

# Auto mode development
.PHONY: dev-auto
dev-auto: build
	sudo $(BUILD_DIR)/$(BINARY_NAME) auto

# Generate dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Create initial config file
.PHONY: config
config:
	@echo "Creating default configuration..."
	install -Dm644 /dev/null /etc/legionbatctl.conf
	echo "CONSERVATION_ENABLED=1" >> /etc/legionbatctl.conf
	echo "CHARGE_THRESHOLD=80" >> /etc/legionbatctl.conf
	@echo "Configuration created at /etc/legionbatctl.conf"

# Test auto mode manually
.PHONY: test-auto
test-auto: build
	@echo "Testing auto mode (requires root)..."
	sudo $(BUILD_DIR)/$(BINARY_NAME) auto --dry-run

# Help
.PHONY: help
help:
	@echo "LegionBatCTL Go Version - Available targets:"
	@echo ""
	@echo "  build          - Build binary"
	@echo "  install        - Build and install to system"
	@echo "  uninstall      - Remove from system"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  security       - Run security check"
	@echo "  check          - Run all checks (fmt, lint, test, security)"
	@echo "  build-all      - Build for multiple architectures"
	@echo "  dev            - Build and run status command"
	@echo "  dev-auto       - Build and run auto command"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  config         - Create default configuration file"
	@echo "  test-auto      - Test auto mode in dry-run"
	@echo "  help           - Show this help message"