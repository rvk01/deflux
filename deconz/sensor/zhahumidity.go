package sensor

// ZHAHumidity represents the state of a humidity sensor
type ZHAHumidity struct {
	State
	Humidity int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAHumidity) Fields() map[string]interface{} {
	return map[string]interface{}{
		"humidity": float64(z.Humidity) / 100,
	}
}
