package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/fixje/deflux/deconz"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	yaml "gopkg.in/yaml.v2"
)

// YmlFileName is the filename
const YmlFileName = "deflux.yml"

type InfluxConfig struct {
	Url    string
	Token  string
	Org    string
	Bucket string
}

// Configuration holds data for Deconz and influxdb configuration
type Configuration struct {
	Deconz       deconz.Config
	InfluxConfig InfluxConfig
}

func main() {
	config, err := loadConfiguration()
	if err != nil {
		log.Printf("no configuration could be found: %s", err)
		outputDefaultConfiguration()
		return
	}

	sensorChan, err := sensorEventChan(config.Deconz)
	if err != nil {
		panic(err)
	}

	log.Printf("Connected to deCONZ at %s", config.Deconz.Addr)

	influxClient := influxdb2.NewClientWithOptions(
		config.InfluxConfig.Url,
		config.InfluxConfig.Token,
		influxdb2.DefaultOptions().SetBatchSize(20))

	// Get non-blocking write client
	writeAPI := influxClient.WriteAPI(config.InfluxConfig.Bucket, config.InfluxConfig.Bucket)
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
				log.Printf("not adding event to influx batch: %s", err)
				continue
			}

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

func sensorEventChan(c deconz.Config) (chan *deconz.SensorEvent, error) {
	// get an event reader from the API
	d := deconz.API{Config: c}
	reader, err := d.EventReader()
	if err != nil {
		return nil, err
	}

	// Dial the reader
	err = reader.Dial()
	if err != nil {
		return nil, err
	}

	// create a new reader, embedding the event reader
	sensorEventReader := d.SensorEventReader(reader)
	channel := make(chan *deconz.SensorEvent)
	// start it, it starts its own thread
	sensorEventReader.Start(channel)
	// return the channel
	return channel, nil
}

func loadConfiguration() (*Configuration, error) {
	data, err := readConfiguration()
	if err != nil {
		return nil, fmt.Errorf("could not read configuration: %s", err)
	}

	var config Configuration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse configuration: %s", err)
	}
	return &config, nil
}

// readConfiguration tries to read pwd/deflux.yml or /etc/deflux.yml
func readConfiguration() ([]byte, error) {
	// first try to load ${pwd}/deflux.yml
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current work directory: %s", err)
	}

	pwdPath := path.Join(pwd, YmlFileName)
	data, pwdErr := ioutil.ReadFile(pwdPath)
	if pwdErr == nil {
		log.Printf("Using configuration %s", pwdPath)
		return data, nil
	}

	// if we reached this code, we where unable to read a "local" Configuration
	// try from /etc/deflux.yml
	etcPath := path.Join("/etc", YmlFileName)
	data, etcErr := ioutil.ReadFile(etcPath)
	if etcErr != nil {
		return nil, fmt.Errorf("\n%s\n%s", pwdErr, etcErr)
	}

	log.Printf("Using configuration %s", etcPath)
	return data, nil
}

func outputDefaultConfiguration() {

	c := defaultConfiguration()

	// try to pair with deconz
	u, err := url.Parse(c.Deconz.Addr)
	if err == nil {
		apikey, err := deconz.Pair(*u)
		if err != nil {
			log.Printf("unable to pair with deconz: %s, please fill out APIKey manually", err)
		}
		c.Deconz.APIKey = string(apikey)
	}

	// we need to use a proxy struct to encode yml as the influxdb influxdb2 configuration struct
	// includes a Proxy: func() field that the yml encoder cannot handle
	yml, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to generate default configuration: %s", err)
	}

	log.Printf("Outputting default configuration, save this to /etc/deflux.yml")
	// to stdout
	fmt.Print(string(yml))
}

func defaultConfiguration() *Configuration {
	// this is the default configuration
	c := Configuration{
		Deconz: deconz.Config{
			Addr:   "http://127.0.0.1:8080/",
			APIKey: "change me",
		},
		InfluxConfig: InfluxConfig{
			Url:    "http://localhost:8086",
			Token:  "SECRET",
			Org:    "organization",
			Bucket: "default",
		},
	}

	// lets see if we are able to discover a gateway, and overwrite parts of the
	// default congfiguration
	discovered, err := deconz.Discover()
	if err != nil {
		log.Printf("discovery of deconz gateway failed: %s, please fill configuration manually..", err)
		return &c
	}

	// TODO: discover is actually a slice of multiple discovered gateways,
	// but for now we use only the first available
	deconz := discovered[0]
	addr := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", deconz.InternalIPAddress, deconz.InternalPort),
		Path:   "/api",
	}
	c.Deconz.Addr = addr.String()

	return &c
}
