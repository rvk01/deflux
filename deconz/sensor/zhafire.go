package sensor

// ZHAFire represents the state of a smoke detector
type ZHAFire struct {
	State
	Fire       bool
	Lowbattery bool
	Tampered   bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAFire) Fields() map[string]interface{} {
	return map[string]interface{}{
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
		"fire":       z.Fire,
	}
}
