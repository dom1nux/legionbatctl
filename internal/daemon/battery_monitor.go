package daemon

import (
	"fmt"
	"time"
)

// monitorBattery monitors battery level and adjusts conservation mode accordingly
func (d *Daemon) monitorBattery() {
	// Create ticker for periodic checks
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.checkBatteryAndAdjust()
		case <-d.done:
			return
		}
	}
}

// checkBatteryAndAdjust checks battery level and adjusts conservation mode if needed
func (d *Daemon) checkBatteryAndAdjust() {
	if d.stateManager == nil {
		return
	}

	// Read current battery information
	batteryLevel, conservationMode, charging, err := d.readBatteryInfo()
	if err != nil {
		fmt.Printf("Failed to read battery info: %v\n", err)
		return
	}

	// Update state with current battery info
	if err := d.stateManager.UpdateBatteryInfo(batteryLevel, conservationMode, charging); err != nil {
		fmt.Printf("Failed to update battery info in state: %v\n", err)
		return
	}

	// Only process if we're on AC power and management is enabled
	if !charging || !d.stateManager.GetConservationEnabled() {
		fmt.Printf("Skipping check: AC connected=%v, conservation enabled=%v\n",
			charging, d.stateManager.GetConservationEnabled())
		return
	}

	// Determine if we need to change conservation mode
	shouldEnable := d.stateManager.ShouldEnableConservation()
	shouldDisable := d.stateManager.ShouldDisableConservation()

	// Change conservation mode if needed
	if shouldEnable && !conservationMode {
		if err := d.setConservationMode(true); err != nil {
			fmt.Printf("Failed to enable conservation mode: %v\n", err)
		} else {
			fmt.Printf("Enabled conservation mode (battery: %d%%, threshold: %d%%)\n",
				batteryLevel, d.stateManager.GetChargeThreshold())
		}
	} else if shouldDisable && conservationMode {
		if err := d.setConservationMode(false); err != nil {
			fmt.Printf("Failed to disable conservation mode: %v\n", err)
		} else {
			fmt.Printf("Disabled conservation mode (battery: %d%%, threshold: %d%%)\n",
				batteryLevel, d.stateManager.GetChargeThreshold())
		}
	}

	// Adjust check interval based on proximity to threshold
	d.adjustCheckInterval(batteryLevel)
}

// adjustCheckInterval adjusts the monitoring interval based on battery level
func (d *Daemon) adjustCheckInterval(batteryLevel int) {
	if d.stateManager == nil {
		return
	}

	threshold := d.stateManager.GetChargeThreshold()
	difference := abs(batteryLevel - threshold)

	var newInterval time.Duration

	if difference < 5 {
		// Within 5% of threshold - check every 15 seconds
		newInterval = 15 * time.Second
	} else if difference < 15 {
		// Within 15% of threshold - check every 30 seconds
		newInterval = 30 * time.Second
	} else {
		// Far from threshold - check every 2 minutes
		newInterval = 2 * time.Minute
	}

	// Update interval if it changed
	if newInterval != d.checkInterval {
		d.checkInterval = newInterval
		fmt.Printf("Adjusted check interval to %v (battery: %d%%, threshold: %d%%)\n",
			newInterval, batteryLevel, threshold)
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// GetMonitoringStatus returns the current monitoring status
func (d *Daemon) GetMonitoringStatus() MonitoringStatus {
	if d.stateManager == nil {
		return MonitoringStatus{
			Enabled: false,
			Interval: d.checkInterval,
		}
	}

	return MonitoringStatus{
		Enabled:          d.stateManager.GetConservationEnabled(),
		Threshold:        d.stateManager.GetChargeThreshold(),
		CurrentBattery:   d.stateManager.GetBatteryLevel(),
		ConservationMode: d.stateManager.GetConservationMode(),
		Charging:         d.stateManager.IsCharging(),
		Interval:         d.checkInterval,
	}
}

// MonitoringStatus represents the current battery monitoring status
type MonitoringStatus struct {
	Enabled          bool          `json:"enabled"`
	Threshold        int           `json:"threshold"`
	CurrentBattery   int           `json:"current_battery"`
	ConservationMode bool          `json:"conservation_mode"`
	Charging         bool          `json:"charging"`
	Interval         time.Duration `json:"interval"`
}

// SetMonitoringInterval sets a custom monitoring interval
func (d *Daemon) SetMonitoringInterval(interval time.Duration) {
	d.SetCheckInterval(interval)
}

// GetNextCheckTime returns when the next battery check will occur
func (d *Daemon) GetNextCheckTime() time.Time {
	return time.Now().Add(d.checkInterval)
}