package deconz

import (
	"fmt"
	"github.com/fixje/deflux/deconz/event"
	"strconv"
)

// Sensors is a map of sensors indexed by their id
type Sensors map[int]Sensor

// SensorRepository provides sensor information
type SensorRepository interface {
	Sensors() (*Sensors, error)
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
	*event.Event
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
