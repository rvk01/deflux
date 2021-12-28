package event

// ZHAHumidity represents a presure change
type ZHAHumidity struct {
	State
	Humidity int
}

// Fields returns timeseries data for influxdb
func (z *ZHAHumidity) Fields() map[string]interface{} {
	return map[string]interface{}{
		"humidity": float64(z.Humidity) / 100,
	}
}
