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
	sensorCache *CachingSensorProvider
}

// Config holds properties of the deCONZ API
type Config struct {
	Addr   string
	APIKey string
	wsAddr string
}

// config is used to parse the things we need from the deCONZ config endpoint
type config struct {
	Websocketport int
}

// Sensors returns a map of sensors as received from the deCONZ /sensors endpoint
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
		return nil, fmt.Errorf("unable to decode deCONZ /sensors response: %s", err)
	}

	return &sensors, nil
}

// discoverWebsocket tries to retrieve the websocket address from the deCONZ REST API
// using the /config endpoint
func (c *Config) discoverWebsocket() error {
	u, err := url.Parse(c.Addr)
	if err != nil {
		return fmt.Errorf("unable to discover websocket while parsing config: %s", err)
	}
	u.Path = path.Join(u.Path, c.APIKey, "config")

	resp, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("unable to discover websocket while getting config: %s", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var conf config
	err = dec.Decode(&conf)
	if err != nil {
		return fmt.Errorf("unable to discover websocket while decoding response: %s", err)
	}

	// change our old parsed url to websocket, it should connect to the websocket endpoint of deCONZ
	u.Scheme = "ws"
	u.Path = "/"
	u.Host = fmt.Sprintf("%s:%d", u.Hostname(), conf.Websocketport)

	c.wsAddr = u.String()
	return nil
}
