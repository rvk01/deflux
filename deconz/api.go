package deconz

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fixje/deflux/deconz/event"
)

// API represents the deCONZ rest api
// It implements the SensorRepository interface
type API struct {
	Config Config

	// sensorCache is used to look up types of sensors when a new event is received via websocket
	sensorCache *CachedSensorStore
}

// Sensors returns a map of sensors
func (a *API) Sensors() (*Sensors, error) {

	url := fmt.Sprintf("%s/%s/sensors", a.Config.Addr, a.Config.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s: %s", url, err)
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

// WsReader returns an event.WsReader with a default CachedSensorStore
func (a *API) WsReader() (*event.WsReader, error) {

	if a.sensorCache == nil {
		a.sensorCache = &CachedSensorStore{SensorRepository: a}
	}

	if a.Config.wsAddr == "" {
		err := a.Config.discoverWebsocket()
		if err != nil {
			return nil, err
		}
	}

	return &event.WsReader{TypeRepository: a.sensorCache, WebsocketAddr: a.Config.wsAddr}, nil
}

// SensorEventReader takes an event reader and returns a SensorEventReader
func (a *API) SensorEventReader(r *event.WsReader) *SensorEventReader {

	if a.sensorCache == nil {
		a.sensorCache = &CachedSensorStore{SensorRepository: a}
	}

	return &SensorEventReader{lookup: a.sensorCache, reader: r}
}
