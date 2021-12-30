package sensor

// CLIPPresence represents the state of a presence sensor
type CLIPPresence struct {
	State
	Presence bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *CLIPPresence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"presence": z.Presence,
	}
}
