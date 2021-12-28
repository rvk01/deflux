package sensor

// ZHATemperature represents a temperature change
type ZHATemperature struct {
	State
	Temperature int
}

// Fields returns timeseries data for influxdb
func (z *ZHATemperature) Fields() map[string]interface{} {
	return map[string]interface{}{
		"temperature": float64(z.Temperature) / 100,
	}
}
