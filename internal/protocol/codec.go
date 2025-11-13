package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

// Codec handles encoding and decoding of protocol messages
type Codec struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

// NewCodec creates a new codec for the given reader/writer
func NewCodec(rw io.ReadWriter) *Codec {
	return &Codec{
		encoder: json.NewEncoder(rw),
		decoder: json.NewDecoder(rw),
	}
}

// Encode writes a message to the writer
func (c *Codec) Encode(msg *Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	return c.encoder.Encode(msg)
}

// Decode reads a message from the reader
func (c *Codec) Decode() (*Message, error) {
	var msg Message
	if err := c.decoder.Decode(&msg); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	return &msg, nil
}

// SendRequest encodes and sends a request message
func (c *Codec) SendRequest(command string, params map[string]interface{}) (*Message, error) {
	msg := NewRequest(command, params)
	if err := c.Encode(msg); err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	return msg, nil
}

// SendResponse encodes and sends a response message
func (c *Codec) SendResponse(requestID string, success bool, data interface{}, errMsg string) error {
	msg := NewResponse(requestID, success, data, errMsg)
	return c.Encode(msg)
}

// SendErrorResponse encodes and sends an error response
func (c *Codec) SendErrorResponse(requestID string, err error) error {
	msg := NewErrorResponse(requestID, err)
	return c.Encode(msg)
}

// SendSuccessResponse encodes and sends a success response
func (c *Codec) SendSuccessResponse(requestID string, data interface{}) error {
	msg := NewSuccessResponse(requestID, data)
	return c.Encode(msg)
}

// ReceiveMessage reads and returns a message
func (c *Codec) ReceiveMessage() (*Message, error) {
	return c.Decode()
}

// WriteMessage writes a message to the writer (convenience function)
func WriteMessage(w io.Writer, msg *Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	encoder := json.NewEncoder(w)
	return encoder.Encode(msg)
}

// ReadMessage reads a message from the reader (convenience function)
func ReadMessage(r io.Reader) (*Message, error) {
	decoder := json.NewDecoder(r)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return nil, fmt.Errorf("read message failed: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	return &msg, nil
}