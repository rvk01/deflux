package deconz

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// CachedSensorStore is a cached typestore which provides LookupType for event passing
// it will be our default store
type CachedSensorStore struct {
	SensorRepository
	cache *Sensors
}

// LookupType lookups deCONZ event types though a cache
// TODO: if we where unable to lookup an ID we should try to refetch the cache
// - there could have been an sensor added we dont know about
// TODO use "enum": https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
func (c *CachedSensorStore) LookupType(i int) (string, error) {
	var err error
	if c.cache == nil {
		err = c.populateCache()
		if err != nil {
			return "", fmt.Errorf("unable to populate sensors: %s", err)
		}
	}

	if s, found := (*c.cache)[i]; found {
		return s.Type, nil
	}

	return "", errors.New("no such sensor")
}

// LookupSensor returns a sensor for a sensor id
func (c *CachedSensorStore) LookupSensor(i int) (*Sensor, error) {
	var err error
	if c.cache == nil {
		err = c.populateCache()
		if err != nil {
			return nil, fmt.Errorf("unable to populate sensors: %s", err)
		}
	}

	if s, found := (*c.cache)[i]; found {
		return &s, nil
	}

	return nil, errors.New("no such sensor")
}

func (c *CachedSensorStore) populateCache() error {
	var err error
	c.cache, err = c.Sensors()
	if err != nil {
		return err
	}

	log.Infof("SensorStore updated, found %d sensors", len((*c.cache)))

	return nil
}
