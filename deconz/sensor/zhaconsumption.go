package sensor

// ZHAConsumption represents a power consumption sensor
type ZHAConsumption struct {
	State
	Consumption int32
	Power       int32
}

// Fields returns timeseries data for influxdb
func (z *ZHAConsumption) Fields() map[string]interface{} {
	return map[string]interface{}{
		"consumption": z.Consumption,
		"power":       z.Power,
	}
}
