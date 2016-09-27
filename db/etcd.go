/*
   ToDD databasePackage implementation for etcd

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/server/objects"
)

// newEtcdDB is a factory function that produces a new instance of etcdDB with the configuration
// loaded and ready to be used.
func newEtcdDB(cfg config.Config) *etcdDB {
	etcdLoc := fmt.Sprintf("http://%s:%s", cfg.DB.Host, cfg.DB.Port)
	etcdCfg := client.Config{
		Endpoints: []string{etcdLoc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(etcdCfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	return &etcdDB{config: cfg, keysAPI: kapi}
}

type etcdDB struct {
	config  config.Config
	keysAPI client.KeysAPI
}

func (etcddb *etcdDB) Init() error {

	_, err := etcddb.keysAPI.Get(context.Background(), "/todd/agents", &client.GetOptions{Recursive: true})
	if err == nil {
		log.Info("Deleting '/todd/agents' key")
		_, err = etcddb.keysAPI.Delete(context.Background(), "/todd/agents", &client.DeleteOptions{Recursive: true, Dir: true})
		if err != nil {
			return err
		}
	}

	// TODO(mierdin): Consider deleting the entire /todd key here, and recreating all of the various subkeys, just to start from scratch.
	// Shouldn't need any previous data if you restart the server.

	return nil
}

// SetAgent will ingest an agent advertisement, and update or insert the agent record
// in the database as needed.
func (etcddb *etcdDB) SetAgent(adv defs.AgentAdvert) error {
	log.Infof("Setting '/todd/agents/%s' key", adv.Uuid)

	advJSON, err := json.Marshal(adv)
	if err != nil {
		log.Error("Problem converting Agent Advertisement to JSON")
		return err
	}

	// TODO(mierdin): TTL needs to be user-configurable
	resp, err := etcddb.keysAPI.Set(
		context.Background(),                      // context
		fmt.Sprintf("/todd/agents/%s", adv.Uuid),  // key
		string(advJSON),                           // value
		&client.SetOptions{TTL: time.Second * 30}, //optional args
	)
	if err != nil {
		log.Error("Problem setting agent in etcd")
		return err
	}

	log.Infof("Agent set in etcd. This advertisement is good until %s", resp.Node.Expiration)

	return nil
}

// GetAgent will retrieve a specific agent from the database by UUID
func (etcddb *etcdDB) GetAgent(uuid string) (*defs.AgentAdvert, error) {
	keyStr := fmt.Sprintf("/todd/agents/%s", uuid)

	log.Printf("Getting %q key value", keyStr)

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Errorf("Agent %s not found.", uuid)
		return nil, err
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	adv, err := nodeToAgentAdvert(resp.Node, uuid)
	if err != nil {
		return nil, err
	}

	return adv, nil
}

// nodeToAgentAdvert takes a etcd Node representing an AgentAdvert and returns an AgentAdvert
func nodeToAgentAdvert(node *client.Node, expectedUUID string) (*defs.AgentAdvert, error) {
	adv := new(defs.AgentAdvert)

	// Marshal API data into object
	err := json.Unmarshal([]byte(node.Value), adv)
	if err != nil {
		log.Error("Failed to unmarshal json into agent advertisement")
		return nil, err
	}

	// We want to use the TTLDuration field from etcd for simplicity
	adv.Expires = node.TTLDuration()

	// The etcd key should always match the inner JSON
	if expectedUUID != adv.Uuid {
		return nil, errors.New("UUID in etcd does not match inner JSON text")
	}

	return adv, nil
}

// GetAgents will retrieve all agents from the database
func (etcddb *etcdDB) GetAgents() ([]defs.AgentAdvert, error) {

	retAdv := []defs.AgentAdvert{}

	log.Print("Getting /todd/agents' key value")

	keyStr := "/todd/agents"

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Warn("Agent list empty when queried")
		return retAdv, nil
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	if !resp.Node.Dir {
		return nil, errors.New("Expected dir in etcd for agents - encountered single node")
	}

	for _, node := range resp.Node.Nodes {

		// Extract UUID from key string
		uuid := strings.Replace(node.Key, "/todd/agents/", "", 1)

		adv, err := nodeToAgentAdvert(node, uuid)
		if err != nil {
			return nil, err
		}

		retAdv = append(retAdv, *adv)

	}

	return retAdv, nil

}

// RemoveAgent will delete an agent advertisement present in etcd. This function exists for the rare situation when
// an Agent needs to be removed immediately, as opposed to simply waiting for the TTL to expire.
func (etcddb *etcdDB) RemoveAgent(adv defs.AgentAdvert) error {
	_, err := etcddb.keysAPI.Delete(context.Background(), fmt.Sprintf("/todd/agents/%s", adv.Uuid), &client.DeleteOptions{Recursive: true, Dir: true})
	if err != nil {
		return err
	}

	log.Infof("Removed '/todd/agents/%s' key", adv.Uuid)

	return nil
}

// GetObjects retrieves a list of ToddObjects stored within etcd, and returns this as a slice.
// This requires an "objType" string to specify the type of object being looked up.
func (etcddb *etcdDB) GetObjects(objType string) ([]objects.ToddObject, error) {

	retObj := []objects.ToddObject{}

	// Construct the path to the key depending on the objType param
	if objType == "" {
		// TODO(mierdin): support this empty type, and return an entire list of objects regardless of type
		return nil, errors.New("Object API queried with no type argument")
	}

	//Construct a path to the key based on the provided type
	keyStr := fmt.Sprintf("/todd/objects/%s/", objType)

	log.Info("Accessing objects at", keyStr)

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Warn("ToDD object store empty when queried")
		return retObj, nil
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	// We are expecting that this node is a directory
	if !resp.Node.Dir {
		// We are definitely expecting an ETCd directory, so we should return nothing if this is not the case.
		return nil, errors.New("Etcd query for objects did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range resp.Node.Nodes {
		log.Printf("Parsing object %s \n", node.Value)

		// Marshal API data into ToddObject
		var baseobj objects.BaseObject
		err = json.Unmarshal([]byte(node.Value), &baseobj)
		if err != nil {
			return nil, err
		}

		// Generate a more specific Todd Object based on the JSON data
		finalobj := baseobj.ParseToddObject([]byte(node.Value))

		retObj = append(retObj, finalobj)
	}

	return retObj, nil
}

// SetObject will insert or update a ToddObject within etcd
func (etcddb *etcdDB) SetObject(tobj objects.ToddObject) error {

	objJSON, err := json.Marshal(tobj)
	if err != nil {
		return err
	}

	log.Debugf("Setting '/todd/objects/%s/%s' key", tobj.GetType(), tobj.GetLabel())

	// Here, we set the key string, using the following format:
	// /todd/objects/<type>/<label(name)>
	keyStr := fmt.Sprintf("/todd/objects/%s/%s", tobj.GetType(), tobj.GetLabel())

	_, err = etcddb.keysAPI.Set(
		context.Background(), // context
		keyStr,               // key
		string(objJSON),      // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting object in etcd")
		return err
	}

	log.Infof("Wrote new Todd Object to etcd: %s/%s", tobj.GetType(), tobj.GetLabel())
	return nil
}

// DeleteObject will delete a ToddObject from etcd
func (etcddb *etcdDB) DeleteObject(label string, objtype string) error {
	_, err := etcddb.keysAPI.Delete(context.Background(), fmt.Sprintf("/todd/objects/%s/%s", objtype, label), &client.DeleteOptions{Recursive: true, Dir: true})
	if err != nil {
		return err
	}

	log.Infof("Removed '/todd/objects/%s/%s' key", objtype, label)

	return nil
}

// SetGroupMapping will update etcd with the results of a grouping calculation
func (etcddb *etcdDB) SetGroupMap(groupmap map[string]string) error {

	gmapJSON, err := json.Marshal(groupmap)
	if err != nil {
		log.Error("Problem converting group map to JSON")
		return err
	}

	log.Debug("Setting '/todd/groupmap' key")

	keyStr := "/todd/groupmap"

	_, err = etcddb.keysAPI.Set(
		context.Background(), // context
		keyStr,               // key
		string(gmapJSON),     // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting group map in etcd")
		return err
	}

	log.Infof("Updated group map in etcd: %s", gmapJSON)

	return nil
}

// GetGroupMap returns a map containing agent-to-group mappings. Agent UUIDs are used for keys
func (etcddb *etcdDB) GetGroupMap() (map[string]string, error) {

	retMap := map[string]string{}

	keyStr := "/todd/groupmap"

	log.Debug("Retrieving group map")

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Warnf("Error retrieving group mapping: %v", err)
		return retMap, nil
	}

	// Marshal etcd data into map
	err = json.Unmarshal([]byte(resp.Node.Value), &retMap)
	if err != nil {
		log.Error("Failed to retrieve group map from etcd")
		return nil, err
	}

	return retMap, nil
}

// InitTestRun is responsible for initializing a new test run within the database. This includes creating an entry for the test itself
// using the provided UUID for uniqueness, but also in the case of etcd, a nested entry for each agent participating in the test. Each
// Agent entry will be initially populated with that agent's current group and an initial status, but it will also house the result of
// that agent's testrun data, which will be aggregate dafter all agents have checked back in.
func (etcddb *etcdDB) InitTestRun(testUUID string, testAgentMap map[string]map[string]string) error {

	// Create high-level UUID key for this testrun
	log.Debug("Creating entry in etcd for testrun ", testUUID)
	_, err := etcddb.keysAPI.Set(
		context.Background(),                       // context
		fmt.Sprintf("/todd/testruns/%s", testUUID), // key
		"", // value
		&client.SetOptions{Dir: true, TTL: time.Second * 3000}, //optional args
		// TODO(mierdin): I set the TTL here so that I didn't dirty etcd with a bunch of old testruns while I develop this feature.
		// Need to decide if doing our own garbage collection is a better approach.
	)
	if err != nil {
		log.Error("Problem setting testrun UUID: ", testUUID)
		return err
	}

	// TODO(vcabbage): Parallelize this?
	// Create agent entry for each agent that is in the provided map
	for _, uuidmappings := range testAgentMap {

		// _ is either "targets" or "sources".
		// uuidmappings is a map[string]string that contains uuid (key) to group name (value) mappings for this test.

		for agent, group := range uuidmappings {
			// Create agent entry within this testrun
			log.Debugf("Creating agent entry within testrun %s for agent %s", testUUID, agent)
			_, err = etcddb.keysAPI.Set(
				context.Background(),                                        // context
				fmt.Sprintf("/todd/testruns/%s/agents/%s", testUUID, agent), // key
				"", // value
				&client.SetOptions{Dir: true}, //optional args
			)
			if err != nil {
				log.Error("Problem setting initial agent placeholder in testrun: ", testUUID)
				log.Error(err)
				return err
			}

			var initAgentProps = map[string]string{
				"group":  group,
				"status": "init",
				// Intentially omitting the "testdata" key here, because we will create it when the testdata is ready
			}

			for k, v := range initAgentProps {

				_, err = etcddb.keysAPI.Set(
					context.Background(),                                              // context
					fmt.Sprintf("/todd/testruns/%s/agents/%s/%s", testUUID, agent, k), // key
					v,   // value
					nil, //optional args
				)
				if err != nil {
					log.Error("Problem setting initial agent placeholder in testrun: ", testUUID)
					log.Error(err)
					return err
				}
			}
		}

	}

	return nil
}

// SetAgentTestStatus sets the status for an agent in a particular testrun key.
func (etcddb *etcdDB) SetAgentTestStatus(testUUID, agentUUID, status string) error {
	_, err := etcddb.keysAPI.Set(
		context.Background(),                                                   // context
		fmt.Sprintf("/todd/testruns/%s/agents/%s/status", testUUID, agentUUID), // key
		status, // value
		nil,    //optional args
	)
	if err != nil {
		log.Errorf("Problem updating status for agent %s in test %s", agentUUID, testUUID)
		log.Error(err)
		return err
	}

	return nil
}

// SetAgentTestData sets the post-test data for an agent in a particular testrun
func (etcddb *etcdDB) SetAgentTestData(testUUID, agentUUID, testData string) error {
	_, err := etcddb.keysAPI.Set(
		context.Background(),                                                     // context
		fmt.Sprintf("/todd/testruns/%s/agents/%s/testdata", testUUID, agentUUID), // key
		testData, // value
		nil,      //optional args
	)
	if err != nil {
		log.Errorf("Problem updating testdata for agent %s in test %s", agentUUID, testUUID)
		log.Error(err)
		return err
	}

	return nil
}

// GetTestStatus returns a map containing a list of agent UUIDs that are participating in the provided test, and their status in this test.
func (etcddb *etcdDB) GetTestStatus(testUUID string) (map[string]string, error) {

	retMap := make(map[string]string)

	keyStr := fmt.Sprintf("/todd/testruns/%s/agents", testUUID)

	log.Debug("Retrieving detailed test status for ", testUUID)

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Errorf("Error - empty test encountered for %q: %v", testUUID, err)
		return retMap, nil
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	// We are expecting that this node is a directory
	if !resp.Node.Dir {
		return nil, errors.New("Etcd query for detailed test status did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range resp.Node.Nodes {
		statusKey := fmt.Sprintf("%s/status", node.Key)

		// Extract UUID from key string
		agentUUID := strings.Replace(node.Key, fmt.Sprintf("/todd/testruns/%s/agents/", testUUID), "", 1)

		statusResp, err := etcddb.keysAPI.Get(context.Background(), statusKey, nil)
		if err != nil {
			log.Errorf("Error - empty agent status encountered: %s", testUUID)
			continue
		}

		retMap[agentUUID] = statusResp.Node.Value
	}

	return retMap, nil
}

// GetAgentTestData returns un-sanitized data from the individual agents. For a report of all agents' data,
// which has been sanitized by the server, see GetCleanTestData
func (etcddb *etcdDB) GetAgentTestData(testUUID, sourceGroup string) (map[string]string, error) {

	retMap := make(map[string]string)

	keyStr := fmt.Sprintf("/todd/testruns/%s/agents", testUUID)

	log.Debug("Retrieving detailed test data for ", testUUID)

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Errorf("Error - empty test encountered for %q: %v", testUUID, err)
		return nil, err
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	// We are expecting that this node is a directory
	if !resp.Node.Dir {
		return nil, errors.New("Etcd query for detailed test data did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range resp.Node.Nodes {
		// Extract UUID from key string
		agentUUID := strings.Replace(node.Key, fmt.Sprintf("/todd/testruns/%s/agents/", testUUID), "", 1)

		groupKey := fmt.Sprintf("%s/group", node.Key)
		groupResp, err := etcddb.keysAPI.Get(context.Background(), groupKey, nil)
		if err != nil {
			log.Errorf("Error retrieving group of agent in: %s", testUUID)
			return nil, err
		}

		if groupResp.Node.Value != sourceGroup {
			continue
		}

		testRunDataKey := fmt.Sprintf("%s/testdata", node.Key)
		dataResp, err := etcddb.keysAPI.Get(context.Background(), testRunDataKey, nil)
		if err != nil {
			log.Errorf("Error retrieving testdata of agent in: %s", testUUID)
			return nil, err
		}
		retMap[agentUUID] = dataResp.Node.Value
	}

	return retMap, nil
}

// WriteCleanTestData will write the post-test metrics data that has been cleaned up and
// ready to be displayed or exported to the database
func (etcddb *etcdDB) WriteCleanTestData(testUUID string, testData string) error {

	log.Debugf("/todd/testruns/%s/cleandata/", testUUID)

	keyStr := fmt.Sprintf("/todd/testruns/%s/cleandata/", testUUID)

	_, err := etcddb.keysAPI.Set(
		context.Background(), // context
		keyStr,               // key
		testData,             // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting object in etcd")
		return err
	}

	log.Infof("Wrote clean test data to test uuid: %s", testUUID)

	return nil
}

// GetCleanTestData will retrieve clean test data from the database
func (etcddb *etcdDB) GetCleanTestData(testUUID string) (string, error) {

	keyStr := fmt.Sprintf("/todd/testruns/%s/cleandata", testUUID)

	log.Debug("Retrieving clean test data for ", testUUID)

	resp, err := etcddb.keysAPI.Get(context.Background(), keyStr, &client.GetOptions{Recursive: true})
	if err, ok := err.(client.Error); ok {
		switch err.Code {
		case client.ErrorCodeKeyNotFound:
			return "", ErrNotExist
		default:
			log.Error(err)
			log.Errorf("Error - empty test data: %s", testUUID)
			return "", err
		}
	}

	return string(resp.Node.Value), nil
}

// TODO (mierdin): I have commented this out for now - may use this in the future to ensure that only one test is activated at a time.
//
// SetFlag will update etcd with the flag that indicates if tests can be run.
// func (etcddb EtcdDB) SetFlag(s string) errors.Error {

//     // Enforce the two valid values for this function
//     if s != "nogo" && s != "go" {
//         return errors.New("Invalid value for SetFlag")
//     }

//     cfg := client.Config{
//         Endpoints: []string{"http://192.168.59.103:4001"},
//         Transport: client.DefaultTransport,
//         // set timeout per request to fail fast when the target endpoint is unavailable
//         HeaderTimeoutPerRequest: time.Second,
//     }
//     c, err := client.New(cfg)
//     if err != nil {
//         log.Fatal(err)
//     }
//     kapi := client.NewKeysAPI(c)

//     _, err = kapi.Set(
//         context.Background(), // context
//         "/todd/flag",         // key
//         string(s),            // value
//         nil,                  //optional args
//     )
//     common.FailOnError(err, "Problem setting flag in etcd")

//     log.Infof("Set '/todd/flag' key to %s", s)

//     return nil
// }
