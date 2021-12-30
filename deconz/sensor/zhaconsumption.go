package sensor

// ZHAConsumption represents the state of a power consumption sensor
type ZHAConsumption struct {
	State
	Consumption int32
	Power       int32
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAConsumption) Fields() map[string]interface{} {
	return map[string]interface{}{
		"consumption": z.Consumption,
		"power":       z.Power,
	}
}
