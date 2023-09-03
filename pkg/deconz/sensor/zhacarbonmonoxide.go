package sensor

// ZHACarbonMonoxide represents the state of a carbon monoxide sensor
type ZHACarbonMonoxide struct {
	State
	Carbonmonoxide bool
	Lowbattery     bool
	Tampered       bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHACarbonMonoxide) Fields() map[string]interface{} {
	return mergeFields(z.State.Fields(),
		map[string]interface{}{
			"CO":         z.Carbonmonoxide,
			"lowbattery": z.Lowbattery,
			"tampered":   z.Tampered,
		})
}
