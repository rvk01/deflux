package sensor

// ZHAWater represents the state of a flood sensor
type ZHAWater struct {
	State
	Lowbattery bool
	Tampered   bool
	Water      bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAWater) Fields() map[string]interface{} {
	return map[string]interface{}{
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
		"water":      z.Water,
	}
}
