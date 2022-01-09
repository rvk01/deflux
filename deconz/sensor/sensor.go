package sensor

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Sensors is a map of sensors indexed by their id
type Sensors map[int]Sensor

// Provider provides information about sensors
type Provider interface {
	// Sensors provides info about all known sensors
	Sensors() (*Sensors, error)

	// Sensor gets a sensor by id
	Sensor(int) (*Sensor, error)
}

// fielder is an interface that provides fields for InfluxDB
type Fielder interface {
	Fields() map[string]interface{}
}

// Timeseries provides tags and fields for the time series database
type TimeSeries interface {
	Timeseries() (map[string]string, map[string]interface{}, error)
}

// Sensor is a deCONZ sensor
// We only implement required fields for event decoding
type Sensor struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	LastSeen string `json:"lastseen"`

	RawState json.RawMessage `json:"state"`
	StateDef interface{}

	Config Config

	Id int
}

type Config struct {
	// Battery state in percent; not present for all sensors
	Battery uint32 `json:"battery"`
}

// State contains properties that are provided by all sensors
// It is embedded in specific sensors' State
type State struct {
	Lastupdated string
}

// EmptyState is an empty struct used to indicate no state was parsed
type EmptyState struct{}


// Timeseries returns tags and fields for use in InfluxDB
func (s *Sensor) Timeseries() (map[string]string, map[string]interface{}, error) {
	f, ok := s.StateDef.(Fielder)
	if !ok {
		return nil, nil, fmt.Errorf("this sensor (%T:%s) has no time series data", s.StateDef, s.Name)
	}

	return map[string]string{"name": s.Name, "type": s.Type, "id": strconv.Itoa(s.Id), "source": "rest"}, f.Fields(), nil
}

// DecodeSensorState tries to unmarshal the appropriate state based
// on the given sensor type
func DecodeSensorState(rawState json.RawMessage, sensorType string) (interface{}, error) {

	var err error

	switch sensorType {
	case "CLIPPresence":
		var s CLIPPresence
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "Daylight":
		var s Daylight
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAAirQuality":
		var s ZHAAirQuality
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHABattery":
		var s ZHABattery
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHACarbonMonoxide":
		var s ZHACarbonMonoxide
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAConsumption":
		var s ZHAConsumption
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAFire":
		var s ZHAFire
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAHumidity":
		var s ZHAHumidity
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHALightLevel":
		var s ZHALightLevel
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAOpenClose":
		var s ZHAOpenClose
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAPower":
		var s ZHAPower
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAPresence":
		var s ZHAPresence
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAPressure":
		var s ZHAPressure
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHASwitch":
		var s ZHASwitch
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHATemperature":
		var s ZHATemperature
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAVibration":
		var s ZHAVibration
		err = json.Unmarshal(rawState, &s)
		return &s, err

	case "ZHAWater":
		var s ZHAWater
		err = json.Unmarshal(rawState, &s)
		return &s, err
	}

	return nil, fmt.Errorf("%s is not a known sensor type", sensorType)
}
