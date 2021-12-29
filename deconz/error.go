package deconz

// EventError represents an error during the retrieval or decoding of events
// It wraps another error
// Users can specify if the error is recoverable
type EventError struct {
	error
	recoverable bool
}

// NewEventError creates a new EventError that wraps another error
func NewEventError(err error, recoverable bool) EventError {
	return EventError{err, recoverable}
}

func (e EventError) Recoverable() bool {
	return e.recoverable
}
