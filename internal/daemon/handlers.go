package daemon

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dom1nux/legionbatctl/internal/protocol"
)

// RunDaemon starts the daemon in the current process
func RunDaemon(socketPath, statePath string) error {
	daemon := NewDaemon(socketPath, statePath)

	// Check if already running
	if isDaemonRunning(socketPath) {
		return fmt.Errorf("daemon is already running (socket: %s)", socketPath)
	}

	fmt.Printf("legionbatctl daemon starting...\n")
	fmt.Printf("Socket: %s\n", daemon.GetSocketPath())
	fmt.Printf("State: %s\n", daemon.GetStatePath())
	fmt.Printf("PID: %d\n", daemon.GetPID())

	// Run daemon (blocks until shutdown)
	return daemon.Run()
}

// StopDaemon stops a running daemon
func StopDaemon(socketPath string) error {
	if !isDaemonRunning(socketPath) {
		return fmt.Errorf("daemon is not running")
	}

	// Try graceful shutdown via socket
	client, err := NewDaemonClient(socketPath)
	if err == nil {
		// Send shutdown command (if implemented)
		_ = client.Close()
	}

	// Remove socket file to force stop
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	// Remove PID file
	pidPath := filepath.Join(filepath.Dir(socketPath), "legionbatctl.pid")
	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	return nil
}

// RestartDaemon restarts the daemon
func RestartDaemon(socketPath, statePath string) error {
	// Stop existing daemon
	if err := StopDaemon(socketPath); err != nil {
		// Don't fail if daemon wasn't running
		fmt.Printf("Warning: %v\n", err)
	}

	// Start new daemon
	return RunDaemon(socketPath, statePath)
}

// DaemonStatus returns the status of the daemon
func DaemonStatus(socketPath string) (*DaemonStatusInfo, error) {
	if !isDaemonRunning(socketPath) {
		return &DaemonStatusInfo{
			Running: false,
		}, nil
	}

	client, err := NewDaemonClient(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer client.Close()

	response, err := client.SendRequest(protocol.CmdDaemonStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get daemon status: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("daemon returned error: %s", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	status := &DaemonStatusInfo{
		Running:    true,
		SocketPath: socketPath,
	}

	if pid, ok := data["pid"].(float64); ok {
		status.PID = int(pid)
	}

	if uptime, ok := data["uptime"].(string); ok {
		status.Uptime = uptime
	}

	if version, ok := data["version"].(string); ok {
		status.Version = version
	}

	if stateFile, ok := data["state_file"].(string); ok {
		status.StateFile = stateFile
	}

	return status, nil
}

// DaemonStatusInfo represents daemon status information
type DaemonStatusInfo struct {
	Running    bool   `json:"running"`
	PID        int    `json:"pid"`
	Uptime     string `json:"uptime"`
	Version    string `json:"version"`
	SocketPath string `json:"socket_path"`
	StateFile  string `json:"state_file"`
}

// isDaemonRunning checks if the daemon is running
func isDaemonRunning(socketPath string) bool {
	// Check if socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return false
	}

	// Try to connect to socket
	client, err := NewDaemonClient(socketPath)
	if err != nil {
		return false
	}
	defer client.Close()

	// Try to send a simple status request
	response, err := client.SendRequest(protocol.CmdDaemonStatus, nil)
	if err != nil {
		return false
	}

	return response.Success
}

// DaemonClient represents a client for communicating with the daemon
type DaemonClient struct {
	socketPath string
}

// NewDaemonClient creates a new daemon client
func NewDaemonClient(socketPath string) (*DaemonClient, error) {
	client := &DaemonClient{
		socketPath: socketPath,
	}

	// Test connection
	conn, err := client.connect()
	if err != nil {
		return nil, err
	}
	conn.Close()

	return client, nil
}

// connect creates a connection to the daemon
func (c *DaemonClient) connect() (net.Conn, error) {
	return net.Dial("unix", c.socketPath)
}

// SendRequest sends a request to the daemon
func (c *DaemonClient) SendRequest(command string, params map[string]interface{}) (*protocol.Response, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	codec := protocol.NewCodec(conn)

	// Send request
	_, err = codec.SendRequest(command, params)
	if err != nil {
		return nil, err
	}

	// Receive response
	msg, err := codec.ReceiveMessage()
	if err != nil {
		return nil, err
	}

	return msg.GetResponse(), nil
}

// Close closes the daemon client
func (c *DaemonClient) Close() error {
	// No persistent connection to close
	return nil
}

// GetDaemonPID returns the PID of a running daemon
func GetDaemonPID(socketPath string) (int, error) {
	pidPath := filepath.Join(filepath.Dir(socketPath), "legionbatctl.pid")

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	_, err = fmt.Sscanf(string(data), "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return pid, nil
}

// KillDaemon kills the daemon by PID
func KillDaemon(socketPath string) error {
	pid, err := GetDaemonPID(socketPath)
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Try graceful shutdown first
	err = process.Signal(os.Signal(syscall.SIGTERM))
	if err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if process is still running
	err = process.Signal(os.Signal(syscall.Signal(0)))
	if err == nil {
		// Still running, force kill
		err = process.Kill()
		if err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	}

	return nil
}
