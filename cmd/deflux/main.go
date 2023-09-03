package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fixje/deflux/pkg/config"
	"github.com/fixje/deflux/pkg/deconz"
	"github.com/fixje/deflux/pkg/deconz/sensor"
	"github.com/fixje/deflux/pkg/sink"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"
)

func main() {
	flagLoglevel := flag.String("loglevel", "warn", "debug | error | warn | info")
	flagConfigGen := flag.Bool("config-gen", false, "generate a default config and print it on stdout")
	flagConfig := flag.String("config", "", "specify the location of the config file (default: ./deflux.yml or /etc/deflux.yml)")
	flagOnce := flag.Bool("1", false, "write sensor state from REST API once and exit")
	flag.Parse()

	var logLevel = new(slog.LevelVar)
	err := logLevel.UnmarshalText([]byte(*flagLoglevel))
	if err != nil {
		fmt.Printf("Error parsing log level: %v", err)
		logLevel.Set(slog.LevelInfo)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	if *flagConfigGen {
		config.OutputDefaultConfiguration()
		os.Exit(0)
	}

	cfg, err := config.LoadConfiguration(*flagConfig)
	if err != nil {
		slog.Error("No config file: %s", err)
		os.Exit(2)
	}

	if *flagOnce {
		os.Exit(runOnce(cfg))
	}

	os.Exit(runWebsocket(cfg))
}

// runOnce pulls sensor state from API, writes to InfluxDB and returns the program's exit code.
func runOnce(cfg *config.Configuration) int {
	// set up output to InfluxDB
	influx := sink.NewInfluxSink(cfg)
	defer influx.Close()

	dAPI := deconz.API{Config: cfg.Deconz}

	sensors, err := dAPI.Sensors()
	if err != nil {
		slog.Error("Failed to fetch sensors: %s", err)
		return 1
	}
	for _, s := range *sensors {
		writeSensorState(&s, &s, influx, time.Now(), nil)
	}

	return 0
}

// writeSensorState writes a sensor measurement to InfluxDB
func writeSensorState(ts deconz.Timeserieser, s *sensor.Sensor, influx *sink.InfluxSink, t time.Time, last map[int]*time.Time) {
	tags, fields, err := ts.Timeseries()
	if err != nil {
		slog.Warn("not adding sensor state to influx: %s", err)
		return
	}

	slog.Debug("Writing point", "sensor", s.Type, "tags", tags, "fields", fields)

	influx.Write(
		fmt.Sprintf("deflux_%s", s.Type),
		tags,
		fields,
		t,
	)

	if last != nil {
		last[s.ID] = &t
	}
}

// runWebsocket continuously processes events from the deCONZ websocket
func runWebsocket(cfg *config.Configuration) int {
	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)

	// set up input from deCONZ websocket
	dAPI := deconz.API{Config: cfg.Deconz}
	// TODO configurable update interval
	sensorProvider, err := deconz.NewCachingSensorProvider(dAPI, 1*time.Minute)

	if err != nil {
		slog.Error("Could not create websocket reader: %s", err)
		return 1
	}

	// create a new WebsocketEventReader using the websocket connection
	eventReader, err := deconz.NewWebsocketEventReader(dAPI, sensorProvider)
	if err != nil {
		slog.Error("Could not create websocket reader: %s", err)
		return 1
	}

	ctx1, cancel := context.WithCancel(context.Background())
	done := make(chan bool, 1)

	// set up output to InfluxDB
	influx := sink.NewInfluxSink(cfg)

	// start websocket consumer background job
	sensorsCh, err := eventReader.Start(ctx1)
	if err != nil {
		cancel()
		slog.Error("Could not start websocket reader: %s", err)
		return 2
	}

	slog.Info(fmt.Sprintf("Connected to deCONZ at %s", cfg.Deconz.Addr))

	lastWrite := make(map[int]*time.Time)
	ticker := time.NewTicker(1 * time.Minute)
	if cfg.FillValues.Enabled {
		slog.Info(fmt.Sprintf("Filling sensor values enabled. Fill interval is %v, timeout is %v", cfg.FillValues.FillInterval, cfg.FillValues.LastSeenTimeout))

		// TODO if InitialFill is false, compare "lastupdated" timestamp to current time and write
		if cfg.FillValues.InitialFill {
			sensors, err := sensorProvider.Sensors()
			if err != nil {
				slog.Error("Failed to fetch sensors for initial fill: %s", err)
			}
			for _, s := range *sensors {
				now := time.Now()

				if s.LastSeen.IsZero() {
					continue
				}
				if s.LastSeen.Add(cfg.FillValues.LastSeenTimeout).Before(now) {
					slog.Warn(fmt.Sprintf("sensor %d last seen %s ago -> assuming it's offline", s.ID, now.Sub(s.LastSeen)))
					continue
				}

				writeSensorState(&s, &s, influx, now, lastWrite)
			}
		}
	}

	// bring it all together
	go func(ctx context.Context) {
		for {
			select {
			case sensorEvent := <-sensorsCh:
				if sensorEvent == nil {
					continue
				}

				writeSensorState(sensorEvent, sensorEvent.Sensor, influx, time.Now(), lastWrite)

			case <-ticker.C:
				if !cfg.FillValues.Enabled {
					continue
				}

				slog.Debug(fmt.Sprintf("Checking sensor values older than %s", cfg.FillValues.FillInterval))

				for id, t := range lastWrite {
					now := time.Now()

					if t.Add(cfg.FillValues.FillInterval).After(now) {
						continue
					}

					s, err := sensorProvider.Sensor(id)
					if err != nil {
						slog.Warn(fmt.Sprintf("Could not retrieve sensor with id %d: %s", id, err))
						continue
					}

					if s.LastSeen.Add(cfg.FillValues.LastSeenTimeout).Before(now) {
						slog.Warn(fmt.Sprintf("sensor %d last seen %s ago -> assuming it's offline", s.ID, now.Sub(s.LastSeen)))
						continue
					}

					writeSensorState(s, s, influx, now, lastWrite)
				}

			case <-ctx.Done():
				ticker.Stop()
				influx.Close()
			}
		}

	}(ctx1)

	// signal handling
	go func() {
		select {
		case sig := <-sigsCh:
			slog.Debug("Received signal", "signal", sig)
			cancel()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			eventReader.Shutdown(ctx)
			cancel()
			done <- true
			return
		}
	}()

	<-done
	slog.Info("Exiting")
	return 0
}
