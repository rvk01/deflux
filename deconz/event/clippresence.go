package event

// CLIPPresence represents a presence Sensor
type CLIPPresence struct {
	State
	Presence bool
}

// Fields returns timeseries data for influxdb
func (z *CLIPPresence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"presence": z.Presence,
	}
}
