package deconz

import (
	"encoding/json"
	"fmt"
	"github.com/fixje/deflux/config"
	"github.com/fixje/deflux/deconz/sensor"
	"log/slog"
	"net/http"
)

// API represents the deCONZ REST API
type API struct {
	Config config.APIConfig

	// sensorCache is used to look up types of sensors when a new event is received via websocket
	sensorCache *CachingSensorProvider
}

// Sensors returns a map of sensors as received from the deCONZ /sensors endpoint
// The map key is the sensor id.
func (a *API) Sensors() (*sensor.Sensors, error) {

	uri := fmt.Sprintf("%s/%s/sensors", a.Config.Addr, a.Config.APIKey)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s: %s", uri, err)
	}

	defer resp.Body.Close()

	var sensors sensor.Sensors

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&sensors)
	if err != nil {
		return nil, fmt.Errorf("unable to decode deCONZ /sensors response: %s", err)
	}

	for id := range sensors {
		s := sensors[id]

		s.ID = id

		sensors[id] = s
		slog.Debug(fmt.Sprintf("got sensor: %v, state: %v", s, s.StateDef))
	}

	return &sensors, nil
}
