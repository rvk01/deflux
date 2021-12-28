package sensor

// ZHACarbonMonoxide represents a CarbonMonoxide Sensor
type ZHACarbonMonoxide struct {
	State
	Carbonmonoxide bool
	Lowbattery     bool
	Tampered       bool
}

// Fields returns timeseries data for influxdb
func (z *ZHACarbonMonoxide) Fields() map[string]interface{} {
	return map[string]interface{}{
		"CO":         z.Carbonmonoxide,
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
	}
}
