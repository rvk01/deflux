package main

import (
	"context"
	"flag"
	"fmt"
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
	flagConfig := flag.Bool("config-gen", false, "generates a default config and prints it to stdout")
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

	if *flagConfig {
		config.OutputDefaultConfiguration()
		return
	}

	cfg, err := config.LoadConfiguration()
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

		tags, fields, err := s.Timeseries()
		if err != nil {
			log.Warningf("not adding sensor state to influx: %s", err)
			continue
		}

		log.Debugf("Writing point for sensor %s, tags = %v, fields = %v", s.Type, tags, fields)

		influx.Write(
			fmt.Sprintf("deflux_%s", s.Type),
			tags,
			fields,
			time.Now(), // TODO: we should use the time associated with the event...
		)
	}

	return 0
}

// runWebsocket continuously processes events from the deCONZ websocket
func runWebsocket(err error, cfg *config.Configuration) int {
	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)

	// set up input from deCONZ websocket
	eventReader, err := eventReader(cfg.Deconz)
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

				influx.Write(
					fmt.Sprintf("deflux_%s", sensorEvent.Sensor.Type),
					tags,
					fields,
					time.Now(), // TODO: we should use the time associated with the event...
				)

			case <-ctx.Done():
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

func eventReader(c config.ApiConfig) (*deconz.WebsocketEventReader, error) {
	dApi := deconz.API{Config: c}
	// TODO configurable update interval
	store, err := deconz.NewCachingSensorProvider(dApi, 1*time.Minute)

	if err != nil {
		return nil, err
	}

	// create a new WebsocketEventReader using the websocket connection
	return deconz.NewWebsocketEventReader(dApi, store)
}
