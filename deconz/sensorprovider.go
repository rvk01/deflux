package deconz

import (
	"errors"
	"fmt"
	"github.com/fixje/deflux/deconz/sensor"
	log "github.com/sirupsen/logrus"
)

// CachingSensorProvider is a SensorProvider that retrieves sensor info from the dCONZ REST API and caches results
// It is the default SensorProvider
type CachingSensorProvider struct {
	api   API
	cache *sensor.Sensors
}

var sensorProvider *CachingSensorProvider

func NewCachingSensorProvider(api API) (*CachingSensorProvider, error) {
	if sensorProvider != nil {
		return sensorProvider, nil
	}

	sensorProvider = &CachingSensorProvider{api: api}

	err := sensorProvider.populateCache()
	if err != nil {
		sensorProvider = nil
		return nil, fmt.Errorf("unable to populate sensor cache: %s", err)
	}

	return sensorProvider, nil
}

// TODO: if we where unable to sensorProvider an ID we should try to refetch the cache
// - there could have been an sensor added we dont know about
// TODO use "enum": https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go

// Sensor returns a sensor for a sensor id
func (c *CachingSensorProvider) Sensor(i int) (*sensor.Sensor, error) {
	if s, found := (*c.cache)[i]; found {
		return &s, nil
	}

	return nil, errors.New("no such sensor")
}

func (c *CachingSensorProvider) Sensors() (*sensor.Sensors, error) {
	return c.cache, nil
}

func (c *CachingSensorProvider) populateCache() error {
	var err error
	c.cache, err = c.api.Sensors()
	if err != nil {
		return err
	}

	log.Infof("Sensor cache updated, found %d sensors", len((*c.cache)))

	return nil
}
