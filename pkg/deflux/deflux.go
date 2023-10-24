package deflux

import (
	"context"
	"fmt"
	"github.com/rvk01/deflux/pkg/config"
	"github.com/rvk01/deflux/pkg/deconz"
	"github.com/rvk01/deflux/pkg/deconz/sensor"
	"github.com/rvk01/deflux/pkg/sink"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// ExitOK is a return code that indicates successful termination
	ExitOK int = 0
	// ExitFailConnect is a return code indicating that the application could not connect to a data source or sink
	ExitFailConnect = 1
	// ExitFailConfig is a return code indicating that there was no configuration found
	ExitFailConfig = 2
)

// RunOnce pulls sensor state from API, writes to InfluxDB and returns the program's exit code.
func RunOnce(cfg *config.Configuration) int {
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

	return ExitOK
}

// RunWebsocket continuously processes events from the deCONZ websocket
func RunWebsocket(cfg *config.Configuration) int {
	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)

	// set up input from deCONZ websocket
	dAPI := deconz.API{Config: cfg.Deconz}
	// TODO configurable update interval
	sensorProvider, err := deconz.NewCachingSensorProvider(dAPI, 1*time.Minute)

	if err != nil {
		slog.Error("Could not create websocket reader: %s", err)
		return ExitFailConnect
	}

	// create a new WebsocketEventReader using the websocket connection
	eventReader, err := deconz.NewWebsocketEventReader(dAPI, sensorProvider)
	if err != nil {
		slog.Error("Could not create websocket reader: %s", err)
		return ExitFailConnect
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
		return ExitFailConnect
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
	return ExitOK
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
