package deconz

import (
	ctx "context"
	"strconv"
	"testing"
)

type testLookup struct {
}

func (t *testLookup) Sensors() (*Sensors, error) {
	return nil, nil
}

func (t *testLookup) Sensor(i int) (*Sensor, error) {
	return &Sensor{Name: "Test Sensor", Type: "ZHAFire"}, nil
}

func (t *testLookup) LookupType(i int) (string, error) {
	return "ZHAFire", nil
}

type testReader struct {
}

func (t testReader) ReadEvent() (Event, error) {
	return DecodeEvent(&testLookup{}, []byte(smokeDetectorNoFireEventPayload))
}
func (t testReader) Dial(ctx ctx.Context) error {
	return nil
}
func (t testReader) Close() error {
	return nil
}

func TestSensorEventReader(t *testing.T) {

	r := SensorEventReader{reader: testReader{}}

	channel, err := r.Start(ctx.Background())
	if err != nil {
		t.Fail()
	}
	e := <-channel
	if strconv.Itoa(e.Event.Id()) != "5" {
		t.Fail()
	}
	tags, fields, err := e.Timeseries()
	if err != nil {
		t.Logf(err.Error())
		t.FailNow()
	}
	if tags["name"] != "Test Sensor" {
		t.Fail()
	}
	if tags["id"] != "5" {
		t.Fail()
	}

	if fields["fire"] != false {
		t.Fail()
	}

}
