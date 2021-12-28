package event

// ZHAPower represents a Power Sensor
type ZHAPower struct {
	State
	Current int32
	Power   int32
	Voltage int16
}

// Fields returns timeseries data for influxdb
func (z *ZHAPower) Fields() map[string]interface{} {
	return map[string]interface{}{
		"current": z.Current,
		"power":   z.Power,
		"voltage": z.Voltage,
	}
}
