/*
	ToDD Agent - Test Reporting

	This file contains functions that watch the agent cache and upload test data when present.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package testing

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/agent/cache"
	"github.com/Mierdin/todd/agent/responses"
	"github.com/Mierdin/todd/comms"
	"github.com/Mierdin/todd/config"
)

// WatchForFinishedTestRuns simply watches the local cache for any test runs that have test data.
// It will periodically look at the table and send any present test data back to the server as a response.
// When the server has successfully received this data, it will send a task back to this specific agent
// to delete this row from the cache.
func WatchForFinishedTestRuns(cfg config.Config) error {

	var ac = cache.NewAgentCache(cfg)

	agentUuid := ac.GetKeyValue("uuid")

	for {

		time.Sleep(5000 * time.Millisecond)

		testruns, err := ac.GetFinishedTestRuns()
		if err != nil {
			log.Error("Problem retrieving finished test runs")
			return errors.New("Problem retrieving finished test runs")
		}

		for testUuid, testData := range testruns {

			log.Debug("Found ripe testrun: ", testUuid)

			var utdr = responses.UploadTestDataResponse{
				TestUuid: testUuid,
				TestData: testData,
			}
			utdr.AgentUuid = agentUuid
			utdr.Type = "TestData" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?

			tc, err := comms.NewToDDComms(cfg)
			if err != nil {
				return err
			}
			tc.CommsPackage.SendResponse(utdr)

		}

	}

	return nil

}
