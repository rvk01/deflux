package deconz

type EventError struct {
	error
	recoverable bool
}

func NewEventError(err error, recoverable bool) EventError {
	return EventError{err, recoverable}
}

func (e EventError) Recoverable() bool {
	return e.recoverable
}
