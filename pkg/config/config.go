package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log/slog"
	"net/url"
	"os"
	"path"
	"time"
)

// YmlFileName is the name of the default config file
const YmlFileName = "deflux.yml"

// InfluxDB stores the InfluxDB configuration
type InfluxDB struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

// Configuration holds data for Deconz and InfluxDB configuration
type Configuration struct {
	Deconz     APIConfig
	InfluxDB   InfluxDB
	FillValues FillConfig
}

// FillConfig holds configuration for polling sensor measurements from the REST API
type FillConfig struct {
	// Enabled is set true if sensor values shall be added from the REST API, if no updates have been received
	// for some time over the websocket.
	Enabled bool

	// InitialFill set true will write the most recent measurement for each sensor from the REST API on startup
	InitialFill bool

	// FillInterval defines the duration after which the last sensor value from the REST API is inserted into the
	// database, if no more events have been seen from the websocket
	FillInterval time.Duration

	// LastSeenTimeout defines the duration after which a sensor is considered offline
	// It compares the wallclock time to the "lastseen" field of the /sensors REST endpoint
	LastSeenTimeout time.Duration
}

// LoadConfiguration loads the deflux configuration from a file.
// The file parameter provides a location. If it is empty, deflux tries the default config file locations
// ./deflux.yml and /etc/deflux.yml
func LoadConfiguration(file string) (*Configuration, error) {
	data, err := readConfiguration(file)
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

// readConfiguration reads the config file.
// If file is an empty string, it tries to read $(pwd)/deflux.yml and /etc/deflux.yml
func readConfiguration(file string) ([]byte, error) {
	if file != "" {
		return ioutil.ReadFile(file)
	}

	// first try to load ${pwd}/deflux.yml
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current work directory: %s", err)
	}

	pwdPath := path.Join(pwd, YmlFileName)
	data, pwdErr := ioutil.ReadFile(pwdPath)
	if pwdErr == nil {
		slog.Info(fmt.Sprintf("Using configuration %s", pwdPath))
		return data, nil
	}

	// ${pwd}/deflux.yml does not exist, try from /etc/deflux.yml
	etcPath := path.Join("/etc", YmlFileName)
	data, etcErr := ioutil.ReadFile(etcPath)
	if etcErr != nil {
		return nil, fmt.Errorf("\n%s\n%s", pwdErr, etcErr)
	}

	slog.Info(fmt.Sprintf("Using configuration %s", etcPath))
	return data, nil
}

// OutputDefaultConfiguration tries to pair with deCONZ and prints the default config to stdout
func OutputDefaultConfiguration() {

	c := defaultConfiguration()

	// try to pair with deCONZ
	u, err := url.Parse(c.Deconz.Addr)
	if err == nil {
		apikey, err := Pair(*u)
		if err != nil {

			if _, err := fmt.Fprintf(os.Stderr, "## Could not pair with deconz: %s\n", err); err != nil {
				panic(err)
			}
			if _, err := fmt.Fprintln(os.Stderr, "## Please add the API key manually"); err != nil {
				panic(err)
			}
		}
		c.Deconz.APIKey = string(apikey)
	}

	yml, err := yaml.Marshal(c)
	if err != nil {
		slog.Error("Unable to generate default configuration: %s", err)
		os.Exit(1)
	}

	// to stdout
	fmt.Print(string(yml))
}

func defaultConfiguration() *Configuration {
	// this is the default configuration
	c := Configuration{
		Deconz: APIConfig{
			Addr:   "http://127.0.0.1:8080/",
			APIKey: "change me",
		},
		InfluxDB: InfluxDB{
			URL:    "http://localhost:8086",
			Token:  "SECRET",
			Org:    "organization",
			Bucket: "default",
		},
		FillValues: FillConfig{
			Enabled:         false,
			InitialFill:     true,
			FillInterval:    30 * time.Minute,
			LastSeenTimeout: 2 * time.Hour,
		},
	}

	// let's see if we are able to discover a gateway, and overwrite parts of the
	// default configuration
	discovered, err := Discover()
	if err != nil {
		if _, err1 := fmt.Fprintf(os.Stderr, "## deCONZ Gateway discovery failed: %s. Complete config manually.\n", err); err1 != nil {
			panic(err1)
		}
		return &c
	}

	if len(discovered) > 1 {
		if _, err1 := fmt.Fprintln(os.Stderr, "## Found multiple  gateways, using the first for the configuration"); err1 != nil {
			panic(err1)
		}
		for i, di := range discovered {
			if _, err1 := fmt.Fprintf(os.Stderr, "### %d - http://%s:%d\n", i+1, di.InternalIPAddress, di.InternalPort); err1 != nil {
				panic(err1)
			}
		}
	}

	d := discovered[0]
	addr := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", d.InternalIPAddress, d.InternalPort),
		Path:   "/api",
	}
	c.Deconz.Addr = addr.String()

	return &c
}
