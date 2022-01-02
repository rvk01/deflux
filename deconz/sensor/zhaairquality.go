package sensor

// ZHAAirQuality represents the state of a an air quality sensor
type ZHAAirQuality struct {
	State
	Airquality    string
	AirqualityPPB uint16
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAAirQuality) Fields() map[string]interface{} {
	return map[string]interface{}{
		"airquality":    z.Airquality,
		"airqualityppb": z.AirqualityPPB,
	}
}
