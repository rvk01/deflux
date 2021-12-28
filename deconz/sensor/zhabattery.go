package sensor

// ZHABattery represents the battery state of a device
type ZHABattery struct {
	State
	Battery int16
}

// Fields returns timeseries data for influxdb
func (z *ZHABattery) Fields() map[string]interface{} {
	return map[string]interface{}{
		"battery": z.Battery,
	}
}
