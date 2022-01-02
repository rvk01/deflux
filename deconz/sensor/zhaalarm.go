package sensor

// ZHAAlarm represents the state of an alarm sensor
type ZHAAlarm struct {
	State
	Lowbattery bool
	Tampered   bool
	Alarm      bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAAlarm) Fields() map[string]interface{} {
	return map[string]interface{}{
		"lowbattery": z.Lowbattery,
		"tampered":   z.Tampered,
		"alarm":      z.Alarm,
	}
}
