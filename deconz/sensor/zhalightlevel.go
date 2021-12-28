package sensor

// ZHALightLevel represents a LightLevel Sensor
type ZHALightLevel struct {
	State
	Dark       bool
	Daylight   bool
	LightLevel int32
	Lux        int16
}

// Fields returns timeseries data for influxdb
func (z *ZHALightLevel) Fields() map[string]interface{} {
	return map[string]interface{}{
		"daylight":   z.Daylight,
		"dark":       z.Dark,
		"lightlevel": z.LightLevel,
		"lux":        z.Lux,
	}
}
