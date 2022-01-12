package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// ApiConfig holds properties of the deCONZ API
type ApiConfig struct {
	Addr   string
	APIKey string
	WsAddr string
}

// config is used to parse the things we need from the deCONZ config endpoint
type config struct {
	Websocketport int
}

// DiscoverWebsocket tries to retrieve the websocket address from the deCONZ REST API
// using the /config endpoint
func (c *ApiConfig) DiscoverWebsocket() error {
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

	// change our old parsed URL to websocket, it should connect to the websocket endpoint of deCONZ
	u.Scheme = "ws"
	u.Path = "/"
	u.Host = fmt.Sprintf("%s:%d", u.Hostname(), conf.Websocketport)

	c.WsAddr = u.String()
	return nil
}
