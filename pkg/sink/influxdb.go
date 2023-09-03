package sink

import (
	"fmt"
	"github.com/fixje/deflux/pkg/config"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

// InfluxSink writes data to InfluxDB
type InfluxSink struct {
	client influxdb2.Client
	writer api.WriteAPI
}

// NewInfluxSink returns a new instance of InfluxSink
// The instance needs to be closed with Close()
func NewInfluxSink(cfg *config.Configuration) *InfluxSink {
	influxClient := influxdb2.NewClientWithOptions(
		cfg.InfluxDB.URL,
		cfg.InfluxDB.Token,
		influxdb2.DefaultOptions().SetBatchSize(20))

	// Get non-blocking write client
	writeAPI := influxClient.WriteAPI(cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)
	// Get errors channel
	errorsCh := writeAPI.Errors()

	// read and log errors in a separate go routine
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
		}
	}()

	return &InfluxSink{
		client: influxClient,
		writer: writeAPI,
	}
}

// Write persists a data point to InfluxDB
// It takes the table name, tags and fields and the time as arguments
func (i *InfluxSink) Write(table string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	i.writer.WritePoint(influxdb2.NewPoint(
		table,
		tags,
		fields,
		t,
	))
}

// Close closes the InfluxSink
func (i *InfluxSink) Close() {
	// FIXME panics 'send on closed channel'
	// i.writer.Flush()
	i.client.Close()
}
