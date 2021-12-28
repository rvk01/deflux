package sensor

// Daylight represents a change in daylight
type Daylight struct {
	State
	Daylight bool
	Status   int
}

// Fields returns timeseries data for influxdb
func (z *Daylight) Fields() map[string]interface{} {
	return map[string]interface{}{
		"daylight": z.Daylight,
		"status":   z.Status,
	}
}
