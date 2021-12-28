package event

// ZHAWater respresents a change from a flood sensor
type ZHAWater struct {
	State
	Lowbattery bool
	Tampered   bool
	Water      bool
}

// Fields returns timeseries data for influxdb
func (z *ZHAWater) Fields() map[string]interface{} {
	return map[string]interface{}{
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
		"water":      z.Water,
	}
}
