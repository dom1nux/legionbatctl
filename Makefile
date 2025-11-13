# LegionBatCTL Go Version Makefile
# Supports daemon architecture with systemd service

# Variables
BINARY_NAME := legionbatctl
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%S)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -s -w"

# Default target
.PHONY: all
all: build

# Build binary
.PHONY: build
build:
	@echo "Building LegionBatCTL $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/legionbatctl

# Quick build in current directory (for development)
.PHONY: build-local
build-local:
	@echo "Building LegionBatCTL $(VERSION) locally..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/legionbatctl

# Install binary and system files
.PHONY: install
install: build
	@echo "Installing LegionBatCTL..."
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) /usr/bin/$(BINARY_NAME)
	install -Dm644 systemd/legionbatctl.service /etc/systemd/system/legionbatctl.service
	systemctl daemon-reload
	@echo "Installation complete. Use 'sudo systemctl enable --now legionbatctl.service' to enable daemon mode."

# Install and start service
.PHONY: install-start
install-start: install
	@echo "Installing and starting LegionBatCTL daemon..."
	systemctl enable --now legionbatctl.service
	@echo "Daemon started. Use 'legionbatctl status' to check status."

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "Uninstalling LegionBatCTL..."
	systemctl stop legionbatctl.service 2>/dev/null || true
	systemctl disable legionbatctl.service 2>/dev/null || true
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f /etc/systemd/system/legionbatctl.service
	rm -f /var/run/$(BINARY_NAME).sock
	rm -f /var/run/$(BINARY_NAME).pid
	rm -f /etc/$(BINARY_NAME).state
	rm -f /etc/$(BINARY_NAME).state.backup
	rm -f /etc/$(BINARY_NAME).state.tmp
	systemctl daemon-reload
	@echo "Uninstallation complete."

# Reinstall (uninstall + install)
.PHONY: reinstall
reinstall: uninstall install

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	go clean

# Service management targets
.PHONY: status start stop restart enable disable logs
status:
	@echo "=== LegionBatCTL Service Status ==="
	systemctl status legionbatctl.service || true
	@echo ""
	@echo "=== CLI Status ==="
	@if command -v $(BINARY_NAME) >/dev/null 2>&1; then \
		$(BINARY_NAME) status; \
	else \
		echo "Binary not found. Use 'sudo make install' first."; \
	fi

start:
	@echo "Starting LegionBatCTL daemon..."
	systemctl start legionbatctl.service
	@echo "Use 'make status' to check if it started successfully."

stop:
	@echo "Stopping LegionBatCTL daemon..."
	systemctl stop legionbatctl.service

restart:
	@echo "Restarting LegionBatCTL daemon..."
	systemctl restart legionbatctl.service
	@echo "Use 'make status' to check if it restarted successfully."

enable:
	@echo "Enabling LegionBatCTL daemon (auto-start on boot)..."
	systemctl enable legionbatctl.service
	@echo "Service will start automatically on boot."

disable:
	@echo "Disabling LegionBatCTL daemon..."
	systemctl disable legionbatctl.service

logs:
	@echo "=== LegionBatCTL Service Logs ==="
	journalctl -u legionbatctl.service -f

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

# Development targets
.PHONY: dev dev-daemon
dev: build-local
	@echo "Testing CLI locally..."
	@if [ -f $(BINARY_NAME) ]; then \
		$(BINARY_NAME) status; \
	else \
		echo "Binary not found. Run 'make build-local' first."; \
	fi

dev-daemon: build-local
	@echo "Testing daemon locally (foreground)..."
	@if [ -f $(BINARY_NAME) ]; then \
		echo "Starting daemon in foreground (Ctrl+C to stop)..."; \
		$(BINARY_NAME) daemon; \
	else \
		echo "Binary not found. Run 'make build-local' first."; \
	fi

# Test CLI commands after daemon is installed
.PHONY: test-cli
test-cli:
	@echo "Testing CLI commands..."
	@if command -v $(BINARY_NAME) >/dev/null 2>&1; then \
		echo "=== Status Command ==="; \
		$(BINARY_NAME) status; \
		echo; \
		echo "=== Set Threshold Command ==="; \
		$(BINARY_NAME) set-threshold 80; \
		echo; \
		echo "=== Enable Command ==="; \
		$(BINARY_NAME) enable; \
		echo; \
		echo "=== Final Status ==="; \
		$(BINARY_NAME) status; \
	else \
		echo "Binary not found. Use 'sudo make install' first."; \
	fi

# Generate dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Create initial state file (for testing)
.PHONY: init-state
init-state:
	@echo "Creating initial state file..."
	sudo install -Dm644 /dev/null /etc/$(BINARY_NAME).state
	@echo "State file created at /etc/$(BINARY_NAME).state"

# Quick installation and test
.PHONY: quick-install
quick-install: build-local install-start test-cli
	@echo "Quick installation and testing complete!"

# Test daemon functionality
.PHONY: test-daemon
test-daemon:
	@echo "=== Testing Daemon Functionality ==="
	@if systemctl is-active --quiet legionbatctl.service; then \
		echo "✅ Daemon is running"; \
		$(BINARY_NAME) status; \
	else \
		echo "❌ Daemon is not running. Use 'sudo make start' first."; \
	fi

# Help
.PHONY: help
help:
	@echo "LegionBatCTL Go Version - Daemon Architecture"
	@echo ""
	@echo "BUILD TARGETS:"
	@echo "  build           - Build binary to build/ directory"
	@echo "  build-local     - Build binary to current directory"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "$(YELLOW)INSTALLATION:$(NC)"
	@echo "  install         - Install binary and systemd service (requires root)"
	@echo "  install-start   - Install and start daemon (requires root)"
	@echo "  uninstall       - Remove from system (requires root)"
	@echo "  reinstall       - Uninstall and install (requires root)"
	@echo ""
	@echo "$(YELLOW)SERVICE MANAGEMENT:$(NC)"
	@echo "  start           - Start daemon (requires root)"
	@echo "  stop            - Stop daemon (requires root)"
	@echo "  restart         - Restart daemon (requires root)"
	@echo "  enable          - Enable auto-start on boot (requires root)"
	@echo "  disable         - Disable auto-start (requires root)"
	@echo "  status          - Show service and CLI status"
	@echo "  logs            - Show live service logs"
	@echo ""
	@echo "$(YELLOW)DEVELOPMENT:$(NC)"
	@echo "  dev             - Build and test CLI locally"
	@echo "  dev-daemon      - Build and run daemon in foreground"
	@echo "  test-cli        - Test all CLI commands"
	@echo "  test-daemon     - Test daemon functionality"
	@echo "  quick-install   - Build, install, start and test (requires root)"
	@echo ""
	@echo "$(YELLOW)TESTING:$(NC)"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo ""
	@echo "$(YELLOW)CODE QUALITY:$(NC)"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
	@echo "  security        - Run security check"
	@echo "  check           - Run all checks (fmt, lint, test, security)"
	@echo "  deps            - Download and tidy dependencies"
	@echo ""
	@echo "$(YELLOW)MISCELLANEOUS:$(NC)"
	@echo "  build-all       - Build for multiple architectures"
	@echo "  init-state      - Create initial state file"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "$(YELLOW)EXAMPLES:$(NC)"
	@echo "  make dev                    # Test CLI locally"
	@echo "  sudo make install-start    # Install and start daemon"
	@echo "  make status                 # Check service status"
	@echo "  make logs                   # View live logs"
	@echo "  sudo make restart           # Restart daemon"
	@echo "  sudo make uninstall         # Remove everything"