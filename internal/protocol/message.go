package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewRequest creates a new request message
func NewRequest(command string, params map[string]interface{}) *Message {
	return &Message{
		Type: "request",
		ID:   generateID(),
		Request: &Request{
			Command: command,
			Params:  params,
		},
	}
}

// NewResponse creates a new response message
func NewResponse(requestID string, success bool, data interface{}, errMsg string) *Message {
	response := &Response{
		Success: success,
		Data:    data,
	}

	if !success && errMsg != "" {
		response.Error = errMsg
	}

	return &Message{
		Type:     "response",
		ID:       requestID,
		Response: response,
	}
}

// NewErrorResponse creates a new error response message
func NewErrorResponse(requestID string, err error) *Message {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	return NewResponse(requestID, false, nil, errMsg)
}

// NewSuccessResponse creates a new success response message
func NewSuccessResponse(requestID string, data interface{}) *Message {
	return NewResponse(requestID, true, data, "")
}

// Validate validates the message format
func (m *Message) Validate() error {
	if m.Type != "request" && m.Type != "response" {
		return fmt.Errorf("invalid message type: %s", m.Type)
	}

	if m.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}

	switch m.Type {
	case "request":
		if m.Request == nil {
			return fmt.Errorf("request message missing request data")
		}
		if !IsValidCommand(m.Request.Command) {
			return ErrInvalidCommand
		}

	case "response":
		if m.Response == nil {
			return fmt.Errorf("response message missing response data")
		}
	}

	return nil
}

// IsRequest returns true if this is a request message
func (m *Message) IsRequest() bool {
	return m.Type == "request"
}

// IsResponse returns true if this is a response message
func (m *Message) IsResponse() bool {
	return m.Type == "response"
}

// GetRequest safely returns the request (for request messages)
func (m *Message) GetRequest() *Request {
	if m.IsRequest() {
		return m.Request
	}
	return nil
}

// GetResponse safely returns the response (for response messages)
func (m *Message) GetResponse() *Response {
	if m.IsResponse() {
		return m.Response
	}
	return nil
}

// generateID generates a unique request ID
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "req-" + hex.EncodeToString(bytes)
}