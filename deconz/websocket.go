package deconz

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WsReader holds a deCONZ websocket server connection
// It implements the EventReader interface
type WsReader struct {
	WebsocketAddr string
	sensorInfo    SensorProvider
	conn          *websocket.Conn
}

// Dial connects connects to the deCONZ websocket
// Use ReadEvent to receive events
func (r *WsReader) Dial() error {

	if r.sensorInfo == nil {
		return errors.New("cannot dial without a sensorInfo to lookup events from")
	}

	// connect
	var err error
	r.conn, _, err = websocket.DefaultDialer.Dial(r.WebsocketAddr, nil)
	if err != nil {
		return fmt.Errorf("unable to dail %s: %s", r.WebsocketAddr, err)
	}
	return nil
}

// ReadEvent reads, parses and returns the next event from the websocket
func (r *WsReader) ReadEvent() (Event, error) {

	_, message, err := r.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("event read error: %s", err)
	}

	log.Debugf("recv: %s", message)

	e, err := DecodeEvent(r.sensorInfo, message)
	if err != nil {
		return nil, NewEventError(fmt.Errorf("unable to parse message: %s", err), true)
	}

	return e, nil
}

// Close closes the deCONZ websocket connection
func (r *WsReader) Close() error {
	return r.conn.Close()
}
