package deconz

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// CachingSensorStore is a cached typestore which provides LookupType for event passing
// it will be our default store
type CachingSensorStore struct {
	api   API
	cache *Sensors
}

func NewCachingSensorStore(api API) (*CachingSensorStore, error) {
	store := &CachingSensorStore{api: api}

	if store.cache == nil {
		err := store.populateCache()
		if err != nil {
			return nil, fmt.Errorf("unable to populate sensors: %s", err)
		}
	}

	return store, nil
}

// LookupType lookups deCONZ event types though a cache
// TODO: if we where unable to sensorProvider an ID we should try to refetch the cache
// - there could have been an sensor added we dont know about
// TODO use "enum": https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
func (c *CachingSensorStore) LookupType(i int) (string, error) {
	if s, found := (*c.cache)[i]; found {
		return s.Type, nil
	}

	return "", errors.New("no such sensor")
}

// LookupSensor returns a sensor for a sensor id
func (c *CachingSensorStore) LookupSensor(i int) (*Sensor, error) {
	if s, found := (*c.cache)[i]; found {
		return &s, nil
	}

	return nil, errors.New("no such sensor")
}

func (c *CachingSensorStore) Sensors() (*Sensors, error) {
	return c.cache, nil
}

func (c *CachingSensorStore) populateCache() error {
	var err error
	c.cache, err = c.api.Sensors()
	if err != nil {
		return err
	}

	log.Infof("SensorStore updated, found %d sensors", len((*c.cache)))

	return nil
}
