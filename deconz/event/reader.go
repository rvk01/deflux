package event

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WsReader holds a deCONZ websocket server connection
// It implements the deconz.EventReader interface
type WsReader struct {
	WebsocketAddr  string
	TypeRepository TypeRepository
	decoder        *Decoder
	conn           *websocket.Conn
}

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

// Dial connects connects to the deCONZ websocket, use ReadEvent to receive events
func (r *WsReader) Dial() error {

	if r.TypeRepository == nil {
		return errors.New("cannot dial without a TypeRepository to lookup events from")
	}

	// create a decoder with the typestore
	r.decoder = &Decoder{TypeRepository: r.TypeRepository}

	// connect
	var err error
	r.conn, _, err = websocket.DefaultDialer.Dial(r.WebsocketAddr, nil)
	if err != nil {
		return fmt.Errorf("unable to dail %s: %s", r.WebsocketAddr, err)
	}
	return nil
}

// ReadEvent reads, parses and returns the next event from the websocket
func (r *WsReader) ReadEvent() (*Event, error) {

	_, message, err := r.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("event read error: %s", err)
	}

	log.Debugf("recv: %s", message)

	e, err := r.decoder.Parse(message)
	if err != nil {
		return nil, NewEventError(fmt.Errorf("unable to parse message: %s", err), true)
	}

	return e, nil
}

// Close closes the connection to deconz
func (r *WsReader) Close() error {
	return r.conn.Close()
}
