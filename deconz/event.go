package deconz

import (
	"encoding/json"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
)

type Event interface {
	EventName() string
	Resource() string
	Id() int
	State() interface{}
}

// DeconzEvent is a message received over the deCONZ websocket
// We are only interested in e = 'change' events of resource type r = 'sensor'.
// Thus we don't implement all fields.
// See https://dresden-elektronik.github.io/deconz-rest-doc/endpoints/websocket/#message-fields
type DeconzEvent struct {
	// type should always be 'event'
	Type string `json:"t"`

	Event        string `json:"e"`
	ResourceName string `json:"r"`
	ID           int    `json:"id,string"`

	// only for e = 'changed'
	RawState json.RawMessage `json:"state"`

	// only for e = 'changed'
	StateDef interface{}
}

func (e DeconzEvent) EventName() string {
	return e.Event
}

func (e DeconzEvent) Resource() string {
	return e.ResourceName
}

func (e DeconzEvent) Id() int {
	return e.ID
}

func (e DeconzEvent) State() interface{} {
	return e.StateDef
}

// DecodeEvent parses events from bytes
func DecodeEvent(sp SensorProvider, b []byte) (Event, error) {
	var e DeconzEvent
	err := json.Unmarshal(b, &e)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json: %s", err)
	}

	// We don't decode anything other than sensor events
	// If there is no state, dont try to parse it
	// TODO: figure out what to do with these
	//       some of them seems to be battery updates
	if e.Resource() != "sensors" || len(e.RawState) == 0 {
		e.StateDef = &sensor.EmptyState{}
		return e, nil
	}

	s, err := sp.Sensor(e.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get sensor with id %d: %s", e.ID, err)
	}

	err = e.DecodeState(s.Type)
	if err != nil {
		return nil, fmt.Errorf("unable to decode state: %s", err)
	}

	return SensorEvent{Sensor: s, Event: e}, nil
}

// DecodeState tries to unmarshal the appropriate state based
// on looking up the id though the SensorProvider
func (e *DeconzEvent) DecodeState(t string) error {

	var err error

	switch t {
	case "CLIPPresence":
		var s sensor.CLIPPresence
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "Daylight":
		var s sensor.Daylight
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAAirQuality":
		var s sensor.ZHAAirQuality
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHABattery":
		var s sensor.ZHABattery
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHACarbonMonoxide":
		var s sensor.ZHACarbonMonoxide
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAConsumption":
		var s sensor.ZHAConsumption
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAFire":
		var s sensor.ZHAFire
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAHumidity":
		var s sensor.ZHAHumidity
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHALightLevel":
		var s sensor.ZHALightLevel
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAOpenClose":
		var s sensor.ZHAOpenClose
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAPower":
		var s sensor.ZHAPower
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAPresence":
		var s sensor.ZHAPresence
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAPressure":
		var s sensor.ZHAPressure
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHASwitch":
		var s sensor.ZHASwitch
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHATemperature":
		var s sensor.ZHATemperature
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAVibration":
		var s sensor.ZHAVibration
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	case "ZHAWater":
		var s sensor.ZHAWater
		err = json.Unmarshal(e.RawState, &s)
		e.StateDef = &s
		break
	default:
		err = fmt.Errorf("%s is not a known sensor type", t)
	}

	return err
}
