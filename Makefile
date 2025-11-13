# legionbatctl - Simplified Makefile for mise-based Go installation

# Variables
BINARY_NAME := legionbatctl
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%S)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -s -w"

# Default target
.PHONY: all
all: build

# Build binary (uses mise for Go)
.PHONY: build
build:
	@echo "Building legionbatctl $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/legionbatctl

# Install binary and systemd files, then start service
# NOTE: Run 'make build' first to create the binary
.PHONY: install
install:
	@echo "Installing legionbatctl..."
	@if [ ! -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		echo "Binary not found. Run 'make build' first."; \
		exit 1; \
	fi
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) /usr/bin/$(BINARY_NAME)
	install -Dm644 systemd/legionbatctl.service /etc/systemd/system/legionbatctl.service
	systemctl daemon-reload
	systemctl enable --now legionbatctl.service
	@echo "Installation complete. Daemon started and enabled."

# Uninstall (stop service, remove files)
.PHONY: uninstall
uninstall:
	@echo "Uninstalling legionbatctl..."
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

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Service management targets
.PHONY: status restart logs
status:
	@echo "=== legionbatctl Service Status ==="
	systemctl status legionbatctl.service || true
	@echo ""
	@echo "=== CLI Status ==="
	@if command -v $(BINARY_NAME) >/dev/null 2>&1; then \
		$(BINARY_NAME) status; \
	else \
		echo "Binary not found. Use 'sudo make install' first."; \
	fi

restart:
	@echo "Restarting legionbatctl daemon..."
	systemctl restart legionbatctl.service
	@echo "Use 'make status' to check if it restarted successfully."

logs:
	@echo "=== legionbatctl Service Logs ==="
	journalctl -u legionbatctl.service -f

# Development targets
.PHONY: dev
dev: build
	@echo "Testing CLI locally..."
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) status; \
	else \
		echo "Binary not found. Run 'make build' first."; \
	fi

# Help
.PHONY: help
help:
	@echo "legionbatctl - Simplified Makefile for mise-based Go installation"
	@echo ""
	@echo "MAIN TARGETS:"
	@echo "  build           - Build binary to build/ directory (uses mise Go)"
	@echo "  install         - Install binary, systemd files and start service (requires root)"
	@echo "                  NOTE: Run 'make build' first"
	@echo "  uninstall       - Stop service and remove all files (requires root)"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "SERVICE MANAGEMENT:"
	@echo "  status          - Show service and CLI status"
	@echo "  restart         - Restart daemon (requires root)"
	@echo "  logs            - Show live service logs"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo "  dev             - Build and test CLI locally"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "WORKFLOW:"
	@echo "  make build                    # Build binary (as user, with mise)"
	@echo "  sudo make install            # Install and start daemon (as root)"
	@echo "  make status                  # Check service status"
	@echo "  make logs                    # View live logs"
	@echo "  sudo make uninstall          # Remove everything"