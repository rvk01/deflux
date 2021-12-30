package sensor

// Daylight represents the state of a daylight sensor
type Daylight struct {
	State
	Daylight bool
	Status   int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *Daylight) Fields() map[string]interface{} {
	return map[string]interface{}{
		"daylight": z.Daylight,
		"status":   z.Status,
	}
}
