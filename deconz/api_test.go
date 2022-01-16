package deconz

import (
	"github.com/fixje/deflux/config"
	"github.com/fixje/deflux/deconz/sensor"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestApiSensors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
		  "4": {
			"config": {
			  "battery": 91,
			  "offset": 0,
			  "on": true,
			  "reachable": true
			},
			"ep": 1,
			"etag": "62dd73fcab1234567cc8218441b48714",
			"lastannounced": null,
			"lastseen": "2022-01-09T17:58Z",
			"manufacturername": "LUMI",
			"modelid": "lumi.weather",
			"name": "th-sz",
			"state": {
			  "lastupdated": "2022-01-09T17:58:29.629",
			  "pressure": 996
			},
			"swversion": "20191205",
			"type": "ZHAPressure",
			"uniqueid": "00:15:8d:12:34:bd:ff:71-01-0403"
		  },
		  "5": {
			"config": {
			  "battery": null,
			  "enrolled": 1,
			  "on": true,
			  "pending": [],
			  "reachable": true
			},
			"ep": 1,
			"etag": "717dfba9e3f0ed123456785c3b3cbf28",
			"lastannounced": "2022-01-04T15:00:40Z",
			"lastseen": "2022-01-09T18:18Z",
			"manufacturername": "LIDL Silvercrest",
			"modelid": "TY0203",
			"name": "wi-wc",
			"state": {
			  "lastupdated": "2022-01-09T18:12:29.179",
			  "lowbattery": false,
			  "open": false,
			  "tampered": false
			},
			"type": "ZHAOpenClose",
			"uniqueid": "68:b0:e2:ff:fe:12:34:ff-01-0500"
		  }
		}`

		if _, err := w.Write([]byte(resp)); err != nil {
			t.Fatalf("failed to send response: %s", err)
		}
	}))
	defer ts.Close()

	api := API{
		Config: config.APIConfig{
			Addr:   ts.URL,
			APIKey: "",
			WsAddr: "",
		},
		sensorCache: nil,
	}

	sensors, err := api.Sensors()
	if err != nil {
		t.Fatalf("failed to get sensors: %s", err)
	}

	lastSeen4, _ := time.Parse("2006-01-02T15:04Z", "2022-01-09T17:58Z")
	lastSeen5, _ := time.Parse("2006-01-02T15:04Z", "2022-01-09T18:18Z")

	want := &sensor.Sensors{
		4: sensor.Sensor{
			Type:     "ZHAPressure",
			Name:     "th-sz",
			LastSeen: lastSeen4,
			StateDef: &sensor.ZHAPressure{
				State:    sensor.State{Lastupdated: "2022-01-09T17:58:29.629"},
				Pressure: 996,
			},
			Config: sensor.Config{Battery: 91},
			ID:     4,
		},
		5: sensor.Sensor{
			Type:     "ZHAOpenClose",
			Name:     "wi-wc",
			LastSeen: lastSeen5,
			StateDef: &sensor.ZHAOpenClose{
				State:      sensor.State{Lastupdated: "2022-01-09T18:12:29.179"},
				Tampered:   false,
				Lowbattery: false,
				Open:       false,
			},
			Config: sensor.Config{Battery: 0},
			ID:     5,
		},
	}

	if !reflect.DeepEqual(want, sensors) {
		t.Fatalf("expected: %v, got: %v", &want, sensors)
	}
}

// Regression test for wrong ZHABattery time series
func TestBatteryTimeSeries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"1": {
				"config": {
					"on": true,
					"reachable": true
				},
				"ep": 1,
				"etag": "500b6bd683da649af22b6bddde61824f",
				"lastannounced": "2021-12-24T11:37:43Z",
				"lastseen": "2022-01-11T11:48Z",
				"manufacturername": "IKEA of Sweden",
				"modelid": "FYRTUR block-out roller blind",
				"name": "batterytest",
				"state": {
					"battery": 75,
					"lastupdated": "2021-12-20T06:03:35.000"
				},
				"swversion": "2.2.009",
				"type": "ZHABattery",
				"uniqueid": "84:71:27:ff:fe:25:f7:b3-01-0001"
			}
		}`

		if _, err := w.Write([]byte(resp)); err != nil {
			t.Fatalf("failed to send response: %s", err)
		}
	}))
	defer ts.Close()

	api := API{
		Config: config.APIConfig{
			Addr:   ts.URL,
			APIKey: "",
			WsAddr: "",
		},
		sensorCache: nil,
	}

	now := time.Now()
	sensors, err := api.Sensors()
	if err != nil {
		t.Fatalf("failed to get sensors: %s", err)
	}

	lastSeen, _ := time.Parse("2006-01-02T15:04Z", "2022-01-11T11:48Z")

	want := &sensor.Sensors{
		1: sensor.Sensor{
			Type:     "ZHABattery",
			Name:     "batterytest",
			LastSeen: lastSeen,
			StateDef: &sensor.ZHABattery{
				State:   sensor.State{Lastupdated: "2021-12-20T06:03:35.000"},
				Battery: 75,
			},
			ID: 1,
		},
	}

	if !reflect.DeepEqual(want, sensors) {
		t.Fatalf("expected: %v, got: %v", &want, sensors)
	}

	for _, s := range *sensors {
		tags, fields, err := s.Timeseries()

		if err != nil {
			t.Fatalf("timeseries has error: %s", err)
		}

		wantTags := map[string]string{
			"name":   "batterytest",
			"type":   "ZHABattery",
			"id":     "1",
			"source": "rest",
		}

		if !reflect.DeepEqual(wantTags, tags) {
			t.Fatalf("expected: %v, got: %v", wantTags, tags)
		}

		wantFields := map[string]interface{}{
			"battery":  int16(75),
			"age_secs": int64(now.Sub(time.Date(2021, 12, 20, 6, 3, 35, 0, time.UTC)).Seconds()),
		}

		if !reflect.DeepEqual(wantFields, fields) {
			t.Fatalf("expected: %v, got: %v", wantFields, fields)
		}

	}
}
