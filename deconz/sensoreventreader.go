package deconz

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

// EventReader is an interface to read single events
type EventReader interface {
	ReadEvent() (Event, error)
	Dial() error
	Close() error
}

// SensorEventReader uses an EventReader (for example WsReader) to provide a channel of SensorEvent
// The SensorEventReader handles connection losses and reconnection attempts of the underlying EventReader
type SensorEventReader struct {
	reader  EventReader
	running bool
}

// Start starts a go routine that reads events from the associated EventReader
// It returns the channel to retrieve events from
func (r *SensorEventReader) Start() (<-chan *SensorEvent, error) {

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
		for r.running {
			// establish connection
			for r.running {
				err := r.reader.Dial()
				if err != nil {
					log.Errorf("Error connecting deCONZ websocket: %s\nAttempting reconnect in 5s...", err)
					time.Sleep(10 * time.Second)
				} else {
					log.Infof("deCONZ websocket connected")
					break
				}
			}
			// read events until connection fails
			for r.running {
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
		// if not running, close connection and return from goroutine
		err := r.reader.Close()
		if err != nil {
			log.Error("Failed to close websocket", err)
			return
		}
		log.Infof("deCONZ websocket closed")
	}()
	return out, nil
}

// StopReadEvents closes the reader, closing the connection to deCONZ and terminating the goroutine
func (r *SensorEventReader) StopReadEvents() {
	r.running = false
}
