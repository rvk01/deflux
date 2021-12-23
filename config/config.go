package config

import (
	"fmt"
	"github.com/fixje/deflux/deconz"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
)

// YmlFileName is the filename
const YmlFileName = "deflux.yml"

type InfluxDB struct {
	Url    string
	Token  string
	Org    string
	Bucket string
}

// Configuration holds data for Deconz and influxdb configuration
type Configuration struct {
	Deconz   deconz.Config
	InfluxDB InfluxDB
}

func LoadConfiguration() (*Configuration, error) {
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

func OutputDefaultConfiguration() {

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
		InfluxDB: InfluxDB{
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
