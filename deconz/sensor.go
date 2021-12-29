package deconz

import (
	"fmt"
	"strconv"
)

// Sensors is a map of sensors indexed by their id
type Sensors map[int]Sensor

// SensorProvider provides information about sensors
type SensorProvider interface {
	// Sensors provides info about all known sensors
	Sensors() (*Sensors, error)

	// Sensor gets a sensor by id
	Sensor(int) (*Sensor, error)
}

// Sensor is a deCONZ sensor
// We only implement required fields for event decoding
type Sensor struct {
	Type string
	Name string
}

// SensorEvent is an Event triggered by a Sensor
type SensorEvent struct {
	*Sensor
	Event
}

// fielder is an interface that provides fields for InfluxDB
type fielder interface {
	Fields() map[string]interface{}
}

// Timeseries returns tags and fields for use in InfluxDB
func (s *SensorEvent) Timeseries() (map[string]string, map[string]interface{}, error) {
	f, ok := s.Event.State().(fielder)
	if !ok {
		return nil, nil, fmt.Errorf("this event (%T:%s) has no time series data", s.State, s.Name)
	}

	return map[string]string{"name": s.Name, "type": s.Sensor.Type, "id": strconv.Itoa(s.Event.Id())}, f.Fields(), nil
}
