package sensor

// ZHAPressure represents the state of a pressure sensor
type ZHAPressure struct {
	State
	Pressure int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAPressure) Fields() map[string]interface{} {
	return map[string]interface{}{
		"pressure": z.Pressure,
	}
}
