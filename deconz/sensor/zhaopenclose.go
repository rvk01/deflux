package sensor

// ZHAOpenClose represents the state of an open/close sensor
type ZHAOpenClose struct {
	State
	Open bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAOpenClose) Fields() map[string]interface{} {
	return map[string]interface{}{
		"open": z.Open,
	}
}
