package deconz

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"time"
)

// WebsocketEventReader uses an EventReader (for example WsReader) to provide a channel of SensorEvent
// The WebsocketEventReader handles connection losses and reconnection attempts of the underlying EventReader
type WebsocketEventReader struct {
	WebsocketAddr  string
	SensorProvider sensor.Provider

	conn    *websocket.Conn
	connCtx ctx.Context
	running bool
}

// NewWebsocketEventReader creates a new WebsocketEventReader that continuously reads events from the deCONZ websocket
// It uses the API to discover the websocket address
// The structure of the JSON messages from the websocket depend on the resource/sensor type. Thus,
// the WsReader requires a sensor.Provider to properly unmarshal those messages.
func NewWebsocketEventReader(api API, si sensor.Provider) (*WebsocketEventReader, error) {
	if api.Config.WsAddr == "" {
		err := api.Config.DiscoverWebsocket()
		if err != nil {
			return nil, err
		}
	}

	return &WebsocketEventReader{
		WebsocketAddr:  api.Config.WsAddr,
		SensorProvider: si,
	}, nil
}

// Start starts a go routine that reads events from the associated EventReader
// It returns the channel to retrieve events from
func (r *WebsocketEventReader) Start(ctx ctx.Context) (<-chan *SensorEvent, error) {

	out := make(chan *SensorEvent)

	if r.running {
		return nil, errors.New("WebsocketEventReader is already running")
	}

	r.running = true
	r.connCtx = ctx

	go func() {
		for {
			select {
			case <-r.connCtx.Done():
				log.Debug("Aborting websocket connection")
				close(out)
				return

			default:
				// read events until connection fails
				e, err := r.readEvent()
				if err != nil {
					if err, ok := err.(EventError); ok && err.Recoverable() {
						log.Errorf("Dropping event due to error: %s", err)
						continue
					}
				}

				// nil event returned, e.g. due to disconnect
				if e == nil {
					continue
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

	}()

	return out, nil
}

// connect connects the EventReader
// If the connection fails, it retries every 10s as long as the given Context is not Done()
func (r *WebsocketEventReader) connect() {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		select {
		case <-r.connCtx.Done():
			log.Debug("Aborting websocket connection")
		default:

			if r.SensorProvider == nil {
				panic("cannot dial without a sensor.Provider")
			}

			if r.conn != nil {
				return
			}

			// connect
			var err error
			r.conn, _, err = websocket.DefaultDialer.DialContext(r.connCtx, r.WebsocketAddr, nil)

			if err != nil {
				log.Errorf("Error connecting deCONZ websocket: %s\nAttempting reconnect in 10s...", err)
			} else {
				log.Infof("deCONZ websocket connected")
				return
			}
		}
	}
}

// readEvent reads, parses and returns the next event from the websocket
func (r *WebsocketEventReader) readEvent() (Event, error) {

	if r.conn == nil {
		r.connect()
	}

	_, message, err := r.conn.ReadMessage()
	if err != nil {
		r.conn = nil

		return nil, fmt.Errorf("event read error: %s", err)
	}

	log.Debugf("recv: %s", message)

	e, err := DecodeEvent(r.SensorProvider, message)
	if err != nil {
		return nil, NewEventError(fmt.Errorf("unable to parse message: %s", err), true)
	}

	return e, nil
}

// Shutdown closes the reader, closing the connection to deCONZ
// The method blocks until all background tasks are terminated or the given Context is aborted
func (r *WebsocketEventReader) Shutdown(ctx ctx.Context) {
	r.running = false
	done := make(chan interface{}, 1)

	go func() {

		if r.conn != nil {
			err := r.conn.Close()
			if err != nil {
				log.Error("Failed to close websocket", err)
				return
			}
			log.Infof("deCONZ websocket closed")
		}

		close(done)
		return
	}()

	select {
	case <-ctx.Done():
		return
	case <-done:
		return
	}
}
