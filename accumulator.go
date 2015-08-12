package telegraf

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdb/influxdb/client"
)

// BatchPoints is used to send a batch of data in a single write from telegraf
// to influx
type BatchPoints struct {
	mu sync.Mutex

	client.BatchPoints

	Debug bool

	Prefix string

	Config *ConfiguredPlugin
}

// Add adds a measurement with one field, "value"
func (bp *BatchPoints) Add(measurement string, val interface{}, tags map[string]string) {
	var timestamp time.Time
	bp.AddWithTime(measurement, val, tags, timestamp)
}

// Adds a measurement with one field, "value" with a provided timestamp
func (bp *BatchPoints) AddWithTime(
	measurement string,
	val interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	values := map[string]interface{}{"value": val}
	bp.AddValuesWithTime(measurement, values, tags, timestamp)
}

// Adds a measurement with a given set of values
func (bp *BatchPoints) AddValues(
	measurement string,
	values map[string]interface{},
	tags map[string]string,
) {
	var timestamp time.Time
	bp.AddValuesWithTime(measurement, values, tags, timestamp)
}

// AddValuesWithTime adds a measurement with a provided timestamp
func (bp *BatchPoints) AddValuesWithTime(
	measurement string,
	values map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	measurement = bp.Prefix + measurement

	if bp.Config != nil {
		if !bp.Config.ShouldPass(measurement, tags) {
			return
		}
	}

	if bp.Debug {
		var tg []string

		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		var vals []string

		for k, v := range values {
			vals = append(vals, fmt.Sprintf("%s=%v", k, v))
		}

		sort.Strings(tg)
		sort.Strings(vals)

		fmt.Printf("> [%s] %s %s\n", strings.Join(tg, " "), measurement, strings.Join(vals, " "))
	}

	newPoint := client.Point{
		Measurement: measurement,
		Tags:        tags,
		Fields:      values,
		Time:        timestamp,
	}

	bp.Points = append(bp.Points, newPoint)
}
