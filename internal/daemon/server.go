package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dom1nux/legionbatctl/internal/protocol"
)

// serveConnections handles incoming socket connections
func (d *Daemon) serveConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.done:
				// Daemon is shutting down
				return
			default:
				// Log error but continue accepting connections
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		// Handle connection in a goroutine
		go d.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg protocol.Message
		if err := decoder.Decode(&msg); err != nil {
			if !isConnectionClosed(err) {
				fmt.Printf("Decode error: %v\n", err)
			}
			return
		}

		// Process request
		response := d.processRequest(&msg)

		// Send response
		if err := encoder.Encode(response); err != nil {
			fmt.Printf("Encode error: %v\n", err)
			return
		}

		// If this is a response message, we're done with this connection
		if msg.IsResponse() {
			return
		}
	}
}

// processRequest processes a single request message
func (d *Daemon) processRequest(req *protocol.Message) *protocol.Message {
	if !req.IsRequest() {
		return protocol.NewErrorResponse(req.ID, fmt.Errorf("invalid message type"))
	}

	request := req.GetRequest()
	if request == nil {
		return protocol.NewErrorResponse(req.ID, fmt.Errorf("missing request data"))
	}

	var response interface{}
	var err error

	switch request.Command {
	case protocol.CmdEnable:
		response, err = d.handleEnable(request.Params)
	case protocol.CmdDisable:
		response, err = d.handleDisable(request.Params)
	case protocol.CmdStatus:
		response, err = d.handleStatus(request.Params)
	case protocol.CmdSetThreshold:
		response, err = d.handleSetThreshold(request.Params)
	case protocol.CmdDaemonStatus:
		response, err = d.handleDaemonStatus(request.Params)
	default:
		err = fmt.Errorf("unknown command: %s", request.Command)
	}

	if err != nil {
		return protocol.NewErrorResponse(req.ID, err)
	}

	return protocol.NewSuccessResponse(req.ID, response)
}

// handleEnable handles the enable command
func (d *Daemon) handleEnable(params map[string]interface{}) (interface{}, error) {
	if d.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Enable conservation management
	if err := d.stateManager.EnableConservation(); err != nil {
		return nil, fmt.Errorf("failed to enable conservation: %w", err)
	}

	// If conservation should be enabled immediately, do it
	if d.stateManager.ShouldEnableConservation() {
		if err := d.setConservationMode(true); err != nil {
			return nil, fmt.Errorf("failed to set conservation mode: %w", err)
		}
	}

	state := d.stateManager.GetState()
	return protocol.EnableData{
		Message:     "Battery management enabled",
		Threshold:   state.ChargeThreshold,
		CurrentMode: state.CurrentMode,
	}, nil
}

// handleDisable handles the disable command
func (d *Daemon) handleDisable(params map[string]interface{}) (interface{}, error) {
	if d.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Disable conservation mode first
	if err := d.setConservationMode(false); err != nil {
		return nil, fmt.Errorf("failed to disable conservation mode: %w", err)
	}

	// Then disable management
	if err := d.stateManager.DisableConservation(); err != nil {
		return nil, fmt.Errorf("failed to disable conservation: %w", err)
	}

	state := d.stateManager.GetState()
	return protocol.DisableData{
		Message:     "Battery management disabled",
		CurrentMode: state.CurrentMode,
	}, nil
}

// handleStatus handles the status command
func (d *Daemon) handleStatus(params map[string]interface{}) (interface{}, error) {
	if d.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Read current battery information
	batteryLevel, conservationMode, charging, err := d.readBatteryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to read battery info: %w", err)
	}

	// Update state with current battery info
	if err := d.stateManager.UpdateBatteryInfo(batteryLevel, conservationMode, charging); err != nil {
		// Don't fail the request, just log the error
		fmt.Printf("Failed to update battery info: %v\n", err)
	}

	state := d.stateManager.GetState()
	return protocol.StatusData{
		ConservationEnabled: state.ConservationEnabled,
		Threshold:          state.ChargeThreshold,
		CurrentMode:        state.CurrentMode,
		BatteryLevel:       batteryLevel,
		ConservationMode:   conservationMode,
		Charging:           charging,
		LastAction:         state.LastAction,
		LastActionTime:     state.LastActionTime,
		DaemonUptime:       d.GetUptime().String(),
		HardwareSupported:  true, // TODO: Implement hardware detection
	}, nil
}

// handleSetThreshold handles the set_threshold command
func (d *Daemon) handleSetThreshold(params map[string]interface{}) (interface{}, error) {
	if d.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Extract threshold from params
	thresholdValue, ok := params["threshold"]
	if !ok {
		return nil, fmt.Errorf("threshold parameter required")
	}

	threshold, ok := thresholdValue.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid threshold value type")
	}

	thresholdInt := int(threshold)

	// Validate threshold
	if err := protocol.ValidateThreshold(thresholdInt); err != nil {
		return nil, err
	}

	// Set threshold
	if err := d.stateManager.SetChargeThreshold(thresholdInt); err != nil {
		return nil, fmt.Errorf("failed to set threshold: %w", err)
	}

	return protocol.SetThresholdData{
		Message:   fmt.Sprintf("Charge threshold set to %d%%", thresholdInt),
		Threshold: thresholdInt,
	}, nil
}

// handleDaemonStatus handles the daemon_status command
func (d *Daemon) handleDaemonStatus(params map[string]interface{}) (interface{}, error) {
	return protocol.DaemonStatusData{
		Running:    d.IsRunning(),
		PID:        d.GetPID(),
		Uptime:     d.GetUptime().String(),
		Version:    d.GetVersion(),
		SocketPath: d.GetSocketPath(),
		StateFile:  d.GetStatePath(),
	}, nil
}

// readBatteryInfo reads current battery information
func (d *Daemon) readBatteryInfo() (int, bool, bool, error) {
	// Read battery capacity
	capacity, err := os.ReadFile("/sys/class/power_supply/BAT0/capacity")
	if err != nil {
		return 0, false, false, fmt.Errorf("failed to read battery capacity: %w", err)
	}

	var batteryLevel int
	_, err = fmt.Sscanf(string(capacity), "%d", &batteryLevel)
	if err != nil {
		return 0, false, false, fmt.Errorf("failed to parse battery capacity: %w", err)
	}

	// Read conservation mode status
	conservationData, err := os.ReadFile("/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode")
	if err != nil {
		return batteryLevel, false, false, fmt.Errorf("failed to read conservation mode: %w", err)
	}

	var conservationMode int
	_, err = fmt.Sscanf(string(conservationData), "%d", &conservationMode)
	if err != nil {
		return batteryLevel, false, false, fmt.Errorf("failed to parse conservation mode: %w", err)
	}

	// Read charging status
	statusData, err := os.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		return batteryLevel, conservationMode == 1, false, fmt.Errorf("failed to read battery status: %w", err)
	}

	status := strings.TrimSpace(string(statusData))
	charging := status == "Charging"

	return batteryLevel, conservationMode == 1, charging, nil
}

// setConservationMode sets the hardware conservation mode
func (d *Daemon) setConservationMode(enable bool) error {
	conservationPath := "/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode"

	var value string
	if enable {
		value = "1"
		fmt.Printf("Enabling conservation mode (writing 1 to %s)\n", conservationPath)
	} else {
		value = "0"
		fmt.Printf("Disabling conservation mode (writing 0 to %s)\n", conservationPath)
	}

	// Write to conservation mode file
	err := os.WriteFile(conservationPath, []byte(value), 0644)
	if err != nil {
		return fmt.Errorf("failed to write conservation mode: %w", err)
	}

	// Verify the change was applied
	data, err := os.ReadFile(conservationPath)
	if err != nil {
		return fmt.Errorf("failed to verify conservation mode change: %w", err)
	}

	actualValue := strings.TrimSpace(string(data))
	expectedValue := strings.TrimSpace(value)

	if actualValue != expectedValue {
		return fmt.Errorf("conservation mode not updated: expected %s, got %s", expectedValue, actualValue)
	}

	return nil
}

// isConnectionClosed checks if the error indicates a closed connection
func isConnectionClosed(err error) bool {
	return err != nil && (err.Error() == "EOF" || err.Error() == "use of closed network connection")
}