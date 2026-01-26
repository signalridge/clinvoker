package output

import "errors"

var (
	// ErrInvalidEventType is returned when trying to parse content for wrong event type.
	ErrInvalidEventType = errors.New("invalid event type for content")

	// ErrUnknownBackend is returned when the backend is not recognized.
	ErrUnknownBackend = errors.New("unknown backend")

	// ErrParseError is returned when event parsing fails.
	ErrParseError = errors.New("failed to parse event")
)
