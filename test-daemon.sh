#!/bin/bash

# LegionBatCTL Daemon Testing Script
# This script provides a safe way to test the daemon with logging and debugging

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project and test configuration
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$PROJECT_DIR/legionbatctl"
TEST_SOCKET="/tmp/legionbatctl-test.sock"
TEST_STATE="/tmp/legionbatctl-test.state"
TEST_PID="/tmp/legionbatctl-test.pid"
LOG_FILE="$PROJECT_DIR/test-daemon.log"

echo -e "${BLUE}=== LegionBatCTL Daemon Testing Script ===${NC}"
echo "Project directory: $PROJECT_DIR"
echo "Binary: $BINARY"
echo "Test socket: $TEST_SOCKET"
echo "Test state: $TEST_STATE"
echo "Log file: $LOG_FILE"
echo

# Function to check if binary exists
check_binary() {
    if [ ! -f "$BINARY" ]; then
        echo -e "${RED}❌ Binary not found at $BINARY${NC}"
        echo "Building binary..."
        cd "$PROJECT_DIR"
        go build -o legionbatctl ./cmd/legionbatctl
        echo -e "${GREEN}✅ Binary built successfully${NC}"
    else
        echo -e "${GREEN}✅ Binary found${NC}"
    fi
}

# Function to cleanup test artifacts
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up test artifacts...${NC}"

    # Stop daemon if running
    if [ -f "$TEST_PID" ]; then
        local pid=$(cat "$TEST_PID" 2>/dev/null || echo "")
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            echo "Stopping daemon (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                kill -9 "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$TEST_PID"
    fi

    # Remove socket and state files
    rm -f "$TEST_SOCKET"
    rm -f "$TEST_STATE"
    rm -f "$TEST_STATE.tmp"
    rm -f "$TEST_STATE.backup"

    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

# Function to test daemon in foreground
test_foreground() {
    echo -e "${BLUE}🚀 Testing daemon in foreground mode...${NC}"
    echo "Press Ctrl+C to stop the daemon"
    echo

    # Set test environment variables
    export SOCKET_PATH="$TEST_SOCKET"
    export STATE_PATH="$TEST_STATE"

    # Start daemon in foreground
    cd "$PROJECT_DIR"
    "$BINARY" daemon
}

# Function to test daemon in background
test_background() {
    echo -e "${BLUE}🚀 Starting daemon in background...${NC}"

    # Cleanup first
    cleanup

    # Set test environment variables
    export SOCKET_PATH="$TEST_SOCKET"
    export STATE_PATH="$TEST_STATE"

    # Start daemon in background
    cd "$PROJECT_DIR"
    nohup "$BINARY" daemon > "$LOG_FILE" 2>&1 &
    local pid=$!

    # Save PID
    echo "$pid" > "$TEST_PID"

    echo -e "${GREEN}✅ Daemon started in background (PID: $pid)${NC}"
    echo "Log file: $LOG_FILE"
    echo "Socket: $TEST_SOCKET"

    # Wait a moment for startup
    sleep 2

    # Check if daemon is responsive
    if [ -S "$TEST_SOCKET" ]; then
        echo -e "${GREEN}✅ Socket created successfully${NC}"
    else
        echo -e "${RED}❌ Socket not found - daemon may have failed to start${NC}"
        echo "Check log file: $LOG_FILE"
        return 1
    fi

    # Test CLI connectivity
    echo -e "${BLUE}🔍 Testing CLI connectivity...${NC}"
    SOCKET_PATH="$TEST_SOCKET" "$BINARY" status || {
        echo -e "${RED}❌ CLI connectivity test failed${NC}"
        return 1
    }

    echo -e "${GREEN}✅ CLI connectivity test passed${NC}"
}

# Function to show daemon status
show_status() {
    echo -e "${BLUE}📊 Daemon Status:${NC}"

    if [ -f "$TEST_PID" ]; then
        local pid=$(cat "$TEST_PID" 2>/dev/null || echo "")
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            echo -e "${GREEN}✅ Daemon is running (PID: $pid)${NC}"

            # Show status using CLI
            if [ -S "$TEST_SOCKET" ]; then
                echo
                SOCKET_PATH="$TEST_SOCKET" "$BINARY" status
            fi
        else
            echo -e "${RED}❌ Daemon is not running${NC}"
            rm -f "$TEST_PID"
        fi
    else
        echo -e "${RED}❌ Daemon is not running${NC}"
    fi

    # Show test artifacts
    echo
    echo "Test artifacts:"
    [ -f "$TEST_PID" ] && echo "  PID file: $TEST_PID (exists)"
    [ -S "$TEST_SOCKET" ] && echo "  Socket: $TEST_SOCKET (exists)"
    [ -f "$TEST_STATE" ] && echo "  State: $TEST_STATE (exists)"
    [ -f "$LOG_FILE" ] && echo "  Log: $LOG_FILE ($(wc -l < "$LOG_FILE") lines)"
}

# Function to show logs
show_logs() {
    echo -e "${BLUE}📋 Daemon Logs:${NC}"

    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        echo -e "${YELLOW}⚠️  Log file not found: $LOG_FILE${NC}"
    fi
}

# Function to test CLI commands
test_cli() {
    echo -e "${BLUE}🧪 Testing CLI Commands:${NC}"

    if [ ! -S "$TEST_SOCKET" ]; then
        echo -e "${RED}❌ Daemon socket not found. Start the daemon first.${NC}"
        return 1
    fi

    export SOCKET_PATH="$TEST_SOCKET"

    echo "Testing status command..."
    "$BINARY" status
    echo

    echo "Testing set-threshold command..."
    "$BINARY" set-threshold 80
    echo

    echo "Testing enable command..."
    "$BINARY" enable
    echo

    echo "Testing status command again..."
    "$BINARY" status
    echo

    echo -e "${GREEN}✅ CLI command tests completed${NC}"
}

# Function to install systemd test service
install_test_service() {
    echo -e "${BLUE}🔧 Installing test systemd service...${NC}"

    sudo cp "$PROJECT_DIR/systemd/legionbatctl-test.service" /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl disable legionbatctl-test.service  # Ensure it doesn't start automatically

    echo -e "${GREEN}✅ Test service installed${NC}"
    echo "Commands:"
    echo "  sudo systemctl start legionbatctl-test.service   # Start test service"
    echo "  sudo systemctl stop legionbatctl-test.service    # Stop test service"
    echo "  sudo systemctl status legionbatctl-test.service  # Check status"
    echo "  sudo journalctl -u legionbatctl-test.service -f  # View logs"
}

# Function to show help
show_help() {
    echo -e "${BLUE}Usage: $0 [COMMAND]${NC}"
    echo
    echo "Commands:"
    echo "  build          Build the binary"
    echo "  foreground     Run daemon in foreground (for debugging)"
    echo "  background     Start daemon in background"
    echo "  stop           Stop background daemon"
    echo "  status         Show daemon status"
    echo "  logs           Show daemon logs (tail -f)"
    echo "  cli            Test CLI commands"
    echo "  cleanup        Clean up test artifacts"
    echo "  install-service Install systemd test service"
    echo "  help           Show this help"
    echo
    echo "Examples:"
    echo "  $0 build && $0 background    # Build and start daemon"
    echo "  $0 status                      # Check daemon status"
    echo "  $0 cli                         # Test CLI commands"
    echo "  $0 logs                        # View live logs"
    echo "  $0 stop                        # Stop daemon"
}

# Main script logic
case "${1:-help}" in
    "build")
        check_binary
        ;;
    "foreground")
        check_binary
        test_foreground
        ;;
    "background")
        check_binary
        test_background
        ;;
    "stop")
        cleanup
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs
        ;;
    "cli")
        test_cli
        ;;
    "cleanup")
        cleanup
        ;;
    "install-service")
        install_test_service
        ;;
    "help"|*)
        show_help
        ;;
esac