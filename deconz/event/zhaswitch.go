package event

// ZHASwitch represents a change from a button or switch
type ZHASwitch struct {
	State
	Buttonevent int
}

// Fields returns timeseries data for influxdb
func (z *ZHASwitch) Fields() map[string]interface{} {
	return map[string]interface{}{
		"buttonevent": z.Buttonevent,
	}
}
