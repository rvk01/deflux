package sensor

// ZHAPresence represents the state of a presence Sensor
type ZHAPresence struct {
	State
	Presence   bool
	Lowbattery bool
	Tampered   bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAPresence) Fields() map[string]interface{} {
	return mergeFields(z.State.Fields(),
		map[string]interface{}{
			"lowbattery": z.Lowbattery,
			"tampered":   z.Tampered,
			"presence":   z.Presence,
		})
}
