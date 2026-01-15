package client

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/dom1nux/legionbatctl/internal/protocol"
)

const (
	DefaultSocketPath = "/var/run/legionbatctl.sock"
	DefaultTimeout    = 10 * time.Second
)

// Client represents a client for communicating with the legionbatctl daemon
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new client instance
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		// Check environment variable first
		socketPath = os.Getenv("SOCKET_PATH")
		if socketPath == "" {
			socketPath = DefaultSocketPath
		}
	}

	return &Client{
		socketPath: socketPath,
		timeout:    DefaultTimeout,
	}
}

// NewClientWithTimeout creates a new client with custom timeout
func NewClientWithTimeout(socketPath string, timeout time.Duration) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}

	return &Client{
		socketPath: socketPath,
		timeout:    timeout,
	}
}

// SetTimeout sets the connection timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout returns the current timeout
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}

// GetSocketPath returns the socket path
func (c *Client) GetSocketPath() string {
	return c.socketPath
}

// IsDaemonRunning checks if the daemon is running
func (c *Client) IsDaemonRunning() bool {
	conn, err := c.connect()
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// SendRequest sends a request to the daemon and returns the response
func (c *Client) SendRequest(command string, params map[string]interface{}) (*protocol.Response, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	codec := protocol.NewCodec(conn)

	// Send request
	_, err = codec.SendRequest(command, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	msg, err := codec.ReceiveMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	if !msg.IsResponse() {
		return nil, fmt.Errorf("expected response message, got %s", msg.Type)
	}

	response := msg.GetResponse()
	if response == nil {
		return nil, fmt.Errorf("missing response data")
	}

	return response, nil
}

// connect creates a connection to the daemon with timeout
func (c *Client) connect() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial socket %s: %w", c.socketPath, err)
	}

	// Set socket timeout
	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set socket timeout: %w", err)
	}

	return conn, nil
}

// Enable enables battery management
func (c *Client) Enable() error {
	response, err := c.SendRequest(protocol.CmdEnable, nil)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("enable command failed: %s", response.Error)
	}

	return nil
}

// Disable disables battery management
func (c *Client) Disable() error {
	response, err := c.SendRequest(protocol.CmdDisable, nil)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("disable command failed: %s", response.Error)
	}

	return nil
}

// SetThreshold sets the charge threshold
func (c *Client) SetThreshold(threshold int) error {
	params := map[string]interface{}{
		"threshold": threshold,
	}

	response, err := c.SendRequest(protocol.CmdSetThreshold, params)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("set_threshold command failed: %s", response.Error)
	}

	return nil
}

// GetStatus retrieves the current system status
func (c *Client) GetStatus() (*protocol.StatusData, error) {
	response, err := c.SendRequest(protocol.CmdStatus, nil)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("status command failed: %s", response.Error)
	}

	// Parse response data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	status := &protocol.StatusData{}

	if conservationEnabled, ok := data["conservation_enabled"].(bool); ok {
		status.ConservationEnabled = conservationEnabled
	}

	if threshold, ok := data["threshold"].(float64); ok {
		status.Threshold = int(threshold)
	}

	if currentMode, ok := data["current_mode"].(string); ok {
		status.CurrentMode = currentMode
	}

	if batteryLevel, ok := data["battery_level"].(float64); ok {
		status.BatteryLevel = int(batteryLevel)
	}

	if conservationMode, ok := data["conservation_mode"].(bool); ok {
		status.ConservationMode = conservationMode
	}

	if charging, ok := data["charging"].(bool); ok {
		status.Charging = charging
	}

	if lastAction, ok := data["last_action"].(string); ok {
		status.LastAction = lastAction
	}

	if daemonUptime, ok := data["daemon_uptime"].(string); ok {
		status.DaemonUptime = daemonUptime
	}

	if hardwareSupported, ok := data["hardware_supported"].(bool); ok {
		status.HardwareSupported = hardwareSupported
	}

	return status, nil
}

// GetDaemonStatus retrieves daemon status information
func (c *Client) GetDaemonStatus() (*protocol.DaemonStatusData, error) {
	response, err := c.SendRequest(protocol.CmdDaemonStatus, nil)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("daemon_status command failed: %s", response.Error)
	}

	// Parse response data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	status := &protocol.DaemonStatusData{}

	if running, ok := data["running"].(bool); ok {
		status.Running = running
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

	if socketPath, ok := data["socket_path"].(string); ok {
		status.SocketPath = socketPath
	}

	if stateFile, ok := data["state_file"].(string); ok {
		status.StateFile = stateFile
	}

	return status, nil
}

// Ping sends a ping to the daemon to check if it's responsive
func (c *Client) Ping() error {
	_, err := c.SendRequest(protocol.CmdDaemonStatus, nil)
	return err
}

// Close closes the client (no-op as connections are short-lived)
func (c *Client) Close() error {
	// No persistent connection to close
	return nil
}

// String returns a string representation of the client
func (c *Client) String() string {
	return fmt.Sprintf("legionbatctl Client{socket: %s, timeout: %v}", c.socketPath, c.timeout)
}
