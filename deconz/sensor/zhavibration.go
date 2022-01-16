package sensor

// ZHAVibration represents the state of a vibration sensor
// TODO not sure if int or float: orientation, tiltangle, vibrationstrength
type ZHAVibration struct {
	State
	Vibration bool
}

// Fields implements the fielder interface and returns time series data for InfluxDB
func (z *ZHAVibration) Fields() map[string]interface{} {
	return mergeFields(z.State.Fields(),
		map[string]interface{}{
			"vibration": z.Vibration,
		})
}
