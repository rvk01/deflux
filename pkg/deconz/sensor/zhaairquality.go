package sensor

// ZHAAirQuality represents the state of a an air quality sensor
type ZHAAirQuality struct {
	State

	//  [65, "excellent", 220, "good", 660, "moderate", 10000, "unhealthy", 65535, "out of scale"]
	// see https://github.com/dresden-elektronik/deconz-rest-plugin/blob/master/devices/xiaomi/xiaomi_airmonitor_acn01.json
	Airquality    string
	AirqualityPPB int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAAirQuality) Fields() map[string]interface{} {
	return mergeFields(z.State.Fields(),
		map[string]interface{}{
			"airquality":    z.Airquality,
			"airqualityppb": z.AirqualityPPB,
		})
}
