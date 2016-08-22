/*
    ToDD tsdbPackage implementation for influxdb

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package tsdb

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"

	"github.com/Mierdin/todd/config"
)

// newInfluxDB is a factory function that produces a new instance of influxDB with the configuration
// loaded and ready to be used.
func newInfluxDB(cfg config.Config) *influxDB {
	var ifdb influxDB
	ifdb.config = cfg
	return &ifdb
}

type influxDB struct {
	config config.Config
}

// WriteData will write the resulting testrun data to influxdb as a batch of points - containing
// important information like metrics and which agent reported them.
func (ifdb influxDB) WriteData(testUuid, testRunName, groupName string, testData map[string]map[string]map[string]string) error {

	// Make client
	c, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: fmt.Sprintf("http://%s:%s", ifdb.config.TSDB.Host, ifdb.config.TSDB.Port),
	})
	if err != nil {
		log.Error("Error creating InfluxDB Client: ", err.Error())
		return err
	}
	defer c.Close()

	// Create a new point batch
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  ifdb.config.TSDB.DatabaseName,
		Precision: "s",
	})
	if err != nil {
		log.Error("Error creating InfluxDB Batch Points: ", err)
		return err
	}

	// Need to publish data from all of the agents that took part in this test
	for agentUuid, agentData := range testData {

		// Also need to differentiate between the various target that these agents tested against
		for targetAddress, metrics := range agentData {

			// Create a point and add to batch
			tags := map[string]string{
				"agent":       agentUuid,
				"target":      targetAddress,
				"sourceGroup": groupName,
				"testUuid":    testUuid,
			}

			// Convert our metrics to float and insert into influx fields
			fields := make(map[string]interface{})
			for k, v := range metrics {
				float_v, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.Errorf("Error parsing metric value from %q, %q=%q: %v\n", targetAddress, k, v, err)
					return err
				}
				fields[k] = float_v
			}
			pt, err := influx.NewPoint(fmt.Sprintf("testrun-%s", testRunName), tags, fields, time.Now())
			if err != nil {
				log.Error("Error creating InfluxDB Point: ", err)
				return err
			}

			bp.AddPoint(pt)
		}

	}

	// Write the batch
	err = c.Write(bp)
	if err != nil {
		log.Error("Error writing InfluxDB Batch Points: ", err)
		return err
	}

	log.Infof("Wrote test data for %s to influxdb", testUuid)

	return nil
}
