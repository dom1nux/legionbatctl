package daemon

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/dom1nux/legionbatctl/internal/state"
)

const (
	DefaultSocketPath = "/var/run/legionbatctl.sock"
	DefaultStatePath  = "/etc/legionbatctl.state"
	DefaultPIDPath    = "/var/run/legionbatctl.pid"
)

// Daemon represents the battery management daemon
type Daemon struct {
	socketPath string
	statePath  string
	pidPath    string

	// Core components
	stateManager *state.Manager
	listener     net.Listener

	// Control
	mutex sync.RWMutex
	done  chan bool
	running bool

	// Configuration
	checkInterval time.Duration
	logLevel      string
}

// NewDaemon creates a new daemon instance
func NewDaemon(socketPath, statePath string) *Daemon {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	if statePath == "" {
		statePath = DefaultStatePath
	}

	return &Daemon{
		socketPath: socketPath,
		statePath:  statePath,
		pidPath:    filepath.Join(filepath.Dir(socketPath), "legionbatctl.pid"),
		done:       make(chan bool),
		running:    false,
		checkInterval: 30 * time.Second, // Default check interval
	}
}

// Start starts the daemon
func (d *Daemon) Start() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.running {
		return fmt.Errorf("daemon is already running")
	}

	// Initialize state manager
	d.stateManager = state.NewManager(d.statePath)

	// Load existing state or create default
	if err := d.stateManager.Load(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Set daemon info in state
	if err := d.stateManager.SetDaemonInfo(os.Getpid()); err != nil {
		return fmt.Errorf("failed to set daemon info: %w", err)
	}

	// Create socket listener
	if err := d.createSocketListener(); err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}

	// Write PID file
	if err := d.writePIDFile(); err != nil {
		d.listener.Close()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Set running flag
	d.running = true

	// Start goroutines
	go d.serveConnections()
	go d.monitorBattery()
	go d.handleSignals()

	return nil
}

// Run starts the daemon and blocks until shutdown
func (d *Daemon) Run() error {
	if err := d.Start(); err != nil {
		return err
	}

	// Block until daemon is stopped
	<-d.done
	return nil
}

// Stop stops the daemon gracefully
func (d *Daemon) Stop() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return nil // Already stopped
	}

	// Signal all goroutines to stop
	close(d.done)
	d.running = false

	// Close socket listener
	if d.listener != nil {
		d.listener.Close()
	}

	// Remove socket file
	os.Remove(d.socketPath)

	// Remove PID file
	os.Remove(d.pidPath)

	return nil
}

// IsRunning returns whether the daemon is running
func (d *Daemon) IsRunning() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.running
}

// GetState returns the current state
func (d *Daemon) GetState() state.State {
	if d.stateManager != nil {
		return d.stateManager.GetState()
	}
	return state.State{}
}

// createSocketListener creates the Unix socket listener
func (d *Daemon) createSocketListener() error {
	// Remove existing socket file if it exists
	os.Remove(d.socketPath)

	// Create socket directory if needed
	socketDir := filepath.Dir(d.socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Create listener
	listener, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", d.socketPath, err)
	}

	// Set socket permissions
	if err := os.Chmod(d.socketPath, 0777); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	d.listener = listener
	return nil
}

// writePIDFile writes the PID file
func (d *Daemon) writePIDFile() error {
	pid := os.Getpid()
	pidDir := filepath.Dir(d.pidPath)

	// Create PID directory if needed
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Write PID file
	if err := os.WriteFile(d.pidPath, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// handleSignals handles system signals for graceful shutdown
func (d *Daemon) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				// Graceful shutdown
				d.Stop()
				return
			case syscall.SIGHUP:
				// Reload configuration (placeholder for future use)
				d.reloadConfiguration()
			}
		case <-d.done:
			return
		}
	}
}

// reloadConfiguration reloads daemon configuration
func (d *Daemon) reloadConfiguration() {
	// Placeholder for future configuration reloading
	// Currently just log that we received SIGHUP
	fmt.Printf("Received SIGHUP, configuration reload not implemented yet\n")
}

// GetPID returns the daemon PID
func (d *Daemon) GetPID() int {
	return os.Getpid()
}

// GetUptime returns how long the daemon has been running
func (d *Daemon) GetUptime() time.Duration {
	if d.stateManager != nil {
		return d.stateManager.GetUptime()
	}
	return 0
}

// GetVersion returns daemon version information
func (d *Daemon) GetVersion() string {
	// This will be injected at build time
	return "dev"
}

// GetSocketPath returns the socket path
func (d *Daemon) GetSocketPath() string {
	return d.socketPath
}

// GetStatePath returns the state file path
func (d *Daemon) GetStatePath() string {
	return d.statePath
}

// SetCheckInterval sets the battery monitoring interval
func (d *Daemon) SetCheckInterval(interval time.Duration) {
	// Apply validation
	if interval < 10*time.Second {
		interval = 10 * time.Second // Minimum 10 seconds
	}
	if interval > 10*time.Minute {
		interval = 10 * time.Minute // Maximum 10 minutes
	}

	d.checkInterval = interval
}

// GetCheckInterval returns the current check interval
func (d *Daemon) GetCheckInterval() time.Duration {
	return d.checkInterval
}

// IsHealthy checks if the daemon is healthy
func (d *Daemon) IsHealthy() bool {
	if !d.IsRunning() {
		return false
	}

	// Check if socket is accessible
	conn, err := net.DialTimeout("unix", d.socketPath, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()

	// Check if state manager is working
	if d.stateManager == nil {
		return false
	}

	return true
}