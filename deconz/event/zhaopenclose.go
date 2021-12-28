package event

// ZHAVibration represents a Vibration Sensor
type ZHAOpenClose struct {
	State
	Open bool
}

// Fields returns timeseries data for influxdb
func (z *ZHAOpenClose) Fields() map[string]interface{} {
	return map[string]interface{}{
		"open": z.Open,
	}
}
