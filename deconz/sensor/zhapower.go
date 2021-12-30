package sensor

// ZHAPower represents the state of a power sensor
type ZHAPower struct {
	State
	Current int32
	Power   int32
	Voltage int16
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAPower) Fields() map[string]interface{} {
	return map[string]interface{}{
		"current": z.Current,
		"power":   z.Power,
		"voltage": z.Voltage,
	}
}
