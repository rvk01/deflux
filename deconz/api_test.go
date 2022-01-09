package deconz

import (
	"encoding/json"
	"github.com/fixje/deflux/config"
	"github.com/fixje/deflux/deconz/sensor"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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
			t.Fatalf("failed to sens response: %s", err)
		}
	}))
	defer ts.Close()

	api := API{
		Config:      config.ApiConfig{
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

	want := &sensor.Sensors{
		4: sensor.Sensor{
			Type:     "ZHAPressure",
			Name:     "th-sz",
			LastSeen: "2022-01-09T17:58Z",
			RawState: json.RawMessage{},
			StateDef: &sensor.ZHAPressure{
				State:    sensor.State{Lastupdated: "2022-01-09T17:58:29.629"},
				Pressure: 996,
			},
			Config:   sensor.Config{Battery: 91},
			Id:       4,
		},
		5: sensor.Sensor{
			Type:     "ZHAOpenClose",
			Name:     "wi-wc",
			LastSeen: "2022-01-09T18:18Z",
			RawState: json.RawMessage{},
			StateDef: &sensor.ZHAOpenClose{
				State:    sensor.State{Lastupdated: "2022-01-09T18:12:29.179"},
				Tampered: false,
				Lowbattery: false,
				Open: false,
			},
			Config:   sensor.Config{Battery: 0},
			Id:       5,
		},
	}

	if !reflect.DeepEqual(want, sensors) {
		t.Fatalf("expected: %v, got: %v", &want, sensors)
	}
}