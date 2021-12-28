# deflux

deflux connects to deCONZ rest api, listens for sensor updates and write these to InfluxDB.

deCONZ supports a variaty of Zigbee sensors but have no historical data about their values - with deflux you'll be able to store all these measurements in influxdb where they can be queried from the command line or graphical tools such as grafana. 

This software was forked from the original
[deflux](https://github.com/fasmide/deflux) and moved to InfluxDB version 2.
Influx did major changes moving from version 1 to version 2, most notably the
introduction of a new query language called
[Flux](https://docs.influxdata.com/influxdb/cloud/query-data/get-started/).

The application supports the following types of sensors:

- Daylight
- CLIPPresence (_EXPERIMENTAL_)
- ZHAAirQuality (_EXPERIMENTAL_)
- ZHABattery (_EXPERIMENTAL_)
- ZHACarbonMonoxide (_EXPERIMENTAL_)
- ZHAConsumption (_EXPERIMENTAL_)
- ZHAFire
- ZHAHumidity
- ZHALightLevel (_EXPERIMENTAL_)
- ZHAOpenClose (_EXPERIMENTAL_)
- ZHAPower (_EXPERIMENTAL_)
- ZHAPresence (_EXPERIMENTAL_)
- ZHAPressure (_EXPERIMENTAL_)
- ZHASwitch
- ZHATemperature
- ZHAVibration (_EXPERIMENTAL_)
- ZHAWater

Sensors marked as _EXPERIMENTAL_ lack proper tests. If you are in posession of such a sensor, it would be nice if you
provided some JSON test data as in [this test](deconz/event/event_test.go).

## Usage

Start off by `go get`'ting deflux:

```
go get github.com/fixje/deflux
```

deflux tries to read `$(pwd)/deflux.yml` or `/etc/deflux.yml` in that order, if both fails it will try to discover deCONZ with their webservice and output a configuration sample to stdout. 

Hint: if you've temporarily unlocked the deconz gateway, it should be able to fill in the api key by it self, this needs some testing though...

First run generates a sample configuration:

```
$ deflux
ERRO[2021-12-26T11:28:03+01:00] no configuration could be found: could not read configuration:
open /home/fixje/hacks/deflux/deflux.yml: no such file or directory
open /etc/deflux.yml: no such file or directory 
ERRO[2021-12-26T11:28:03+01:00] unable to pair with deconz: unable to pair with deconz: link button not pressed, please fill out APIKey manually 
WARN[2021-12-26T11:28:03+01:00] Outputting default configuration, save this to /etc/deflux.yml 
deconz:
  addr: http://172.26.0.2:80/api
  apikey: ""
influxdb:
  url: http://localhost:8086
  token: SECRET
  org: organization
  bucket: default
```

Save the sample configuration and edit it to your needs, then run again.
The default log level of the application is `warning`. You can set the
`-loglevel=` flag to make it a bit more verbose:

```
$ ./deflux -loglevel=debug
INFO[2021-12-26T11:29:15+01:00] Using configuration /home/fixje/hacks/deflux/deflux.yml
INFO[2021-12-26T11:29:15+01:00] Connected to deCONZ at http://172.26.0.2:80/api 
INFO[2021-12-26T11:29:15+01:00] Deconz websocket connected
```

See `deflux -h` for more information on command line flags.


## InfluxDB

Sensor values are added as InfluxDB values and tagged with sensor type, id and name.
Different event types are stored in different measurements, meaning you will end up with multiple InfluxDB measurements.

### Schema Exploration

You can use Flux queries to explore the schema.

```
$ influx query --org YOUR_ORG << EOF
import "influxdata/influxdb/schema"
schema.measurements(bucket: "YOUR_BUCKET")
EOF

Result: _result
Table: keys: []
         _value:string
----------------------
       deflux_Daylight
    deflux_ZHAHumidity
    deflux_ZHAPressure
 deflux_ZHATemperature
```

```
$ influx query --org YOUR_ORG << EOF
import "influxdata/influxdb/schema"

schema.measurementTagKeys(
  bucket: "YOUR_BUCKET",
  measurement: "deflux_ZHATemperature"
)

EOF
Result: _result
Table: keys: []
         _value:string
----------------------
                _start
                 _stop
                _field
          _measurement
                    id
                  name
                  type
```

### Example Queries

Get temperature grouped by sensor name:

```
$ influx query --org YOUR_ORG << EOF
from(bucket: "YOUR_BUCKET")
  |> range(start: -3h)
  |> filter(fn: (r) =>
    r._measurement == "deflux_ZHATemperature"
    )
  |> keep(columns: ["_time", "name", "_value"])
|> group(columns: ["name"])
EOF

Result: _result
Table: keys: [name]
   name:string                      _time:time          _value:float
--------------  ------------------------------  --------------------
         th-sz  2021-12-27T06:37:07.741970950Z                 19.12
         th-sz  2021-12-27T06:59:22.576400599Z                 18.98
         th-sz  2021-12-27T07:01:23.235873787Z                 18.46
         th-sz  2021-12-27T07:04:04.025135987Z                 17.94
...
```

InfluxDB 2 has a nice query builder that will help you creating Flux queries.
Visit InfluxDB's web interface, log in, and click "Explore" in the navigation
bar.


## Development

The software can be built with standard Go tooling (`go build`).

You can cross-compile for Raspberry Pi 4 by setting `GOARCH` and `GOARM`:

```bash
GOOS=linux GOARCH=arm GOARM=7 go build
```
