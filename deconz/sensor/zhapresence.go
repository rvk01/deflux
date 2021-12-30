package sensor

// ZHAPresence represents the state of a presence Sensor
type ZHAPresence struct {
	State
	Presence bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAPresence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"presence": z.Presence,
	}
}
