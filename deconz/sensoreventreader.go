package deconz

import (
	ctx "context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

// EventReader is an interface to read single events
// For testability, it is decoupled from the SensorEventReader which runs the main business logic around connection
// management
type EventReader interface {
	ReadEvent() (Event, error)
	Dial(ctx ctx.Context) error
	Close() error
}

// SensorEventReader uses an EventReader (for example WsReader) to provide a channel of SensorEvent
// The SensorEventReader handles connection losses and reconnection attempts of the underlying EventReader
type SensorEventReader struct {
	reader  EventReader
	running bool
	done chan bool
}

// Start starts a go routine that reads events from the associated EventReader
// It returns the channel to retrieve events from
func (r *SensorEventReader) Start(ctx ctx.Context) (<-chan *SensorEvent, error) {

	out := make(chan *SensorEvent)

	if r.reader == nil {
		return nil, errors.New("Cannot run without an EventReader from which to read events")
	}

	if r.running {
		return nil, errors.New("SensorEventReader is already running")
	}

	r.running = true

	go func() {
	REDIAL:
		for {
			r.connect(ctx)

			for {
				select {
				case <-ctx.Done():
					log.Debug("Aborting websocket connection")
					close(out)

					default:
					// read events until connection fails
					e, err := r.reader.ReadEvent()
					if err != nil {
						if eerr, ok := err.(EventError); ok && eerr.Recoverable() {
							log.Errorf("Dropping event due to error: %s", err)
							continue
						}
						continue REDIAL
					}

					// we only care about sensor events
					se, ok := e.(SensorEvent)
					if !ok {
						log.Debugf("Dropping non-sensor event type %s", e.Resource())
						continue
					}

					// send event on channel
					out <- &se
				}
			}
		}

	}()

	return out, nil
}

// connect connects the EventReader
// If the connection fails, it retries every 10s as long as the given Context is not Done()
func (r *SensorEventReader) connect(ctx ctx.Context) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		select {
		case <-ctx.Done():
			log.Debug("Aborting websocket connection")
		default:
			err := r.reader.Dial(ctx)

			if err != nil {
				log.Errorf("Error connecting deCONZ websocket: %s\nAttempting reconnect in 10s...", err)
			} else {
				log.Infof("deCONZ websocket connected")
				return

			}
		}
	}
}

// Shutdown closes the reader, closing the connection to deCONZ
// The method blocks until all background tasks are terminated or the given Context is aborted
func (r *SensorEventReader) Shutdown(ctx ctx.Context) {
	r.running = false

	go func() {
		err := r.reader.Close()
		if err != nil {
			log.Error("Failed to close websocket", err)
			return
		}

		log.Infof("deCONZ websocket closed")

		r.done <- true
		return
	}()

	select {
	case <-ctx.Done():
		return
	case <-r.done:
		return
	}
}