package sensor

// ZHAVibration represents the state of a vibration sensor
type ZHAVibration struct {
	State
	Vibration bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAVibration) Fields() map[string]interface{} {
	return map[string]interface{}{
		"vibration": z.Vibration,
	}
}
