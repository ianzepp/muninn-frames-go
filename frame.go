package muninnframes

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Status is the lifecycle state of a frame in a request/response exchange.
type Status string

const (
	StatusRequest Status = "request"
	StatusItem    Status = "item"
	StatusBulk    Status = "bulk"
	StatusDone    Status = "done"
	StatusError   Status = "error"
	StatusCancel  Status = "cancel"
)

// ErrInvalidStatus indicates that a status value is outside the supported set.
var ErrInvalidStatus = errors.New("invalid frame status")

// ValidationError indicates a required business field is missing or malformed.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid frame %s: %s", e.Field, e.Message)
}

// Frame is the shared envelope for every request and response.
type Frame struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parent_id,omitempty"`
	CreatedMS int64   `json:"created_ms"`
	ExpiresIn int64   `json:"expires_in"`
	From      *string `json:"from,omitempty"`
	Call      string  `json:"call"`
	Status    Status  `json:"status"`
	Trace     any     `json:"trace,omitempty"`
	Data      any     `json:"data"`
}

// IsTerminal returns true if the status ends a response stream.
func (s Status) IsTerminal() bool {
	return s == StatusDone || s == StatusError || s == StatusCancel
}

// IsValid returns true if the status is one of the supported variants.
func (s Status) IsValid() bool {
	switch s {
	case StatusRequest, StatusItem, StatusBulk, StatusDone, StatusError, StatusCancel:
		return true
	default:
		return false
	}
}

// MarshalJSON serializes a status as a lowercase JSON string.
func (s Status) MarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidStatus, string(s))
	}

	return json.Marshal(string(s))
}

// UnmarshalJSON parses a lowercase JSON string into a status.
func (s *Status) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	status := Status(value)
	if !status.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidStatus, value)
	}

	*s = status
	return nil
}

// Validate checks that the frame has the minimal required fields.
func (f Frame) Validate() error {
	if f.ID == "" {
		return ValidationError{Field: "id", Message: "must not be empty"}
	}
	if f.Call == "" {
		return ValidationError{Field: "call", Message: "must not be empty"}
	}
	if !f.Status.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidStatus, string(f.Status))
	}
	if f.Data == nil {
		return ValidationError{Field: "data", Message: "must not be nil"}
	}
	return nil
}

// EncodeFrame validates and serializes a frame to JSON bytes.
func EncodeFrame(frame Frame) ([]byte, error) {
	if err := frame.Validate(); err != nil {
		return nil, err
	}

	return json.Marshal(frame)
}

// DecodeFrame parses JSON bytes into a validated frame.
func DecodeFrame(data []byte) (Frame, error) {
	var frame Frame
	if err := json.Unmarshal(data, &frame); err != nil {
		return Frame{}, err
	}

	if frame.Data == nil {
		frame.Data = map[string]any{}
	}

	if err := frame.Validate(); err != nil {
		return Frame{}, err
	}

	return frame, nil
}
