package deconz

import (
	"errors"
	"github.com/fixje/deflux/deconz/event"
	log "github.com/sirupsen/logrus"
	"time"
)

// SensorLookup represents an interface for sensor lookup
type SensorLookup interface {
	LookupSensor(int) (*Sensor, error)
}

// EventReader interface
type EventReader interface {
	ReadEvent() (*event.Event, error)
	Dial() error
	Close() error
}

// SensorEventReader reads events from an event.WsReader and returns SensorEvents
type SensorEventReader struct {
	lookup  SensorLookup
	reader  EventReader
	running bool
}

// Start starts a thread reading events
// returns the channel to retrieve events from
func (r *SensorEventReader) Start() (<-chan *SensorEvent, error) {

	out := make(chan *SensorEvent)

	if r.lookup == nil {
		return nil, errors.New("Cannot run without a SensorLookup from which to lookup sensors")
	}
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
					log.Errorf("Error connecting Deconz websocket: %s\nAttempting reconnect in 5s...", err)
					time.Sleep(5 * time.Second) // TODO configurable delay
				} else {
					log.Infof("Deconz websocket connected")
					break
				}
			}
			// read events until connection fails
			for r.running {
				e, err := r.reader.ReadEvent()
				if err != nil {
					if eerr, ok := err.(event.EventError); ok && eerr.Recoverable() {
						log.Errorf("Dropping event due to error: %s", err)
						continue
					}
					continue REDIAL
				}
				// we only care about sensor events
				if e.Resource != "sensors" {
					log.Debugf("Dropping non-sensor event type %s", e.Resource)
					continue
				}

				sensor, err := r.lookup.LookupSensor(e.ID)
				if err != nil {
					log.Warningf("Dropping event. Could not lookup sensor for id %d: %s", e.ID, err)
					continue
				}
				// send event on channel
				out <- &SensorEvent{Event: e, Sensor: sensor}
			}
		}
		// if not running, close connection and return from goroutine
		err := r.reader.Close()
		if err != nil {
			log.Error("Failed to close websocket", err)
			return
		}
		log.Infof("Deconz websocket closed")
	}()
	return out, nil
}

// StopReadEvents closes the reader, closing the connection to deCONZ and terminating the goroutine
func (r *SensorEventReader) StopReadEvents() {
	r.running = false
}
