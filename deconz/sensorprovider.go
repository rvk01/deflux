package deconz

import (
	"errors"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
	log "github.com/sirupsen/logrus"
	"time"
)

// CachingSensorProvider is a sensor.Provider that retrieves sensor info from the dCONZ REST API and caches results
// It is the default sensor.Provider
type CachingSensorProvider struct {
	api            API
	cache          *sensor.Sensors
	nextFetch      time.Time
	updateInterval time.Duration
}

var sensorProvider *CachingSensorProvider

// NewCachingSensorProvider returns a CachingSensorProvider
// The CachingSensorProvider is a singleton.
func NewCachingSensorProvider(api API, updateInterval time.Duration) (*CachingSensorProvider, error) {
	if sensorProvider != nil {
		return sensorProvider, nil
	}

	sensorProvider = &CachingSensorProvider{api: api, updateInterval: updateInterval}

	err := sensorProvider.populateCache()
	if err != nil {
		sensorProvider = nil
		return nil, fmt.Errorf("unable to populate sensor cache: %s", err)
	}

	return sensorProvider, nil
}

// TODO use "enum": https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go

// Sensor returns a sensor for a sensor id
func (c *CachingSensorProvider) Sensor(i int) (*sensor.Sensor, error) {
	if err := c.populateCache(); err != nil {
		log.Errorf("failed to update sensor cache: %s", err)
	}

	if s, found := (*c.cache)[i]; found {
		return &s, nil
	}

	return nil, errors.New("no such sensor")
}

// Sensors returns all sensors in the cache
func (c *CachingSensorProvider) Sensors() (*sensor.Sensors, error) {
	if err := c.populateCache(); err != nil {
		log.Errorf("failed to update sensor cache: %s", err)
	}

	return c.cache, nil
}

func (c *CachingSensorProvider) populateCache() error {
	now := time.Now()

	if now.Before(c.nextFetch) {
		return nil
	}

	var err error
	c.cache, err = c.api.Sensors()
	if err != nil {
		return err
	}

	c.nextFetch = now.Add(c.updateInterval)
	log.Infof("Sensor cache updated, found %d sensors", len((*c.cache)))

	return nil
}
