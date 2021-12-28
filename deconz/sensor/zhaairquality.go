package sensor

// ZHAAirQuality represents a Air Quality Sensor
type ZHAAirQuality struct {
	State
	Airquality    string
	AirqualityPPB int32
}

// Fields returns timeseries data for influxdb
func (z *ZHAAirQuality) Fields() map[string]interface{} {
	return map[string]interface{}{
		"airquality":    z.Airquality,
		"airqualityppb": z.AirqualityPPB,
	}
}
