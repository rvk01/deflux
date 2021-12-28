package event

// ZHAFire represents a change from a smoke detector
type ZHAFire struct {
	State
	Fire       bool
	Lowbattery bool
	Tampered   bool
}

// Fields returns timeseries data for influxdb
func (z *ZHAFire) Fields() map[string]interface{} {
	return map[string]interface{}{
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
		"fire":       z.Fire,
	}
}
