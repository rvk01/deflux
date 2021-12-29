package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/fixje/deflux/config"
	"github.com/fixje/deflux/deconz"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	log "github.com/sirupsen/logrus"
)

func main() {
	flagLoglevel := flag.String("loglevel", "warning", "debug | error | warning | info")
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

	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Errorf("no configuration could be found: %s", err)
		config.OutputDefaultConfiguration()
		return
	}

	sensorChan, err := sensorEventChan(cfg.Deconz)
	if err != nil {
		panic(err)
	}

	log.Infof("Connected to deCONZ at %s", cfg.Deconz.Addr)

	influxClient := influxdb2.NewClientWithOptions(
		cfg.InfluxDB.Url,
		cfg.InfluxDB.Token,
		influxdb2.DefaultOptions().SetBatchSize(20))

	// Get non-blocking write client
	writeAPI := influxClient.WriteAPI(cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)
	// Get errors channel
	errorsCh := writeAPI.Errors()

	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
		}
	}()

	for {

		select {
		case sensorEvent := <-sensorChan:
			tags, fields, err := sensorEvent.Timeseries()
			if err != nil {
				log.Warningf("not adding event to influx: %s", err)
				continue
			}

			log.Debugf("Writing point for sensor %s, tags = %v, fields = %v", sensorEvent.Sensor.Type, tags, fields)

			writeAPI.WritePoint(influxdb2.NewPoint(
				fmt.Sprintf("deflux_%s", sensorEvent.Sensor.Type),
				tags,
				fields,
				time.Now(), // TODO: we should use the time associated with the event...
			))

		}
	}

	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	influxClient.Close()
}

func sensorEventChan(c deconz.Config) (<-chan *deconz.SensorEvent, error) {
	// get an event reader from the API
	d := deconz.API{Config: c}
	store, err := deconz.NewCachingSensorProvider(d)

	if err != nil {
		return nil, err
	}

	reader, err := deconz.CreateWsReader(d, store)
	if err != nil {
		return nil, err
	}

	// Dial the reader
	// TODO should not be needed here
	//err = reader.Dial()
	//if err != nil {
	//	return nil, err
	//}

	// create a new reader, embedding the event reader
	sensorEventReader := deconz.CreateSensorEventReader(reader)

	// start it, it starts its own thread
	return sensorEventReader.Start()
}
