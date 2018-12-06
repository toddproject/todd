/*
    ToDD Test Runs

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package testrun

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/agent/testing"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/hostresources"
	"github.com/toddproject/todd/server/objects"
	"github.com/toddproject/todd/server/tsdb"
)

func Start(cfg config.Config, trObj objects.TestRunObject, sourceOverrideMap map[string]string) string {

	// Generate UUID for test
	testUUID := hostresources.GenerateUUID()

	// Retrieve current group map
	tdb, err := db.NewToddDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}
	allGroupMap, err := tdb.GetGroupMap()
	if err != nil {
		log.Fatalf("Error retrieving group map: %v", err)
	}

	// sourceOverride is a flag to pass into the executeTest function so that it knows how to return test data if the source group has been overridden
	sourceOverride := false

	// Override source params as necessary
	if sourceOverrideMap["SourceApp"] != "" {
		trObj.Spec.Source["app"] = sourceOverrideMap["SourceApp"]
		sourceOverride = true
	}
	if sourceOverrideMap["SourceArgs"] != "" {
		trObj.Spec.Source["args"] = sourceOverrideMap["SourceArgs"]
		sourceOverride = true
	}
	if sourceOverrideMap["SourceGroup"] != "" {
		trObj.Spec.Source["name"] = sourceOverrideMap["SourceGroup"]
		sourceOverride = true
	}

	//testAgentMap is a custom map that only contains agents in the one or two groups relevant to this test
	//The outer map should contain two keys, "targets", and "sources". The values for each key should be another
	// map that uses the agent UUID for keys, and the group for those UUIDs as values.
	testAgentMap := make(map[string]map[string]string)
	testAgentMap["sources"] = make(map[string]string)
	testAgentMap["targets"] = make(map[string]string)

	// Here, we iterate over ALL of the agents, and pick out the ones that are part of this test, as well as what group they're in.
	for agent, group := range allGroupMap {

		// If our target type is group, and the group this agent is in matches the target group provided in the testrun object, add it to our map
		if trObj.Spec.TargetType == "group" && group == trObj.Spec.Target.(map[string]interface{})["name"].(string) {
			testAgentMap["targets"][agent] = group
		} else if group == trObj.Spec.Source["name"] {
			testAgentMap["sources"][agent] = group
		}
	}

	// Reject this topology if there aren't the right number of agents registered in this topology.
	if (trObj.Spec.TargetType == "group" && len(testAgentMap["targets"]) <= 0) || len(testAgentMap["sources"]) <= 0 {
		return "invalidtopology"
	}

	// Start listening for responses from agents
	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		os.Exit(1) //TODO(mierdin): remove
	}
	stopListeningForResponses := make(chan bool, 1)
	go tc.Package.ListenForResponses(&stopListeningForResponses)

	// Initialize test in database. This will create an entry for this test under the UUID we just created, and will also write the
	// list of agents participating in this test, with some kind of default status, for other goroutines to update with a further status.
	err = tdb.InitTestRun(testUUID, testAgentMap)
	if err != nil {
		log.Fatal("Problem initializing testrun in database.")
		return "failure"
	}

	// Prepare testrun instruction for our source agents
	var sourceTr = defs.TestRun{
		UUID:    testUUID,
		Testlet: trObj.Spec.Source["app"],
		Args:    trObj.Spec.Source["args"],
	}
	// Here, the list of target IP addresses is formed based on the target type
	if trObj.Spec.TargetType == "group" {

		// This is a group target type, so we are deriving target IPs from the DefaultAddr property of this agent.
		var targetIPs []string
		for uuid := range testAgentMap["targets"] {
			agent, err := tdb.GetAgent(uuid)
			if err != nil {
				log.Fatalf("Error retrieving agent: %v", err)
			}
			targetIPs = append(targetIPs, agent.DefaultAddr)
		}
		sourceTr.Targets = targetIPs

	} else {

		// This is an uncontrolled target, so we are deriving target IPs directly from what's listed in the testrun object
		var targetIPs []string
		specTargets := trObj.Spec.Target.([]interface{})
		for x := range specTargets {
			targetIPs = append(targetIPs, specTargets[x].(string))
		}

		sourceTr.Targets = targetIPs
	}

	// Prepare a task for carrying the testrun instruction to the agent
	itrTask := tasks.NewInstallTestRun(sourceTr)

	// Send testrun to each agent UUID in the sources group
	// TODO(mierdin): this is something I'd like to improve in the future. Right now this works, and is sort-of resilient, since
	// the testrun will require a response from each agent before actually moving on with execution, but I'd like something better.
	// Something that feels more like a true distributed system. Perfect is the enemy of good, however, and this works well for a prototype.
	for uuid := range testAgentMap["sources"] {
		tc.Package.SendTask(uuid, itrTask)
	}

	// If this testrun is targeted at another todd group, we want to send testrun tasks to those as well
	if trObj.Spec.TargetType == "group" {

		var targetTr = defs.TestRun{
			UUID:    testUUID,
			Targets: []string{"0.0.0.0"}, // Targets are typically running some kind of ongoing service, so we send a single target of 0.0.0.0 to indicate this.
			Testlet: trObj.Spec.Target.(map[string]interface{})["app"].(string),
			Args:    trObj.Spec.Target.(map[string]interface{})["args"].(string),
		}

		itrTask := tasks.NewInstallTestRun(targetTr)

		tc, err := comms.NewToDDComms(cfg)
		if err != nil {
			os.Exit(1) //TODO(mierdin): remove
		}

		// Send testrun to each agent UUID in the targets group
		for uuid := range testAgentMap["targets"] {
			tc.Package.SendTask(uuid, itrTask)
		}
	}

	// Start monitoring service
	leash := make(chan bool, 1)
	go testMonitor(cfg, testUUID, &leash)

	go executeTestRun(testAgentMap, testUUID, trObj, cfg, &leash, &stopListeningForResponses, sourceOverride)

	// Return the testUuid so that the client can subscribe to it.
	return testUUID
}

// executeTestRun will perform three things:
//
// - Monitor the database to determine which agents have which statuses
// - When the status for all agents is "ready", it will send execution tasks to one or both groups
// - It will continue to monitor, and when all agents have finished, it will pull the "leash" to stop the TCP stream to the client
// - After pulling the leash, it will call the function that will aggregate the test data and upload to a third party service
func executeTestRun(testAgentMap map[string]map[string]string, testUUID string, trObj objects.TestRunObject, cfg config.Config, leash, responseLeash *chan bool, sourceOverride bool) {

	// Sleep for 2 seconds so that the client moniting can connect first
	time.Sleep(2000 * time.Millisecond)

	tdb, err := db.NewToddDB(cfg) // TODO(vcabbage): Pass tdb in instead of creating new connection?
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	// First, let's just keep retrieving statuses until all of the agents are reporting ready
readyloop:
	for {
		time.Sleep(1000 * time.Millisecond)

		testStatuses, err := tdb.GetTestStatus(testUUID)
		if err != nil {
			log.Fatalf("Error retrieving test status: %v", err)
		}
		for _, status := range testStatuses {
			switch true {
			case status == testing.StatusFail:
				log.Error("Agent reported failure during testing")
				os.Exit(1)
			case status != testing.StatusReady:
				continue readyloop
			}
		}
		break
	}

	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		os.Exit(1) //TODO(mierdin): remove
	}

	// If this is a group target type, we want to make sure that the targets are set up and reporting a status of "testing"
	// before we spin up the source tests
	if trObj.Spec.TargetType == "group" {
		targetTask := tasks.NewExecuteTestRun(testUUID, cfg.Testing.Timeout)

		// Send testrun to each agent UUID in the targets group
		for uuid := range testAgentMap["targets"] {
			tc.Package.SendTask(uuid, targetTask)
		}

		// Next, we want to wait to make sure that the targets are all "testing" before instructing the source group to execute
	targetsreadyloop:
		for {

			time.Sleep(1000 * time.Millisecond)

			testStatuses, err := tdb.GetTestStatus(testUUID)
			if err != nil {
				log.Fatalf("Error retrieving test status: %v", err)
			}

			// We're creating this map to hold ONLY the targets that are in testStatuses (which also contains sources)
			targetStatuses := make(map[string]string)

			// Iterate all of the sources and agents
			for agent, status := range testStatuses {

				// if this agent is present in the group map....
				targetmap := testAgentMap["targets"]
				if val, ok := targetmap[agent]; ok {

					//...and it's also in our target group, add it to our targets map
					if val == trObj.Spec.Target.(map[string]interface{})["name"].(string) {
						targetStatuses[agent] = status
					}
				}
			}

			for agent, status := range targetStatuses {
				switch {
				case status == testing.StatusFail:
					log.Errorf("Agent %s reported failure during testing", agent)
					os.Exit(1)
				case status != testing.StatusTesting:
					log.Errorf("%s is not ready, so the sources must wait!!", agent)
					continue targetsreadyloop
				}
			}
			break
		}
	}

	// The targets are ready; execute testing on the source agents
	sourceTask := tasks.NewExecuteTestRun(testUUID, 30)

	// Send testrun to each agent UUID in the targets group
	for uuid := range testAgentMap["sources"] {
		tc.Package.SendTask(uuid, sourceTask)
	}

	// Let's wait once more until all agents are stored in the database with a status of "finished"
finishedloop:
	for {
		time.Sleep(1000 * time.Millisecond)

		testStatuses, err := tdb.GetTestStatus(testUUID)
		if err != nil {
			log.Fatalf("Error retrieving test status: %v", err)
		}
		for agent, status := range testStatuses {
			switch true {
			case status == testing.StatusFail:
				log.Errorf("Agent %s reported failure during testing", agent)
				os.Exit(1)
			case status != testing.StatusFinished:
				continue finishedloop
			}
		}
		break
	}

	uncondensedData, err := tdb.GetAgentTestData(testUUID, trObj.Spec.Source["name"])
	if err != nil {
		log.Fatalf("Error retrieving agent test data: %v", err)
	}

	cleanDataMap, err := cleanTestData(uncondensedData)
	if err != nil {
		log.Error("Failed to unmarshal raw test data")
		os.Exit(1)
	}

	cleanDataJSON, err := json.Marshal(cleanDataMap)
	if err != nil {
		log.Error("Problem converting cleaned data to JSON")
		os.Exit(1)
	}

	// Write clean test data to etcd
	tdb.WriteCleanTestData(testUUID, string(cleanDataJSON))

	time.Sleep(1000 * time.Millisecond)

	if !sourceOverride {
		testDataMap := make(map[string]map[string]map[string]interface{})
		err = json.Unmarshal(cleanDataJSON, &testDataMap)
		if err != nil {
			panic("Problem converting post-test data to a map")
		}

		timeDB := tsdb.NewToddTSDB(cfg)
		err = timeDB.Package.WriteData(testUUID, trObj.Label, trObj.Spec.Source["name"], testDataMap)
		if err != nil {
			log.Debug(err)
			log.Error("Problem writing metrics to TSDB")
		}

	}

	// Clean up our goroutines
	*leash <- true
	*responseLeash <- true

}

// cleanTestData cleans up the testing data that comes back from the agents.
// The test data that comes back raw from the agents is "dirty", meaning it is designed to be as flexible
// as possible.
func cleanTestData(dirtyData map[string]string) (map[string]map[string]map[string]interface{}, error) {

	retMap := make(map[string]map[string]map[string]interface{})
	for sourceUUID, agentData := range dirtyData {
		var agentMap map[string]map[string]interface{}

		err := json.Unmarshal([]byte(agentData), &agentMap)
		if err != nil {
			return nil, errors.New("Failed to unmarshal raw test data")
		}
		retMap[sourceUUID] = agentMap
	}

	return retMap, nil
}

// testMonitor offers a basic TCP stream for the ToDD client to subscribe to in order to receive updates during the course of a test.
func testMonitor(cfg config.Config, testUUID string, leash *chan bool) {

	// I implemented this retry functionality as a temporary fix for an issue that came up only once in a while, where
	// the net.Listen would throw an error indicating the port was already in use. Not sure why this happens yet, as I
	// am not running multiple instance of testMonitor at a time. TODO(mierdin): need to revisit this and perhaps
	// replace this retry with a Mutex.
	retries := 0

retrytcpserver:

	if retries > 5 {
		log.Error("TCP Server for testrun status failed to initialize.")
		return
	}

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		retries = retries + 1
		time.Sleep(1000 * time.Millisecond)
		goto retrytcpserver
	}
	defer ln.Close()

	// accept connection on port
	conn, err := ln.Accept()
	if err != nil {
		retries = retries + 1
		time.Sleep(1000 * time.Millisecond)
		goto retrytcpserver
	}
	defer conn.Close()

	tdb, err := db.NewToddDB(cfg)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Constantly poll for test status, and send statuses to client
	for {
		time.Sleep(1000 * time.Millisecond)

		testStatuses, err := tdb.GetTestStatus(testUUID)
		if err != nil {
			log.Fatalf("Error retrieving test status: %v", err)
		}

		statusesJSON, err := json.Marshal(testStatuses)
		if err != nil {
			log.Fatal("Failed to marshal agent test status message")
			os.Exit(1)
		}

		// Send status to client
		newMessage := string(statusesJSON)
		conn.Write([]byte(newMessage + "\n")) // TODO: Append '\n` to statusesJSON instead of converting to string`

		// Detect a client disconnect
		_, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return
		}

		// TODO(mierdin): Need to add failure notification (status of "fail" for any one agent)

		// Check to see if the calling function has asked that we shut down
		// Because of the presence of the default statement, this select statement will not block. It will allow the
		// loop to repeat if the channel does not contain data, and if it does this function will return.
		select {
		case _, _ = <-*leash:
			log.Debug("Killed testrun monitoring goroutine")
			return
		default:
		}
	}
}
