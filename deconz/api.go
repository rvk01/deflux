package deconz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// API represents the deCONZ REST API
type API struct {
	Config Config

	// sensorCache is used to look up types of sensors when a new event is received via websocket
	sensorCache *CachingSensorStore
}

// Config holds properties of a deCONZ gateway
type Config struct {
	Addr   string
	APIKey string
	wsAddr string
}

// config is used to parse the things we need from the deCONZ config endpoint
type config struct {
	Websocketport int
}

// Sensors returns a map of sensors
func (a *API) Sensors() (*Sensors, error) {

	uri := fmt.Sprintf("%s/%s/sensors", a.Config.Addr, a.Config.APIKey)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s: %s", uri, err)
	}

	defer resp.Body.Close()

	var sensors Sensors

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&sensors)
	if err != nil {
		return nil, fmt.Errorf("unable to decode deCONZ response: %s", err)
	}

	return &sensors, nil

}

// CreateWsReader creates a WsReader which consumes events from the deCONZ websocket interface
func CreateWsReader(api API, si SensorInfoProvider) (*WsReader, error) {

	if api.Config.wsAddr == "" {
		err := api.Config.discoverWebsocket()
		if err != nil {
			return nil, err
		}
	}

	return &WsReader{sensorInfo: si, WebsocketAddr: api.Config.wsAddr}, nil
}

// CreateSensorEventReader creates a new SensorEventReader that continuously reads events from the given WsReader
func CreateSensorEventReader(r *WsReader) *SensorEventReader {
	return &SensorEventReader{sensorProvider: r.sensorInfo, reader: r}
}

// discoverWebsocket tries to retrieve the websocket address from the deCONZ REST API
// using the /config endpoint
func (c *Config) discoverWebsocket() error {
	u, err := url.Parse(c.Addr)
	if err != nil {
		return fmt.Errorf("unable to discover websocket: %s", err)
	}
	u.Path = path.Join(u.Path, c.APIKey, "config")

	resp, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("unable to discover websocket: %s", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var conf config
	err = dec.Decode(&conf)
	if err != nil {
		return fmt.Errorf("unable to discover websocket: %s", err)
	}

	// change our old parsed url to websocket, it should connect to the websocket endpoint of deCONZ
	u.Scheme = "ws"
	u.Path = "/"
	u.Host = fmt.Sprintf("%s:%d", u.Hostname(), conf.Websocketport)

	c.wsAddr = u.String()
	return nil
}
