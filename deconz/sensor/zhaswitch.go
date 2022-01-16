package sensor

// ZHASwitch represents the state of a button or switch
type ZHASwitch struct {
	State
	Buttonevent int
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHASwitch) Fields() map[string]interface{} {
	return mergeFields(z.State.Fields(),
		map[string]interface{}{
			"buttonevent": z.Buttonevent,
		})
}
