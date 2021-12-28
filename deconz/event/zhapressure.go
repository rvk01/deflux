package event

// ZHAPressure represents a presure change
type ZHAPressure struct {
	State
	Pressure int
}

// Fields returns timeseries data for influxdb
func (z *ZHAPressure) Fields() map[string]interface{} {
	return map[string]interface{}{
		"pressure": z.Pressure,
	}
}
