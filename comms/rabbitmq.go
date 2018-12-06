/*
    ToDD commsPackage implementation for RabbitMQ

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package comms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/agent/testing"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
)

const (
	connectRetry = 3
)

// newRabbitMQComms is a factory function that produces a new instance of rabbitMQComms with the configuration
// loaded and ready to be used.
func newRabbitMQComms(cfg config.Config) *rabbitMQComms {
	var rmq rabbitMQComms
	rmq.config = cfg

	rmq.queueURL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rmq.config.Comms.User,
		rmq.config.Comms.Password,
		rmq.config.Comms.Host,
		rmq.config.Comms.Port,
	)

	return &rmq
}

type rabbitMQComms struct {
	config   config.Config
	queueURL string
	ac       *cache.AgentCache
}

// connectRabbitMQ wraps the amqp.Dial function in order to provide connection retry functionality
func connectRabbitMQ(queueURL string) (*amqp.Connection, error) {

	conn, err := amqp.Dial(queueURL)

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
	conn, err := connectRabbitMQ(rmq.queueURL)
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
	jsonData, err := json.Marshal(me)
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
			Body:        []byte(jsonData),
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
func (rmq rabbitMQComms) ListenForAgent(assets assetProvider, assetURLPrefix string) error {

	// TODO(mierdin): does func param need to be a pointer?

	conn, err := amqp.Dial(rmq.queueURL)
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
				log.Warnf("Agent %s did not have the required asset files. This advertisement is ignored.", agent.UUID)
				rmq.SendTask(agent.UUID, tasks.NewDownloadAsset(assetList))
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

	conn, err := amqp.Dial(rmq.queueURL)
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

	jsonData, err := json.Marshal(task)
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
			Body:        []byte(jsonData),
		})
	if err != nil {
		log.Error("Failed to publish a task onto message queue")
		log.Debug(err)
		return err
	}

	log.Debugf("Sent task to %s: %s", queueName, jsonData)

	return nil
}

// ListenForTasks is a method that recieves task notices from the server
func (rmq rabbitMQComms) ListenForTasks(uuid string) error {

	// Connect to RabbitMQ with retry logic
	conn, err := connectRabbitMQ(rmq.queueURL)
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
			log.Debugf("Agent task received: %s", d.Body)

			task, err := tasks.Parse(d.Body)
			if err != nil {
				log.Warning("Error parsing task:", err)
				continue
			}

			err = task.Run(&rmq.config, rmq.ac, rmq.SendResponse)
			if err != nil {
				log.Warning("Error running task:", err)
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

// WatchForGroup should be run as a goroutine, like other background services. This is because it will itself spawn a goroutine to
// listen for tasks that are sent to groups, and this goroutine can be restarted when group membership changes
func (rmq rabbitMQComms) WatchForGroup() error {
	// dereg is a channel that allows us to instruct the goroutine that's listening for tests to stop. This allows us to re-register to a new command
	dereg := make(chan bool)
rereg:

	group, err := rmq.ac.GetKeyValue("group")
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
		unackedGroup, err := rmq.ac.GetKeyValue("unackedGroup")
		if err != nil {
			log.Warnf("unable to retrieve unackedGroup: %v\n", err)
			continue
		}
		if unackedGroup == "true" {

			// This will kill the underlying goroutine, and in effect stop listening to the old queue.
			dereg <- true

			// Finally, set the "unackedGroup" to indicate that we've acknowledged the group change, and go back to the "rereg" label
			// to re-register onto the new group name
			err := rmq.ac.SetKeyValue("unackedGroup", "false")
			if err != nil {
				log.Errorf("logging setting unackedGroup: %v\n", err)
			}
			goto rereg
		}
	}
}

// ListenForGroupTasks is a method that recieves tasks from the server that are intended for groups
func (rmq rabbitMQComms) ListenForGroupTasks(groupName string, dereg chan bool) error {

	// Connect to RabbitMQ with retry logic
	conn, err := connectRabbitMQ(rmq.queueURL)
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
			var baseMsg tasks.BaseTask
			err = json.Unmarshal(d.Body, &baseMsg)
			// TODO(mierdin): Need to handle this error

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
func (rmq rabbitMQComms) SendResponse(resp responses.Response) error {

	queueName := "agentresponses"

	conn, err := amqp.Dial(rmq.queueURL)
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

	jsonData, err := json.Marshal(resp)
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
			Body:        []byte(jsonData),
		})
	if err != nil {
		log.Error("Failed to publish a response onto message queue")
		log.Debug(err)
		return err
	}

	log.Debugf("Sent response to %s: %s", queueName, jsonData)

	return nil
}

// ListenForResponses listens for responses from an agent
func (rmq rabbitMQComms) ListenForResponses(stopListeningForResponses *chan bool) error {

	queueName := "agentresponses"

	conn, err := amqp.Dial(rmq.queueURL)
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
				rmq.SendTask(utdr.AgentUUID, tasks.NewDeleteTestData(utdr.TestUUID))

				// Finally, set the status for this agent in the test to "finished"
				err := tdb.SetAgentTestStatus(utdr.TestUUID, utdr.AgentUUID, testing.StatusFinished)
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

func (rmq *rabbitMQComms) setAgentCache(ac *cache.AgentCache) {
	rmq.ac = ac
}
