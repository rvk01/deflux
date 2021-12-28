package event

// ZHAVibration represents a Vibration Sensor
type ZHAVibration struct {
	State
	Vibration bool
}

// Fields returns timeseries data for influxdb
func (z *ZHAVibration) Fields() map[string]interface{} {
	return map[string]interface{}{
		"vibration": z.Vibration,
	}
}
