package deconz

import (
	"encoding/json"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
)

// Event represents a deconz sensor event
type Event struct {
	Type     string          `json:"t"`
	Event    string          `json:"e"`
	Resource string          `json:"r"`
	ID       int             `json:"id,string"`
	RawState json.RawMessage `json:"state"`
	State    interface{}
}

// ParseEvent parses events from bytes
func ParseEvent(si SensorInfoProvider, b []byte) (*Event, error) {
	var e Event
	err := json.Unmarshal(b, &e)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json: %s", err)
	}

	// If there is no state, dont try to parse it
	// TODO: figure out what to do with these
	//       some of them seems to be battery updates
	if e.Resource != "sensors" || len(e.RawState) == 0 {
		e.State = &sensor.EmptyState{}
		return &e, nil
	}

	err = e.ParseState(si)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal state: %s", err)
	}

	return &e, nil
}

// ParseState tries to unmarshal the appropriate state based
// on looking up the id though the SensorInfoProvider
func (e *Event) ParseState(tl SensorInfoProvider) error {

	t, err := tl.LookupType(e.ID)
	if err != nil {
		return fmt.Errorf("unable to lookup event id %d: %s", e.ID, err)
	}

	switch t {
	case "CLIPPresence":
		var s sensor.CLIPPresence
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "Daylight":
		var s sensor.Daylight
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAAirQuality":
		var s sensor.ZHAAirQuality
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHABattery":
		var s sensor.ZHABattery
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHACarbonMonoxide":
		var s sensor.ZHACarbonMonoxide
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAConsumption":
		var s sensor.ZHAConsumption
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAFire":
		var s sensor.ZHAFire
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAHumidity":
		var s sensor.ZHAHumidity
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHALightLevel":
		var s sensor.ZHALightLevel
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAOpenClose":
		var s sensor.ZHAOpenClose
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPower":
		var s sensor.ZHAPower
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPresence":
		var s sensor.ZHAPresence
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPressure":
		var s sensor.ZHAPressure
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHASwitch":
		var s sensor.ZHASwitch
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHATemperature":
		var s sensor.ZHATemperature
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAVibration":
		var s sensor.ZHAVibration
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAWater":
		var s sensor.ZHAWater
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	default:
		err = fmt.Errorf("unable to unmarshal event state: %s is not a known type", t)
	}

	// err should continue to be null if everythings ok
	return err
}
