package comms

/*
  ToDD communicationPackage implementation for AWS Amazon Message Queue

  Inspired from the rabbitMQ implementation
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jjeffery/stomp"
	"crypto/tls"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
)


// newAMQComms is a factory function that produces a new instance of AMQComms with the configuration
// loaded and ready to be used.
func newAMQComms(cfg config.Config) *AMQComms {
	var amq AMQComms
	amq.config = cfg

	amq.queueURL = fmt.Sprintf("%s:%s",
		amq.config.Comms.Host,
		amq.config.Comms.Port,
	)
	amq.Login =	amq.config.Comms.User
	amq.Passcode =	amq.config.Comms.Password
	return &amq
}

type AMQComms struct {
	config   config.Config
	queueURL string
	ac       *cache.AgentCache
	Login,Passcode  string
}

// connectAMQ wraps the amqp.Dial function in order to provide connection retry functionality
func connectAMQ(amq AMQComms) (*stomp.Conn, error) {

	conn, err := tls.Dial("tcp", amq.queueURL,&tls.Config{})
	if err != nil {
		return nil, err
	}
	stomp_conn,err:=stomp.Connect(conn,
		stomp.ConnOpt.Login(amq.config.Comms.User, amq.config.Comms.Password))
	if err != nil {
		return nil, err
	}
	

	return stomp_conn, nil
}

// AdvertiseAgent will place an agent advertisement message on the message queue
func (amq AMQComms) AdvertiseAgent(me defs.AgentAdvert) error {

	// Connect to AMQ with retry logic
	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect()

	// Marshal agent struct to JSON
	jsonData, err := json.Marshal(me)
	if err != nil {
		log.Error("Failed to marshal agent data from queue")
		log.Debug(err)
		return err
	}

	err = conn.Send("/topic/todd/agentadvert", "", jsonData)
	
	if err != nil {
		log.Error("Failed to publish agent advertisement")
		log.Debug(err)
		return err
	}

	log.Infof("AGENTADV -- %s", time.Now().UTC())

	return nil
}

// ListenForAgent will listen on the message queue for new agent advertisements.
// It is meant to be run as a goroutine
func (amq AMQComms) ListenForAgent(assets assetProvider, assetURLPrefix string) error {

	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect()

	sub, _ := conn.Subscribe("/topic/todd/agentadvert", stomp.AckAuto)
	
	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range sub.C {

			log.Debugf("Agent advertisement recieved: %s", d.Body)

			var agent defs.AgentAdvert
			err = json.Unmarshal(d.Body, &agent)

			// assetList is a slice that will contain any URLs that need to be sent to an
			// agent as a response to an incorrect or incomplete list of assets
			var assetList []string

			// assets is the asset map from the SERVER's perspective
			for assetType, assetHashes := range assets.Assets() {

				var agentAssets map[string]string

				// agentAssets is the asset map from the AGENT's perspective
				if assetType == "factcollectors" { // TODO: Use select
					agentAssets = agent.FactCollectors
				} else if assetType == "testlets" {
					agentAssets = agent.Testlets
				}

				for name, hash := range assetHashes {

					// See if the hashes match (a missing asset will also result in False)
					if agentAssets[name] != hash {

						// hashes do not match, so we need to append the asset download URL to the remediate list
						assetURL := fmt.Sprintf("%s/%s/%s", assetURLPrefix, assetType, name)
						assetList = append(assetList, assetURL)

					}
				}

			}

			// Asset list is empty, so we can continue
			if len(assetList) == 0 {

				var tdb, _ = db.NewToddDB(amq.config)
				tdb.SetAgent(agent)

				// This block of code checked that the agent time was within a certain range of the server time. If there was a large enough
				// time skew, the agent advertisement would be rejected.
				// I have disabled this for now - My plan was to use this to synchronize testrun execution amongst agents, but I have
				// a solution to that for now. May revisit this later.
				//
				// Determine difference between server and agent time
				// t1 := time.Now()
				// var diff float64
				// diff = t1.Sub(agent.LocalTime).Seconds()
				//
				// // If Agent is within half a second of server time, add insert to database
				// if diff < 0.5 && diff > -0.5 {
				// } else {
				// 	// We don't want to register an agent if there is a severe time difference,
				// 	// in order to ensure continuity during tests. So, just print log message.
				// 	log.Warn("Agent time not within boundaries.")
				// }

			} else {
				log.Warnf("Agent %s did not have the required asset files. This advertisement is ignored.", agent.UUID)

				var task tasks.DownloadAssetTask
				task.Type = "DownloadAsset" 
				task.Assets = assetList
				amq.SendTask(agent.UUID, task)
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

// SendTask will send a task object onto the specified queue ("queueName"). This could be an agent UUID, or a group name. Agents
// that have been added to a group
func (amq AMQComms) SendTask(queueName string, task tasks.Task) error {

	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect()

	jsonData, err := json.Marshal(task)
	if err != nil {
		log.Error("Failed to marshal object data")
		log.Debug(err)
		return err
	}
	queueName=fmt.Sprintf("/topic/todd/task/%s",queueName)

	err = conn.Send(queueName, "", jsonData)


	log.Debugf("Sent task to %s: %s", queueName, jsonData)

	return nil
}

// ListenForTasks is a method that recieves task notices from the server
func (amq AMQComms) ListenForTasks(uuid string) error {

	// Connect to AMQ with retry logic
	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect()

	queueName:=fmt.Sprintf("/topic/todd/task/%s",uuid)

	sub, _ := conn.Subscribe(queueName, stomp.AckAuto)

	forever := make(chan bool)

	go func() {
		for d := range sub.C {

			// Unmarshal into BaseTaskMessage to determine type
			var baseMsg tasks.BaseTask
			err = json.Unmarshal(d.Body, &baseMsg)

			log.Debugf("Agent task received: %s", d.Body)

			// call agent task method based on type
			switch baseMsg.Type {
			case "DownloadAsset":

				downloadAssetTask := tasks.DownloadAssetTask{
					HTTPClient:   &http.Client{},
					Fs:           tasks.OsFS{},
					Ios:          tasks.IoSys{},
					CollectorDir: fmt.Sprintf("%s/assets/factcollectors", amq.config.LocalResources.OptDir),
					TestletDir:   fmt.Sprintf("%s/assets/testlets", amq.config.LocalResources.OptDir),
				}

				err = json.Unmarshal(d.Body, &downloadAssetTask)

				err = downloadAssetTask.Run(amq.ac)
				if err != nil {
					log.Warning("The KeyValue task failed to initialize")
				}

			case "KeyValue":

				kvTask := tasks.KeyValueTask{
					Config: amq.config,
				}

				err = json.Unmarshal(d.Body, &kvTask)

				err = kvTask.Run(amq.ac)
				if err != nil {
					log.Warning("The KeyValue task failed to initialize")
				}

			case "SetGroup":

				sgTask := tasks.SetGroupTask{
					Config: amq.config,
				}

				err = json.Unmarshal(d.Body, &sgTask)

				err = sgTask.Run(amq.ac)
				if err != nil {
					log.Warning("The SetGroup task failed to initialize")
				}

			case "DeleteTestData":

				dtdtTask := tasks.DeleteTestDataTask{
					Config: amq.config,
				}

				err = json.Unmarshal(d.Body, &dtdtTask)

				err = dtdtTask.Run(amq.ac)
				if err != nil {
					log.Warning("The DeleteTestData task failed to initialize")
				}

			case "InstallTestRun":

				// Retrieve UUID
				uuid, err := amq.ac.GetKeyValue("uuid")
				if err != nil {
					log.Errorf("unable to retrieve UUID: %v", err)
					continue
				}

				itrTask := tasks.InstallTestRunTask{
					Config: amq.config,
				}

				err = json.Unmarshal(d.Body, &itrTask)

				status := "ready"
				err = itrTask.Run(amq.ac)
				if err != nil {
					log.Warning("The InstallTestRun task failed to initialize")
					status = "fail"
				}

				response := responses.NewSetAgentStatus(uuid, itrTask.Tr.UUID, status)

				amq.SendResponse(response)

			case "ExecuteTestRun":

				// Retrieve UUID
				uuid, err := amq.ac.GetKeyValue("uuid")
				if err != nil {
					log.Errorf("unable to retrieve UUID: %v", err)
					continue
				}

				etrTask := tasks.ExecuteTestRunTask{
					Config: amq.config,
				}

				err = json.Unmarshal(d.Body, &etrTask)

				// Send status that the testing has begun, right now.
				response := responses.NewSetAgentStatus(uuid, etrTask.TestUUID, "testing")
				amq.SendResponse(response)

				err = etrTask.Run(amq.ac)
				if err != nil {
					log.Warning("The ExecuteTestRun task failed to initialize")
					response.Status = "fail"
					amq.SendResponse(response)
				}

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received task: %s", baseMsg.Type))
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

// WatchForGroup should be run as a goroutine, like other background services. This is because it will itself spawn a goroutine to
// listen for tasks that are sent to groups, and this goroutine can be restarted when group membership changes
func (amq AMQComms) WatchForGroup() error {
	// dereg is a channel that allows us to instruct the goroutine that's listening for tests to stop. This allows us to re-register to a new command
	dereg := make(chan bool)
rereg:

	group, err := amq.ac.GetKeyValue("group")
	if err != nil {
		return err
	}

	// if the group is nothing, rewrite to "mull". This is being done for now so that we don't have to worry if the goroutine was started or not
	// This way, it's always running, but if the agent is not in a group, it's listening on the "null" queue, which never has anything on it.
	// This is a minor waste of resources on the agent, so TODO(mierdin): you should probably fix this at some point and figure out how to only run
	// the goroutine when needed, but at the same time prevent the dereg channel from blocking unnecessarily in that case
	//
	// This will also handle the cases when the agent first starts up, and the key for this group isn't present in the cache, and therefore is "".
	if group == "" {
		group = "null"
	}

	go func() {
		for {
			err := amq.ListenForGroupTasks(group, dereg)
			if err != nil {
				log.Warn("ListenForGroupTasks reported a failure. Trying again...")
			}
		}
	}()

	// Loop until the unackedGroup flag is set
	for {
		time.Sleep(2 * time.Second)

		// The key "unackedGroup" stores a "true" or "false" to indicate that there has been a group change that we need to acknowledge (handle)
		unackedGroup, err := amq.ac.GetKeyValue("unackedGroup")
		if err != nil {
			log.Warnf("unable to retrieve unackedGroup: %v\n", err)
			continue
		}
		if unackedGroup == "true" {

			// This will kill the underlying goroutine, and in effect stop listening to the old queue.
			dereg <- true

			// Finally, set the "unackedGroup" to indicate that we've acknowledged the group change, and go back to the "rereg" label
			// to re-register onto the new group name
			err := amq.ac.SetKeyValue("unackedGroup", "false")
			if err != nil {
				log.Errorf("logging setting unackedGroup: %v\n", err)
			}
			goto rereg
		}
	}
}

// ListenForGroupTasks is a method that recieves tasks from the server that are intended for groups
func (amq AMQComms) ListenForGroupTasks(groupName string, dereg chan bool) error {

	// Connect to AMQ with retry logic
	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect()

	queueName:=fmt.Sprintf("/topic/todd/task/%s",groupName)

	sub, _ := conn.Subscribe(queueName, stomp.AckAuto)

	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	log.Debug("Agent re-registering onto group queue - ", groupName)

	go func() {
		for d := range sub.C {

			// Unmarshal into BaseTaskMessage to determine type
			var baseMsg tasks.BaseTask
			err = json.Unmarshal(d.Body, &baseMsg)

			log.Debugf("Agent task received: %s", d.Body)

			// call agent task method based on type
			switch baseMsg.Type {

			// This has been removed, as I am moving away from using queues that use the group name.

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received group task: %s", baseMsg.Type))
			}
		}
	}()

	// This will block until something is sent into this channel. This is an indication that we wish to stop listening for
	// new group tasks, ususally because we need to re-register onto a new queue
	<-dereg

	return nil
}

// SendResponse will send a response object onto the statically-defined queue for receiving such messages.
func (amq AMQComms) SendResponse(resp responses.Response) error {
	queueName := "/topic/todd/agentresponses"

	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect() 

	jsonData, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal response data")
		log.Debug(err)
		return err
	}

	err = conn.Send(queueName, "", jsonData)

	if err != nil {
		log.Error("Failed to publish a response onto message queue")
		log.Debug(err)
		return err
	}

	log.Debugf("Sent response to %s: %s", queueName, jsonData)

	return nil
}

// ListenForResponses listens for responses from an agent
func (amq AMQComms) ListenForResponses(stopListeningForResponses *chan bool) error {

	queueName := "/topic/todd/agentresponses"
	
	conn, err := connectAMQ(amq)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to AMQ")
		log.Debug(err)
		return err
	}
	defer conn.Disconnect() 

	sub, _ := conn.Subscribe(queueName, stomp.AckAuto)
	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	tdb, err := db.NewToddDB(amq.config) 
	if err != nil {
		log.Error("Failed to connect to DB")
		log.Debug(err)
		return err
	}

	go func() {
		for d := range sub.C {

			// Unmarshal into BaseResponse to determine type
			var baseMsg responses.Base
			err = json.Unmarshal(d.Body, &baseMsg)
			if err != nil {
				log.Error("Problem unmarshalling baseresponse")
			}

			log.Debugf("Agent response received: %s", d.Body)

			// call agent response method based on type
			switch baseMsg.Type {
			case responses.TypeSetAgentStatus:

				var sasr responses.SetAgentStatus
				err = json.Unmarshal(d.Body, &sasr)
				if err != nil {
					log.Error("Problem unmarshalling AgentStatus")
				}

				log.Debugf("Agent %s is '%s' regarding test %s. Writing to DB.", sasr.AgentUUID, sasr.Status, sasr.TestUUID)
				err := tdb.SetAgentTestStatus(sasr.TestUUID, sasr.AgentUUID, sasr.Status)
				if err != nil {
					log.Errorf("Error writing agent status to DB: %v", err)
				}

			case responses.TypeUploadTestData:

				var utdr responses.UploadTestData
				err = json.Unmarshal(d.Body, &utdr)
				if err != nil {
					log.Error("Problem unmarshalling UploadTestDataResponse")
				}

				err = tdb.SetAgentTestData(utdr.TestUUID, utdr.AgentUUID, utdr.TestData)
				if err != nil {
					log.Error("Problem setting agent test data")
				}

				// Send task to the agent that says to delete the entry
				var dtdt tasks.DeleteTestDataTask
				dtdt.Type = "DeleteTestData" //TODO: This is an extra step. Maybe a factory function for the task could help here?
				dtdt.TestUUID = utdr.TestUUID
				amq.SendTask(utdr.AgentUUID, dtdt)

				// Finally, set the status for this agent in the test to "finished"
				err := tdb.SetAgentTestStatus(dtdt.TestUUID, utdr.AgentUUID, "finished")
				if err != nil {
					log.Errorf("Error writing agent status to DB: %v", err)
				}

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received response: %s", baseMsg.Type))
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-*stopListeningForResponses

	return nil
}

func (amq *AMQComms) setAgentCache(ac *cache.AgentCache) {
	amq.ac = ac
}
