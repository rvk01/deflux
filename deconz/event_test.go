package deconz

import (
	"errors"
	"github.com/fixje/deflux/deconz/sensor"
	"os"
	"reflect"
	"testing"
)

type TestSensorProvider struct {
	Store *sensor.Sensors
}

func (l TestSensorProvider) Sensor(i int) (*sensor.Sensor, error) {
	if s, ok := (*l.Store)[i]; ok {
		return &s, nil
	}
	return nil, errors.New("not found")
}

func (l TestSensorProvider) Sensors() (*sensor.Sensors, error) {
	return l.Store, nil
}

var sensorInfo sensor.Provider

func TestMain(m *testing.M) {

	sensorInfo = TestSensorProvider{Store: &sensor.Sensors{
		1:  sensor.Sensor{Type: "ZHATemperature", Name: "ZHATemperature"},
		2:  sensor.Sensor{Type: "ZHAHumidity", Name: "ZHAHumidity"},
		3:  sensor.Sensor{Type: "ZHAPressure", Name: "ZHAPressure"},
		5:  sensor.Sensor{Type: "ZHAFire", Name: "ZHAFire"},
		6:  sensor.Sensor{Type: "ZHAWater", Name: "ZHAWater"},
		7:  sensor.Sensor{Type: "ZHASwitch", Name: "ZHASwitch"},
		8:  sensor.Sensor{Type: "ZHAOpenClose", Name: "ZHAOpenClose"},
		9:  sensor.Sensor{Type: "ZHAOpenClose", Name: "ZHAOpenClose"},
		10: sensor.Sensor{Type: "ZHABattery", Name: "ZHABattery"},
		11: sensor.Sensor{Type: "ZHAConsumption", Name: "ZHAConsumption"},
		12: sensor.Sensor{Type: "CLIPPresence", Name: "CLIPPresence"},
		13: sensor.Sensor{Type: "ZHAPower", Name: "ZHAPower"},
		14: sensor.Sensor{Type: "ZHALightLevel", Name: "ZHALightLevel"},
		15: sensor.Sensor{Type: "ZHAAirQuality", Name: "ZHAAirQuality"},
	}}

	os.Exit(m.Run())
}

var sensorTests = map[string]struct {
	jsonInput string
	want      interface{}
}{
	// Xiaomi smoke detector
	"ZHAFire": {
		jsonInput: `{
			"e": "changed",
			"id": "5",
			"r": "sensors",
			"state": {
			"fire": false,
				"lastupdated": "2018-03-13T19:46:03",
				"lowbattery": false,
				"tampered": false
			},
			"t": "event"
		}`,
		want: &sensor.ZHAFire{
			State:      sensor.State{Lastupdated: "2018-03-13T19:46:03"},
			Fire:       false,
			Lowbattery: false,
			Tampered:   false,
		},
	},

	// Xiaomi flood detector
	"ZHAWater": {
		jsonInput: `{ 
			"e": "changed", 
			"id": "6",
			"r": "sensors",
			"state": {
				"lastupdated": "2018-03-13T20:46:03",
				"lowbattery": false,
				"tampered": false,
				"water": true
			},
			"t": "event"
		}`,
		want: &sensor.ZHAWater{
			State:      sensor.State{Lastupdated: "2018-03-13T20:46:03"},
			Water:      true,
			Lowbattery: false,
			Tampered:   false,
		},
	},

	// Examples of the Xiaomi temperature/pressure/humidity sensor
	"ZHAPressure": {
		jsonInput: `{
			"e": "changed",
			"id": "3",
			"r": "sensors",
			"state": {
				"lastupdated": "2018-03-08T19:35:24",
				"pressure": 993
			},
			"t": "event"
		}`,
		want: &sensor.ZHAPressure{
			State:    sensor.State{Lastupdated: "2018-03-08T19:35:24"},
			Pressure: 993,
		},
	},
	"ZHATemperature": {
		jsonInput: `{
			"e": "changed",
			"id": "1",
			"r": "sensors",
			"state": {
				"lastupdated": "2018-03-08T19:35:24",
				"temperature": 2062
			},
			"t": "event"
		}`,
		want: &sensor.ZHATemperature{
			State:       sensor.State{Lastupdated: "2018-03-08T19:35:24"},
			Temperature: 2062,
		},
	},
	"ZHAHumidity": {
		jsonInput: `{
			"e": "changed",
			"id": "2",
			"r": "sensors",
			"state": {
				"humidity": 2985,
				"lastupdated": "2018-03-08T19:35:24"
			},
			"t": "event"
		}`,
		want: &sensor.ZHAHumidity{
			State:    sensor.State{Lastupdated: "2018-03-08T19:35:24"},
			Humidity: 2985,
		},
	},

	// Xiaomi random switch "sensor"
	"ZHASwitch": {
		jsonInput: `{
			"e": "changed",
			"id": "7",
			"r": "sensors",
			"state": {
				"buttonevent": 1000,
				"lastupdated": "2018-03-20T20:52:18"
			},
			"t": "event"
		}`,
		want: &sensor.ZHASwitch{
			State:       sensor.State{Lastupdated: "2018-03-20T20:52:18"},
			Buttonevent: 1000,
		},
	},

	// State of the following events was retrieved via the /sensors REST endpoint
	// The rest of the messages are made up
	"ZHAOpenClose with tamper": {
		jsonInput: `{
			"e": "changed",
			"id": "8",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-01T12:39:38.370",
				"lowbattery": true,
				"open": true,
				"tampered": false
			}
		}`,
		want: &sensor.ZHAOpenClose{
			State:      sensor.State{Lastupdated: "2022-01-01T12:39:38.370"},
			Lowbattery: true,
			Open:       true,
			Tampered:   false,
		},
	},
	"ZHAOpenClose without tamper": {
		jsonInput: `{
			"e": "changed",
			"id": "9",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-04T05:57:50.067",
				"open": true
			}
		}`,
		want: &sensor.ZHAOpenClose{
			State: sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Open:  true,
		},
	},
	"ZHABattery": {
		jsonInput: `{
			"e": "changed",
			"id": "10",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-04T05:57:50.067",
				"battery": 77
			}
		}`,
		want: &sensor.ZHABattery{
			State:   sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Battery: 77,
		},
	},
	"ZHAConsumption": {
		jsonInput: `{
			"e": "changed",
			"id": "11",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-04T05:57:50.067",
				"consumption": 8,
				"power": 0
			}
		}`,
		want: &sensor.ZHAConsumption{
			State:       sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Consumption: 8,
			Power:       0,
		},
	},
	"CIPPresence": {
		jsonInput: `{
			"e": "changed",
			"id": "12",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-04T05:57:50.067",
				"presence": true
			}
		}`,
		want: &sensor.CLIPPresence{
			State:    sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Presence: true,
		},
	},
	"ZHAPower": {
		jsonInput: `{
			"e": "changed",
			"id": "13",
			"r": "sensors",
			"t": "event",
			"state": {
				"lastupdated": "2022-01-04T05:57:50.067",
				"current": 0,
				"power": 0,
				"voltage": 236
			}
		}`,
		want: &sensor.ZHAPower{
			State:   sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Power:   0,
			Current: 0,
			Voltage: 236,
		},
	},
	"ZHALightLevel": {
		jsonInput: `{
			"e": "changed",
			"id": "14",
			"r": "sensors",
			"t": "event",
			"state": {
				"dark": true,
				"daylight": false,
				"lastupdated": "2022-01-04T05:57:50.067",
				"lightlevel": 4772,
				"lux": 3
			}
		}`,
		want: &sensor.ZHALightLevel{
			State:      sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			LightLevel: 4772,
			Dark:       true,
			Daylight:   false,
			Lux:        3,
		},
	},
	"ZHAAirQuality": {
		jsonInput: `{
			"e": "changed",
			"id": "15",
			"r": "sensors",
			"t": "event",
			"state": {
				"airquality": "good",
				"airqualityppb": 79,
				"lastupdated": "2022-01-04T05:57:50.067"
			}
		}`,
		want: &sensor.ZHAAirQuality{
			State:         sensor.State{Lastupdated: "2022-01-04T05:57:50.067"},
			Airquality:    "good",
			AirqualityPPB: 79,
		},
	},
}

func TestSensors(t *testing.T) {
	for name, tc := range sensorTests {
		t.Run(name, func(t *testing.T) {
			got := decodeTest(t, name, tc.jsonInput)
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func decodeTest(t *testing.T, name string, input string) interface{} {
	result, err := DecodeEvent(sensorInfo, []byte(input))
	if err != nil {
		t.Logf("unable to unmarshal %s: %s", name, err)
		t.FailNow()
	}

	return result.State()
}
