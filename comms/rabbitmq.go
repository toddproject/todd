/*
    ToDD commsPackage implementation for RabbitMQ

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package comms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Mierdin/todd/agent/cache"
	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/agent/responses"
	"github.com/Mierdin/todd/agent/tasks"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/db"
	"github.com/Mierdin/todd/hostresources"
	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	connectRetry = 3
)

// newRabbitMQComms is a factory function that produces a new instance of rabbitMQComms with the configuration
// loaded and ready to be used.
func newRabbitMQComms(cfg config.Config) *rabbitMQComms {
	var rmq rabbitMQComms
	rmq.config = cfg

	rmq.queueUrl = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rmq.config.Comms.User,
		rmq.config.Comms.Password,
		rmq.config.Comms.Host,
		rmq.config.Comms.Port,
	)

	return &rmq
}

type rabbitMQComms struct {
	config   config.Config
	queueUrl string
}

// connectRabbitMQ wraps the amqp.Dial function in order to provide connection retry functionality
func connectRabbitMQ(queueUrl string) (*amqp.Connection, error) {

	conn, err := amqp.Dial(queueUrl)

	for retries := 0; err != nil; {
		if retries > connectRetry {
			return nil, err
		}

		retries++
		log.Warnf("Failure connecting to RabbitMQ - retry #%d", retries)
		time.Sleep(1 * time.Second)
	}

	return conn, nil
}

// AdvertiseAgent will place an agent advertisement message on the message queue
func (rmq rabbitMQComms) AdvertiseAgent(me defs.AgentAdvert) error {

	// Connect to RabbitMQ with retry logic
	conn, err := connectRabbitMQ(rmq.queueUrl)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange", // name
		"direct",        // kind
		false,           // durable
		false,           // delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to declare an exchange")
		log.Debug(err)
		return err
	}

	q, err := ch.QueueDeclare(
		"agentadvert", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	err = ch.QueueBind(
		"agentadvert",   // name
		"agentadvert",   // routing key
		"test_exchange", // exchange
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to bind exchange to queue")
		log.Debug(err)
		return err
	}

	// Marshal agent struct to JSON
	json_data, err := json.Marshal(me)
	if err != nil {
		log.Error("Failed to marshal agent data from queue")
		log.Debug(err)
		return err
	}

	err = ch.Publish(
		"test_exchange", // exchange
		q.Name,          // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Expiration:  "5000", // expiration in milliseconds (we don't want these messages to pile up if the server isn't running)
			Body:        []byte(json_data),
		})
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
func (rmq rabbitMQComms) ListenForAgent(assets map[string]map[string]string) error {

	// TODO(mierdin): does func param need to be a pointer?

	conn, err := amqp.Dial(rmq.queueUrl)
	if err != nil {
		log.Error("(ListenForAgent) Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"agentadvert", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	msgs, err := ch.Consume(
		q.Name,        // queue
		"agentadvert", // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Debugf("Agent advertisement recieved: %s", d.Body)

			var agent defs.AgentAdvert
			err = json.Unmarshal(d.Body, &agent)
			// TODO(mierdin): Need to handle this error

			// assetList is a slice that will contain any URLs that need to be sent to an
			// agent as a response to an incorrect or incomplete list of assets
			var assetList []string

			// assets is the asset map from the SERVER's perspective
			for asset_type, asset_hashes := range assets {

				var agentAssets map[string]string

				// agentAssets is the asset map from the AGENT's perspective
				if asset_type == "factcollectors" {
					agentAssets = agent.FactCollectors
				} else if asset_type == "testlets" {
					agentAssets = agent.Testlets
				}

				for name, hash := range asset_hashes {

					// See if the hashes match (a missing asset will also result in False)
					if agentAssets[name] != hash {

						// hashes do not match, so we need to append the asset download URL to the remediate list
						var default_ip string
						if rmq.config.LocalResources.IPAddrOverride != "" {
							default_ip = rmq.config.LocalResources.IPAddrOverride
						} else {
							default_ip = hostresources.GetIPOfInt(rmq.config.LocalResources.DefaultInterface).String()
						}
						asset_url := fmt.Sprintf("http://%s:%s/%s/%s", default_ip, rmq.config.Assets.Port, asset_type, name)

						assetList = append(assetList, asset_url)

					}
				}

			}

			// Asset list is empty, so we can continue
			if len(assetList) == 0 {

				var tdb, _ = db.NewToddDB(rmq.config)
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
				log.Warnf("Agent %s did not have the required asset files. This advertisement is ignored.", agent.Uuid)

				var task tasks.DownloadAssetTask
				task.Type = "DownloadAsset" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?
				task.Assets = assetList
				rmq.SendTask(agent.Uuid, task)
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

// SendTask will send a task object onto the specified queue ("queueName"). This could be an agent UUID, or a group name. Agents
// that have been added to a group
func (rmq rabbitMQComms) SendTask(queueName string, task tasks.Task) error {

	conn, err := amqp.Dial(rmq.queueUrl)
	if err != nil {
		log.Error("(ListenForAgent) Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange", // name
		"direct",        // kind
		false,           // durable
		false,           // delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to declare an exchange")
		log.Debug(err)
		return err
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	err = ch.QueueBind(
		queueName,       // name
		queueName,       // routing key
		"test_exchange", // exchange
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to bind exchange to queue")
		log.Debug(err)
		return err
	}

	json_data, err := json.Marshal(task)
	if err != nil {
		log.Error("Failed to marshal object data")
		log.Debug(err)
		return err
	}

	err = ch.Publish(
		"test_exchange", // exchange
		queueName,       // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(json_data),
		})
	if err != nil {
		log.Error("Failed to publish a task onto message queue")
		log.Debug(err)
		return err
	}

	log.Debugf("Sent task to %s: %s", queueName, json_data)

	return nil
}

// ListenForTasks is a method that recieves task notices from the server
func (rmq rabbitMQComms) ListenForTasks(uuid string) error {

	// Connect to RabbitMQ with retry logic
	conn, err := connectRabbitMQ(rmq.queueUrl)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to RabbitMQ")
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		os.Exit(1)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		uuid,  // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		os.Exit(1)
	}

	msgs, err := ch.Consume(
		uuid,  // queue
		uuid,  // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error("Failed to register a consumer")
		os.Exit(1)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			// Unmarshal into BaseTaskMessage to determine type
			var base_msg tasks.BaseTask
			err = json.Unmarshal(d.Body, &base_msg)
			// TODO(mierdin): Need to handle this error

			log.Debugf("Agent task received: %s", d.Body)

			// call agent task method based on type
			switch base_msg.Type {
			case "DownloadAsset":

				downloadAssetTask := tasks.DownloadAssetTask{
					HTTPClient:   &http.Client{},
					Fs:           tasks.OsFS{},
					Ios:          tasks.IoSys{},
					CollectorDir: fmt.Sprintf("%s/assets/factcollectors", rmq.config.LocalResources.OptDir),
					TestletDir:   fmt.Sprintf("%s/assets/testlets", rmq.config.LocalResources.OptDir),
				}

				err = json.Unmarshal(d.Body, &downloadAssetTask)
				// TODO(mierdin): Need to handle this error

				err = downloadAssetTask.Run()
				if err != nil {
					log.Warning("The KeyValue task failed to initialize")
				}

			case "KeyValue":

				kv_task := tasks.KeyValueTask{
					Config: rmq.config,
				}

				err = json.Unmarshal(d.Body, &kv_task)
				// TODO(mierdin): Need to handle this error

				err = kv_task.Run()
				if err != nil {
					log.Warning("The KeyValue task failed to initialize")
				}

			case "SetGroup":

				sg_task := tasks.SetGroupTask{
					Config: rmq.config,
				}

				err = json.Unmarshal(d.Body, &sg_task)
				// TODO(mierdin): Need to handle this error

				err = sg_task.Run()
				if err != nil {
					log.Warning("The SetGroup task failed to initialize")
				}

			case "DeleteTestData":

				dtdt_task := tasks.DeleteTestDataTask{
					Config: rmq.config,
				}

				err = json.Unmarshal(d.Body, &dtdt_task)
				// TODO(mierdin): Need to handle this error

				err = dtdt_task.Run()
				if err != nil {
					log.Warning("The DeleteTestData task failed to initialize")
				}

			case "InstallTestRun":

				// Retrieve UUID
				var ac = cache.NewAgentCache(rmq.config)
				uuid := ac.GetKeyValue("uuid")

				itr_task := tasks.InstallTestRunTask{
					Config: rmq.config,
				}

				err = json.Unmarshal(d.Body, &itr_task)
				// TODO(mierdin): Need to handle this error

				var response responses.SetAgentStatusResponse
				response.Type = "AgentStatus" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?
				response.AgentUuid = uuid
				response.TestUuid = itr_task.Tr.Uuid

				err = itr_task.Run()
				if err != nil {
					log.Warning("The InstallTestRun task failed to initialize")
					response.Status = "fail"
				} else {
					response.Status = "ready"
				}
				rmq.SendResponse(response)

			case "ExecuteTestRun":

				// Retrieve UUID
				var ac = cache.NewAgentCache(rmq.config)
				uuid := ac.GetKeyValue("uuid")

				etr_task := tasks.ExecuteTestRunTask{
					Config: rmq.config,
				}

				err = json.Unmarshal(d.Body, &etr_task)
				// TODO(mierdin): Need to handle this error

				// Send status that the testing has begun, right now.
				response := responses.SetAgentStatusResponse{
					TestUuid: etr_task.TestUuid,
					Status:   "testing",
				}
				response.AgentUuid = uuid     // TODO(mierdin): Can't declare this in the literal, it's that embedding behavior again. Need to figure this out.
				response.Type = "AgentStatus" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?
				rmq.SendResponse(response)

				err = etr_task.Run()
				if err != nil {
					log.Warning("The ExecuteTestRun task failed to initialize")
					response.Status = "fail"
					rmq.SendResponse(response)
				}

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received task: %s", base_msg.Type))
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

// WatchForGroup should be run as a goroutine, like other background services. This is because it will itself spawn a goroutine to
// listen for tasks that are sent to groups, and this goroutine can be restarted when group membership changes
func (rmq rabbitMQComms) WatchForGroup() {

	var ac = cache.NewAgentCache(rmq.config)

	// dereg is a channel that allows us to instruct the goroutine that's listening for tests to stop. This allows us to re-register to a new command
	dereg := make(chan bool)
rereg:

	group := ac.GetKeyValue("group")

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
			err := rmq.ListenForGroupTasks(group, dereg)
			if err != nil {
				log.Warn("ListenForGroupTasks reported a failure. Trying again...")
			}
		}
	}()

	// Loop until the unackedGroup flag is set
	for {
		time.Sleep(2 * time.Second)

		// The key "unackedGroup" stores a "true" or "false" to indicate that there has been a group change that we need to acknowledge (handle)
		if ac.GetKeyValue("unackedGroup") == "true" {

			// This will kill the underlying goroutine, and in effect stop listening to the old queue.
			dereg <- true

			// Finally, set the "unackedGroup" to indicate that we've acknowledged the group change, and go back to the "rereg" label
			// to re-register onto the new group name
			ac.SetKeyValue("unackedGroup", "false")
			goto rereg
		}
	}

}

// ListenForGroupTasks is a method that recieves tasks from the server that are intended for groups
func (rmq rabbitMQComms) ListenForGroupTasks(groupName string, dereg chan bool) error {

	// Connect to RabbitMQ with retry logic
	conn, err := connectRabbitMQ(rmq.queueUrl)
	if err != nil {
		log.Error("(AdvertiseAgent) Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		groupName, // name
		false,     // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	msgs, err := ch.Consume(
		groupName, // queue
		groupName, // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	log.Debug("Agent re-registering onto group queue - ", groupName)

	go func() {
		for d := range msgs {

			// Unmarshal into BaseTaskMessage to determine type
			var base_msg tasks.BaseTask
			err = json.Unmarshal(d.Body, &base_msg)
			// TODO(mierdin): Need to handle this error

			log.Debugf("Agent task received: %s", d.Body)

			// call agent task method based on type
			switch base_msg.Type {

			// This has been removed, as I am moving away from using queues that use the group name.

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received group task: %s", base_msg.Type))
			}
		}
	}()

	// This will block until something is sent into this channel. This is an indication that we wish to stop listening for
	// new group tasks, ususally because we need to re-register onto a new queue
	<-dereg

	return nil
}

// SendResponse will send a response object onto the statically-defined queue for receiving such messages.
func (rmq rabbitMQComms) SendResponse(resp responses.Response) error {

	queueName := "agentresponses"

	conn, err := amqp.Dial(rmq.queueUrl)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange", // name
		"direct",        // kind
		false,           // durable
		false,           // delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to declare an exchange")
		log.Debug(err)
		return err
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	err = ch.QueueBind(
		queueName,       // name
		queueName,       // routing key
		"test_exchange", // exchange
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Error("Failed to bind exchange to queue")
		log.Debug(err)
		return err
	}

	json_data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal response data")
		log.Debug(err)
		return err
	}

	err = ch.Publish(
		"test_exchange", // exchange
		queueName,       // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(json_data),
		})
	if err != nil {
		log.Error("Failed to publish a response onto message queue")
		log.Debug(err)
		return err
	}

	log.Debugf("Sent response to %s: %s", queueName, json_data)

	return nil
}

// ListenForResponses listens for responses from an agent
func (rmq rabbitMQComms) ListenForResponses(stopListeningForResponses *chan bool) error {

	queueName := "agentresponses"

	conn, err := amqp.Dial(rmq.queueUrl)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ")
		log.Debug(err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Error("Failed to declare a queue")
		log.Debug(err)
		return err
	}

	msgs, err := ch.Consume(
		queueName, // queue
		queueName, // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Error("Failed to register a consumer")
		log.Debug(err)
		return err
	}

	tdb, err := db.NewToddDB(rmq.config) // TODO(vcabbage): Consider moving this into the rabbitMQComms struct
	if err != nil {
		log.Error("Failed to connect to DB")
		log.Debug(err)
		return err
	}

	go func() {
		for d := range msgs {

			// Unmarshal into BaseResponse to determine type
			var base_msg responses.BaseResponse
			err = json.Unmarshal(d.Body, &base_msg)
			// TODO(mierdin): Need to handle this error

			log.Debugf("Agent response received: %s", d.Body)

			// call agent response method based on type
			switch base_msg.Type {
			case "AgentStatus":

				var sasr responses.SetAgentStatusResponse
				err = json.Unmarshal(d.Body, &sasr)
				// TODO(mierdin): Need to handle this error

				log.Debugf("Agent %s is '%s' regarding test %s. Writing to DB.", sasr.AgentUuid, sasr.Status, sasr.TestUuid)
				err := tdb.SetAgentTestStatus(sasr.TestUuid, sasr.AgentUuid, sasr.Status)
				if err != nil {
					log.Errorf("Error writing agent status to DB: %v", err)
				}

			case "TestData":

				var utdr responses.UploadTestDataResponse
				err = json.Unmarshal(d.Body, &utdr)
				// TODO(mierdin): Need to handle this error

				err = tdb.SetAgentTestData(utdr.TestUuid, utdr.AgentUuid, utdr.TestData)
				// TODO(mierdin): Need to handle this error

				// Send task to the agent that says to delete the entry
				var dtdt tasks.DeleteTestDataTask
				dtdt.Type = "DeleteTestData" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?
				dtdt.TestUuid = utdr.TestUuid
				rmq.SendTask(utdr.AgentUuid, dtdt)

				// Finally, set the status for this agent in the test to "finished"
				err := tdb.SetAgentTestStatus(dtdt.TestUuid, utdr.AgentUuid, "finished")
				if err != nil {
					log.Errorf("Error writing agent status to DB: %v", err)
				}

			default:
				log.Errorf(fmt.Sprintf("Unexpected type value for received response: %s", base_msg.Type))
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-*stopListeningForResponses

	return nil
}
