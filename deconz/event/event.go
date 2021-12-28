package event

import (
	"encoding/json"
	"fmt"
)

// TypeLookuper is the interface that we require to lookup types from id's
type TypeLookuper interface {
	LookupType(int) (string, error)
}

// Event represents a deconz sensor event
type Event struct {
	Type     string          `json:"t"`
	Event    string          `json:"e"`
	Resource string          `json:"r"`
	ID       int             `json:"id,string"`
	RawState json.RawMessage `json:"state"`
	State    interface{}
}

// Decoder is able to decode deCONZ events
type Decoder struct {
	TypeStore TypeLookuper
}

// Parse parses events from bytes
func (d *Decoder) Parse(b []byte) (*Event, error) {
	var e Event
	err := json.Unmarshal(b, &e)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json: %s", err)
	}

	// If there is no state, dont try to parse it
	// TODO: figure out what to do with these
	//       some of them seems to be battery updates
	if e.Resource != "sensors" || len(e.RawState) == 0 {
		e.State = &EmptyState{}
		return &e, nil
	}

	err = e.ParseState(d.TypeStore)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal state: %s", err)
	}

	return &e, nil
}

// ParseState tries to unmarshal the appropriate state based
// on looking up the id though the TypeStore
func (e *Event) ParseState(tl TypeLookuper) error {

	t, err := tl.LookupType(e.ID)
	if err != nil {
		return fmt.Errorf("unable to lookup event id %d: %s", e.ID, err)
	}

	switch t {
	case "ZHAAirQuality":
		var s ZHAAirQuality
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHABattery":
		var s ZHABattery
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAConsumption":
		var s ZHAConsumption
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAFire":
		var s ZHAFire
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPower":
		var s ZHAPower
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHATemperature":
		var s ZHATemperature
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPressure":
		var s ZHAPressure
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAHumidity":
		var s ZHAHumidity
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAWater":
		var s ZHAWater
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHASwitch":
		var s ZHASwitch
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "Daylight":
		var s Daylight
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAPresence":
		var s ZHAPresence
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "CLIPPresence":
		var s CLIPPresence
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHALightLevel":
		var s ZHALightLevel
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAVibration":
		var s ZHAVibration
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHAOpenClose":
		var s ZHAOpenClose
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	case "ZHACarbonMonoxide":
		var s ZHACarbonMonoxide
		err = json.Unmarshal(e.RawState, &s)
		e.State = &s
		break
	default:
		err = fmt.Errorf("unable to unmarshal event state: %s is not a known type", t)
	}

	// err should continue to be null if everythings ok
	return err
}

// State is for embedding into event states
type State struct {
	Lastupdated string
}

// CLIPPresence represents a presence Sensor
type CLIPPresence struct {
	State
	Presence bool
}

// Fields returns timeseries data for influxdb
func (z *CLIPPresence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"presence": z.Presence,
	}
}

// EmptyState is an empty struct used to indicate no state was parsed
type EmptyState struct{}
