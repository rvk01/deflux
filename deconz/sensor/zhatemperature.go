package sensor

// ZHATemperature represents the state of a temperature sensor
type ZHATemperature struct {
	State
	Temperature int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHATemperature) Fields() map[string]interface{} {
	return map[string]interface{}{
		"temperature": float64(z.Temperature) / 100,
	}
}
