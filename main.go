package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
	"github.com/fixje/deflux/sink"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fixje/deflux/config"
	"github.com/fixje/deflux/deconz"
	log "github.com/sirupsen/logrus"
)

func main() {
	flagLoglevel := flag.String("loglevel", "warning", "debug | error | warning | info")
	flagConfigGen := flag.Bool("config-gen", false, "generate a default config and print it on stdout")
	flagConfig := flag.String("config", "", "specify the location of the config file (default: ./deflux.yml or /etc/deflux.yml)")
	flagOnce := flag.Bool("1", false, "write sensor state from REST API once and exit")
	flag.Parse()

	level, err := log.ParseLevel(*flagLoglevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if *flagConfigGen {
		config.OutputDefaultConfiguration()
		os.Exit(0)
	}

	cfg, err := config.LoadConfiguration(*flagConfig)
	if err != nil {
		log.Errorf("No config file: %s", err)
		os.Exit(2)
	}

	if *flagOnce {
		os.Exit(runOnce(cfg))
	}

	os.Exit(runWebsocket(err, cfg))
}

// runOnce pulls sensor state from API, writes to InfluxDB and returns the program's exit code.
func runOnce(cfg *config.Configuration) int {
	// set up output to InfluxDB
	influx := sink.NewInfluxSink(cfg)
	defer influx.Close()

	dApi := deconz.API{Config: cfg.Deconz}

	sensors, err := dApi.Sensors()
	if err != nil {
		log.Errorf("Failed to fetch sensors: %s", err)
		return 1
	}
	for _, s := range *sensors {
		writeSensorState(s, influx, time.Now(), nil)
	}

	return 0
}

// writeSensorState writes a sensor measurement to InfluxDB
func writeSensorState(s sensor.Sensor, influx *sink.InfluxSink, t time.Time, last map[int]*time.Time) {
	tags, fields, err := s.Timeseries()
	if err != nil {
		log.Warningf("not adding sensor state to influx: %s", err)
		return
	}

	log.Debugf("Writing point for sensor %s, tags = %v, fields = %v", s.Type, tags, fields)

	influx.Write(
		fmt.Sprintf("deflux_%s", s.Type),
		tags,
		fields,
		t,
	)

	if last != nil {
		last[s.Id] = &t
	}
}

// runWebsocket continuously processes events from the deCONZ websocket
func runWebsocket(err error, cfg *config.Configuration) int {
	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)

	// set up input from deCONZ websocket
	dApi := deconz.API{Config: cfg.Deconz}
	// TODO configurable update interval
	sensorProvider, err := deconz.NewCachingSensorProvider(dApi, 1*time.Minute)

	if err != nil {
		log.Errorf("Could not create websocket reader: %s", err)
		return 1
	}

	// create a new WebsocketEventReader using the websocket connection
	eventReader, err := deconz.NewWebsocketEventReader(dApi, sensorProvider)
	if err != nil {
		log.Errorf("Could not create websocket reader: %s", err)
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
		log.Errorf("Could not start websocket reader: %s", err)
		return 2
	}

	log.Infof("Connected to deCONZ at %s", cfg.Deconz.Addr)

	lastWrite := make(map[int]*time.Time)
	ticker := time.NewTicker(1 * time.Minute)
	if cfg.FillValues.Enabled {
		log.Infof("Filling sensor values enabled. Fill interval is %v, timeout is %v", cfg.FillValues.FillInterval, cfg.FillValues.LastSeenTimeout)

		// TODO if InitialFill is false, compare "lastupdated" timestamp to current time and write
		if cfg.FillValues.InitialFill {
			sensors, err := sensorProvider.Sensors()
			if err != nil {
				log.Errorf("Failed to fetch sensors for initial fill: %s", err)
			}
			for _, s := range *sensors {
				now := time.Now()

				if s.LastSeen.IsZero() {
					continue
				}
				if s.LastSeen.Add(cfg.FillValues.LastSeenTimeout).Before(now) {
					log.Warningf("sensor %d last seen %s ago -> assuming it's offline", s.Id, now.Sub(s.LastSeen))
					continue
				}

				writeSensorState(s, influx, now, lastWrite)
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

				tags, fields, err := sensorEvent.Timeseries()
				if err != nil {
					log.Warningf("not adding event to influx: %s", err)
					continue
				}

				log.Debugf("Writing point for sensor %s, tags = %v, fields = %v", sensorEvent.Sensor.Type, tags, fields)

				now := time.Now()

				influx.Write(
					fmt.Sprintf("deflux_%s", sensorEvent.Sensor.Type),
					tags,
					fields,
					now,
				)

				lastWrite[sensorEvent.ResourceId()] = &now

			case <-ticker.C:
				if !cfg.FillValues.Enabled {
					continue
				}

				log.Debugf("Checking sensor values older than %s", cfg.FillValues.FillInterval)

				for id, t := range lastWrite {
					now := time.Now()

					if t.Add(cfg.FillValues.FillInterval).After(now) {
						continue
					}

					s, err := sensorProvider.Sensor(id)
					if err != nil {
						log.Warningf("Could not retrieve sensor with id %d: %s", id, err)
						continue
					}

					if s.LastSeen.Add(cfg.FillValues.LastSeenTimeout).Before(now) {
						log.Warningf("sensor %d last seen %s ago -> assuming it's offline", s.Id, now.Sub(s.LastSeen))
						continue
					}

					writeSensorState(*s, influx, now, lastWrite)
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
			log.Debugf("Received signal %s", sig)
			cancel()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			eventReader.Shutdown(ctx)
			cancel()
			done <- true
			return
		}
	}()

	<-done
	log.Info("Exiting")
	return 0
}
