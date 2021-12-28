package deconz

import (
	"fmt"
	"strconv"
)

// Sensors is a map of sensors indexed by their id
type Sensors map[int]Sensor

// SensorInfoProvider provides information about sensors
type SensorInfoProvider interface {
	Sensors() (*Sensors, error)

	// LookupSensor gets a sensor by id
	LookupSensor(int) (*Sensor, error)

	// LookupType returns the type for a given sensor id
	LookupType(int) (string, error)
}

// Sensor is a deCONZ sensor, not that we only implement fields needed
// for event parsing to work
type Sensor struct {
	Type string
	Name string
}

// SensorEvent is a sensor with an additional
type SensorEvent struct {
	*Sensor
	*Event
}

// fielder is an interface that provides fields for InfluxDB
type fielder interface {
	Fields() map[string]interface{}
}

// Timeseries returns tags and fields for use in InfluxDB
func (s *SensorEvent) Timeseries() (map[string]string, map[string]interface{}, error) {
	f, ok := s.Event.State.(fielder)
	if !ok {
		return nil, nil, fmt.Errorf("this event (%T:%s) has no time series data", s.State, s.Name)
	}

	return map[string]string{"name": s.Name, "type": s.Sensor.Type, "id": strconv.Itoa(s.Event.ID)}, f.Fields(), nil
}
