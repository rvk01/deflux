package sensor

// ZHAPresence represents a presence Sensor
type ZHAPresence struct {
	State
	Presence bool
}

// Fields returns timeseries data for influxdb
func (z *ZHAPresence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"presence": z.Presence,
	}
}
