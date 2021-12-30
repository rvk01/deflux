package sensor

// ZHABattery represents the battery state of a device
type ZHABattery struct {
	State
	Battery int16
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHABattery) Fields() map[string]interface{} {
	return map[string]interface{}{
		"battery": z.Battery,
	}
}
