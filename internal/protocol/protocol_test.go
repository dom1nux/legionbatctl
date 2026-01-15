package protocol

import (
	"strings"
	"testing"
)

func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "valid request",
			msg: &Message{
				Type: "request",
				ID:   "test-123",
				Request: &Request{
					Command: "enable",
					Params:  map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name: "valid response",
			msg: &Message{
				Type: "response",
				ID:   "test-123",
				Response: &Response{
					Success: true,
					Data:    "test data",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			msg: &Message{
				Type: "invalid",
				ID:   "test-123",
			},
			wantErr: true,
		},
		{
			name: "empty ID",
			msg: &Message{
				Type: "request",
				ID:   "",
				Request: &Request{
					Command: "enable",
				},
			},
			wantErr: true,
		},
		{
			name: "request missing request data",
			msg: &Message{
				Type: "request",
				ID:   "test-123",
			},
			wantErr: true,
		},
		{
			name: "response missing response data",
			msg: &Message{
				Type: "response",
				ID:   "test-123",
			},
			wantErr: true,
		},
		{
			name: "invalid command",
			msg: &Message{
				Type: "request",
				ID:   "test-123",
				Request: &Request{
					Command: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	params := map[string]interface{}{
		"threshold": 80,
	}

	msg := NewRequest("set_threshold", params)

	if msg.Type != "request" {
		t.Errorf("Expected type 'request', got %s", msg.Type)
	}

	if msg.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if !strings.HasPrefix(msg.ID, "req-") {
		t.Errorf("Expected ID to start with 'req-', got %s", msg.ID)
	}

	if msg.Request.Command != "set_threshold" {
		t.Errorf("Expected command 'set_threshold', got %s", msg.Request.Command)
	}

	if msg.Request.Params["threshold"] != 80 {
		t.Errorf("Expected threshold 80, got %v", msg.Request.Params["threshold"])
	}
}

func TestNewResponse(t *testing.T) {
	data := map[string]interface{}{
		"result": "success",
	}

	// Test success response
	msg := NewResponse("test-123", true, data, "")

	if msg.Type != "response" {
		t.Errorf("Expected type 'response', got %s", msg.Type)
	}

	if msg.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", msg.ID)
	}

	if !msg.Response.Success {
		t.Error("Expected success to be true")
	}

	if msg.Response.Error != "" {
		t.Errorf("Expected empty error, got %s", msg.Response.Error)
	}

	// Test error response
	msgErr := NewResponse("test-456", false, nil, "test error")

	if msgErr.Response.Success {
		t.Error("Expected success to be false")
	}

	if msgErr.Response.Error != "test error" {
		t.Errorf("Expected error 'test error', got %s", msgErr.Response.Error)
	}
}

func TestIsValidCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{CmdEnable, true},
		{CmdDisable, true},
		{CmdStatus, true},
		{CmdSetThreshold, true},
		{CmdDaemonStatus, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := IsValidCommand(tt.cmd)
			if got != tt.want {
				t.Errorf("IsValidCommand(%s) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestValidateThreshold(t *testing.T) {
	tests := []struct {
		threshold int
		wantErr   bool
	}{
		{60, false},  // Minimum valid
		{80, false},  // Valid
		{100, false}, // Maximum valid
		{59, true},   // Below minimum
		{101, true},  // Above maximum
		{0, true},    // Invalid
		{-10, true},  // Invalid
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.threshold)), func(t *testing.T) {
			err := ValidateThreshold(tt.threshold)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThreshold(%d) error = %v, wantErr %v", tt.threshold, err, tt.wantErr)
			}
		})
	}
}

func TestMessageTypeHelpers(t *testing.T) {
	// Test request message
	reqMsg := NewRequest("enable", nil)
	if !reqMsg.IsRequest() {
		t.Error("Expected IsRequest() to return true")
	}
	if reqMsg.IsResponse() {
		t.Error("Expected IsResponse() to return false")
	}
	if reqMsg.GetRequest() == nil {
		t.Error("Expected GetRequest() to return non-nil")
	}
	if reqMsg.GetResponse() != nil {
		t.Error("Expected GetResponse() to return nil")
	}

	// Test response message
	respMsg := NewResponse("test-123", true, nil, "")
	if respMsg.IsRequest() {
		t.Error("Expected IsRequest() to return false")
	}
	if !respMsg.IsResponse() {
		t.Error("Expected IsResponse() to return true")
	}
	if respMsg.GetRequest() != nil {
		t.Error("Expected GetRequest() to return nil")
	}
	if respMsg.GetResponse() == nil {
		t.Error("Expected GetResponse() to return non-nil")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{"invalid threshold", ErrInvalidThreshold, "threshold must be between 60 and 100"},
		{"daemon not running", ErrDaemonNotRunning, "daemon not running"},
		{"hardware not supported", ErrHardwareNotSupported, "hardware not supported"},
		{"permission denied", ErrPermissionDenied, "permission denied"},
		{"invalid command", ErrInvalidCommand, "invalid command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error.Error() = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}

	if id2 == "" {
		t.Error("Expected non-empty ID")
	}

	if id1 == id2 {
		t.Error("Expected different IDs")
	}

	if !strings.HasPrefix(id1, "req-") {
		t.Errorf("Expected ID to start with 'req-', got %s", id1)
	}

	if !strings.HasPrefix(id2, "req-") {
		t.Errorf("Expected ID to start with 'req-', got %s", id2)
	}

	if len(id1) != len("req-")+16 { // "req-" + 16 hex characters
		t.Errorf("Expected ID length 20, got %d", len(id1))
	}
}
