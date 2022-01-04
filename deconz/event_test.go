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

// State of the following events was retrieved via the /sensors REST endpoint
// The rest of the messages are made up
const openCloseEventPayload1 = `{
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
}`
const openCloseEventPayload2 = `{
	"e": "changed",
	"id": "9",
	"r": "sensors",
	"t": "event",
	"state": {
		"lastupdated": "2022-01-04T05:57:50.067",
		"open": true
	}
}`
const batteryEvent = `{
	"e": "changed",
	"id": "10",
	"r": "sensors",
	"t": "event",
	"state": {
		"lastupdated": "2022-01-04T05:57:50.067",
		"battery": 77
	}
}`
const consumptionEvent = `{
	"e": "changed",
	"id": "11",
	"r": "sensors",
	"t": "event",
	"state": {
		"lastupdated": "2022-01-04T05:57:50.067",
		"consumption": 8,
		"power": 0
	}
}`
const clipPresenceEvent = `{
	"e": "changed",
	"id": "12",
	"r": "sensors",
	"t": "event",
	"state": {
		"lastupdated": "2022-01-04T05:57:50.067",
		"presence": true
	}
}`
const powerEvent = `{
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
}`
const lightlevelEvent = `{
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
}`
const airQualityEvent = `{
	"e": "changed",
	"id": "15",
	"r": "sensors",
	"t": "event",
	"state": {
		"airquality": "good",
		"airqualityppb": 79,
		"lastupdated": "2022-01-04T05:57:50.067"
	}
}`

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
		1:  Sensor{Type: "ZHATemperature", Name: "ZHATemperature"},
		2:  Sensor{Type: "ZHAHumidity", Name: "ZHAHumidity"},
		3:  Sensor{Type: "ZHAPressure", Name: "ZHAPressure"},
		5:  Sensor{Type: "ZHAFire", Name: "ZHAFire"},
		6:  Sensor{Type: "ZHAWater", Name: "ZHAWater"},
		7:  Sensor{Type: "ZHASwitch", Name: "ZHASwitch"},
		8:  Sensor{Type: "ZHAOpenClose", Name: "ZHAOpenClose"},
		9:  Sensor{Type: "ZHAOpenClose", Name: "ZHAOpenClose"},
		10: Sensor{Type: "ZHABattery", Name: "ZHABattery"},
		11: Sensor{Type: "ZHAConsumption", Name: "ZHAConsumption"},
		12: Sensor{Type: "CLIPPresence", Name: "CLIPPresence"},
		13: Sensor{Type: "ZHAPower", Name: "ZHAPower"},
		14: Sensor{Type: "ZHALightLevel", Name: "ZHALightLevel"},
		15: Sensor{Type: "ZHAAirQuality", Name: "ZHAAirQuality"},
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

func TestOpenCloseEventWithTampered(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(openCloseEventPayload1))
	if err != nil {
		t.Logf("Could not parse openclose event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHAOpenClose)
	if !success {
		t.Logf("unable assert openclose event")
		t.FailNow()
	}

	if s.Open == false {
		t.Fail()
	}

	if s.Lowbattery == false {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-01T12:39:38.370" {
		t.Fail()
	}
}

func TestOpenCloseEventWithoutTampered(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(openCloseEventPayload2))
	if err != nil {
		t.Logf("Could not parse openclose event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHAOpenClose)
	if !success {
		t.Logf("unable assert openclose event")
		t.FailNow()
	}

	if s.Open == false {
		t.Fail()
	}

	if s.Lowbattery == true {
		t.Fail()
	}

	if s.Tampered == true {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestBatteryEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(batteryEvent))
	if err != nil {
		t.Logf("Could not parse battery event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHABattery)
	if !success {
		t.Logf("unable assert battery event")
		t.FailNow()
	}

	if s.Battery != 77 {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestConsumptionEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(consumptionEvent))
	if err != nil {
		t.Logf("Could not parse consumption event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHAConsumption)
	if !success {
		t.Logf("unable assert consumption event")
		t.FailNow()
	}

	if s.Consumption != 8 {
		t.Fail()
	}

	if s.Power != 0 {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestClipPresenceEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(clipPresenceEvent))
	if err != nil {
		t.Logf("Could not parse presence event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.CLIPPresence)
	if !success {
		t.Logf("unable assert presence event")
		t.FailNow()
	}

	if !s.Presence {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestPowerEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(powerEvent))
	if err != nil {
		t.Logf("Could not parse power event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHAPower)
	if !success {
		t.Logf("unable assert power event")
		t.FailNow()
	}

	if s.Power != 0 {
		t.Fail()
	}

	if s.Current != 0 {
		t.Fail()
	}

	if s.Voltage != 236 {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestLightLevelEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(lightlevelEvent))
	if err != nil {
		t.Logf("Could not parse lightllevel event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHALightLevel)
	if !success {
		t.Logf("unable assert lightllevel event")
		t.FailNow()
	}

	if !s.Dark {
		t.Fail()
	}

	if s.Daylight {
		t.Fail()
	}

	if s.Lux != 3 {
		t.Fail()
	}

	if s.LightLevel != 4772 {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}

func TestAirQualityEvent(t *testing.T) {
	result, err := DecodeEvent(sensorInfo, []byte(airQualityEvent))
	if err != nil {
		t.Logf("Could not parse air quality event: %s", err)
		t.FailNow()
	}

	s, success := result.State().(*sensor.ZHAAirQuality)
	if !success {
		t.Logf("unable assert air quality event")
		t.FailNow()
	}

	if s.AirqualityPPB != 79 {
		t.Fail()
	}

	if s.Airquality != "good" {
		t.Fail()
	}

	if s.Lastupdated != "2022-01-04T05:57:50.067" {
		t.Fail()
	}
}
