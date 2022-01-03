package deconz

import (
	"errors"
	"github.com/fixje/deflux/deconz/sensor"
	"os"
	"testing"
)

// examples from the xiaomi temp/hum/pressure sensor
const temperatureEventPayload = `{"e":"changed","id":"1","r":"sensors","state":{"lastupdated":"2018-03-08T19:35:24","temperature":2062},"t":"event"}`
const humidityEventPayload = `{"e":"changed","id":"2","r":"sensors","state":{"humidity":2985,"lastupdated":"2018-03-08T19:35:24"},"t":"event"}`
const pressureEventPayload = `{"e":"changed","id":"3","r":"sensors","state":{"lastupdated":"2018-03-08T19:35:24","pressure":993},"t":"event"}`

// xiaomi smoke detector
const smokeDetectorNoFireEventPayload = `{	"e": "changed",	"id": "5",	"r": "sensors",	"state": {	  "fire": false,	  "lastupdated": "2018-03-13T19:46:03",	  "lowbattery": false,	  "tampered": false	},	"t": "event"  }`

// xiaomi flood detector
const floodDetectorFloodDetectedEventPayload = `{ "e": "changed", "id": "6", "r": "sensors", "state": { "lastupdated": "2018-03-13T20:46:03", "lowbattery": false, "tampered": false, "water": true }, "t": "event"   }`

// xiaomi random switch "sensor"
const switchSensorEventPayload = `{	"e": "changed",	"id": "7",	"r": "sensors",	"state": {	  "buttonevent": 1000,	  "lastupdated": "2018-03-20T20:52:18"	},	"t": "event"  }  `

type TestSensorProvider struct {
	Store *Sensors
}

func (l TestSensorProvider) Sensor(i int) (*Sensor, error) {
	if s, ok := (*l.Store)[i]; ok {
		return &s, nil
	}
	return nil, errors.New("not found")
}

func (l TestSensorProvider) Sensors() (*Sensors, error) {
	return l.Store, nil
}

var sensorInfo SensorProvider

func TestMain(m *testing.M) {

	sensorInfo = TestSensorProvider{Store: &Sensors{
		1: Sensor{Type: "ZHATemperature", Name: "ZHATemperature"},
		2: Sensor{Type: "ZHAHumidity", Name: "ZHAHumidity"},
		3: Sensor{Type: "ZHAPressure", Name: "ZHAPressure"},
		5: Sensor{Type: "ZHAFire", Name: "ZHAFire"},
		6: Sensor{Type: "ZHAWater", Name: "ZHAWater"},
		7: Sensor{Type: "ZHASwitch", Name: "ZHASwitch"},
	}}

	os.Exit(m.Run())
}

func TestSmokeDetectorNoFireEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(smokeDetectorNoFireEventPayload))
	if err != nil {
		t.Logf("unable to unmarshal smoke detector event: %s", err)
		t.FailNow()
	}

	smokeDetectorEvent, success := result.State().(*sensor.ZHAFire)
	if !success {
		t.Log("unable to type assert smoke detector event")
		t.FailNow()
	}

	if smokeDetectorEvent.Fire != false {
		t.Fail()
	}
}

func TestFloodDetectorEvent(t *testing.T) {

	result, err := DecodeEvent(sensorInfo, []byte(floodDetectorFloodDetectedEventPayload))
	if err != nil {
		t.Logf("Could not parse flood detector event: %s", err)
		t.FailNow()
	}

	floodEvent, success := result.State().(*sensor.ZHAWater)
	if !success {
		t.Log("Unable to type assert floodevent")
		t.FailNow()
	}

	if !floodEvent.Water {
		t.Fail()
	}

}

func TestPressureEvent(t *testing.T) {

	result, err := DecodeEvent(sensorInfo, []byte(pressureEventPayload))
	if err != nil {
		t.Logf("Could not parse pressure: %s", err)
		t.FailNow()
	}

	pressure, success := result.State().(*sensor.ZHAPressure)
	if !success {
		t.Log("Coudl not assert to pressureevent")
		t.FailNow()
	}

	if pressure.Pressure != 993 {
		t.Fail()
	}
}

func TestTemperatureEvent(t *testing.T) {

	result, err := DecodeEvent(sensorInfo, []byte(temperatureEventPayload))
	if err != nil {
		t.Logf("Could not parse temperature: %s", err)
		t.FailNow()
	}

	temp, success := result.State().(*sensor.ZHATemperature)
	if !success {
		t.Logf("Could not assert to temperature event")
		t.FailNow()
	}

	if temp.Temperature != 2062 {
		t.Fail()
	}
}

func TestHumidityEvent(t *testing.T) {

	result, err := DecodeEvent(sensorInfo, []byte(humidityEventPayload))
	if err != nil {
		t.Logf("Could not parse humidity: %s", err)
		t.FailNow()
	}

	humidity, success := result.State().(*sensor.ZHAHumidity)
	if !success {
		t.Logf("unable assert humidity event")
		t.FailNow()
	}

	if humidity.Humidity != 2985 {
		t.Fail()
	}
}

func TestSwitchEvent(t *testing.T) {

	result, err := DecodeEvent(sensorInfo, []byte(switchSensorEventPayload))
	if err != nil {
		t.Logf("Could not parse switch event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHASwitch)
	if !success {
		t.Logf("unable assert switch event")
		t.FailNow()
	}

	if s.Buttonevent != 1000 {
		t.Fail()
	}
}
