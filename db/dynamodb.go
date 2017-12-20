/*
   ToDD databasePackage implementation for AWS dynamoDB

  Inspired from the etcd implementation
*/

package db

import (
	"errors"
	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/server/objects"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"time"
	"encoding/json"
	"strconv"
	"strings"
)

const TTL=time.Duration(30*time.Second);

func newDynamoDB(cfg config.Config) *dynamoDB {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	return &dynamoDB{config: cfg, dyn: svc}
}

type dynamoDB struct {
	config  config.Config
	dyn *dynamodb.DynamoDB
}

type agentTestrun struct{

	Group string
	Status string
	Testdata string
}

type testrun struct {
	Agents map[string]agentTestrun
	Cleandata string
}

func (dynamo *dynamoDB) Init() error {
	return nil
}

func (dynamo *dynamoDB) SetAgent(adv defs.AgentAdvert) error {
	log.Infof("Setting agent %s key", adv.UUID)


	advJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": adv.UUID,
		"objectType": "agent",
		"creationDate": time.Now().Unix(),
		"data": adv,
		"ttl": TTL.Seconds(),
	})

	if err != nil {
		log.Error("Problem converting Agent Advertisement to JSON")
		return err
	}

	err=WriteData(dynamo,"todd",advJSON)

	if err != nil {
		log.Error("Problem setting agent in dynamoDB")
		return err
	}

	log.Infof("Agent Written in db")

	return nil
}


// GetAgent will retrieve a specific agent from the database by UUID
func (dynamo *dynamoDB) GetAgent(uuid string) (*defs.AgentAdvert, error) {

	temp,err:=GetData(dynamo,"agent",uuid)

	if err != nil {
		log.Error("Problem getting agent in dynamoDB")
		return nil, err
	}

	var toReturn defs.AgentAdvert
	dynamodbattribute.Unmarshal(temp.Item["data"],&toReturn)

	//Manually checking for expiration, as dynamoDB delete expired item within 48 hour of expiration, which is not fast enough 
	value,_:=strconv.ParseInt(*(temp.Item["creationDate"].N),10,64)
	toReturn.Expires=time.Until(time.Unix(value+int64(TTL.Seconds()),0))

	if toReturn.Expires < 0{
		RemoveData(dynamo,"agent",*(temp.Item["UUID"].S))
		return nil,nil
	}

	return &toReturn, nil
}


// GetAgents will retrieve all agents from the database
func (dynamo *dynamoDB) GetAgents() ([]defs.AgentAdvert, error) {

	retAdv := []defs.AgentAdvert{}

	object:="agent"
	log.Print("Getting agents' key value")
	_=CreateTable(dynamo,"todd")

	result, err := dynamo.dyn.Query(&dynamodb.QueryInput{
		TableName: aws.String("todd"),
		KeyConditionExpression: aws.String("objectType = :objectType"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":objectType": {
				S: aws.String(object),
			},
		},
	})

	for _,element := range result.Items {
		var temp defs.AgentAdvert;
		dynamodbattribute.Unmarshal(element["data"],&temp)
		//Manually checking for expiration, as dynamoDB delete expired item within 48 hour of expiration, which is not fast enough 

		value,_:=strconv.ParseInt(*(element["creationDate"].N),10,64)
		temp.Expires=time.Until(time.Unix(value+int64(TTL.Seconds()),0))
		if(temp.Expires< 0){
			RemoveData(dynamo,object,*(element["UUID"].S))
			continue
		}
		retAdv=append(retAdv, temp)
	}

	if err != nil {
		return nil, err
	}

	return retAdv, nil

}


// RemoveAgent will delete an agent advertisement present in etcd. This function exists to compensate the fact the dynamoDB delete expired item in a huge time frame 
func (dynamo *dynamoDB) RemoveAgent(adv defs.AgentAdvert) error {


	_, err := RemoveData(dynamo,"agent",adv.UUID)

	log.Infof("Removed agent %s", adv.UUID)
	if err != nil {
		return err
	}

	return nil
}

func (dynamo *dynamoDB) GetObjects(objType string) ([]objects.ToddObject, error) {

	retObj := []objects.ToddObject{}

	// Construct the path to the key depending on the objType param
	if objType == "" {
		return nil, errors.New("Object API queried with no type argument")
	}

	_=CreateTable(dynamo,"todd")

	result, err := dynamo.dyn.Query(&dynamodb.QueryInput{
		TableName: aws.String("todd"),
		KeyConditionExpression: aws.String("objectType = :objectType"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":objectType": {
				S: aws.String("object."+objType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	for _,element := range result.Items {
		var temp objects.BaseObject;
		temp.Type=*(element["objectType"].S)
		temp.Type=strings.Split(temp.Type,".")[1]

		//This convertion is needed to get the todd object from the DynamoDB values
		var returnValue map[string]interface{};
		dynamodbattribute.Unmarshal(element["data"],&returnValue)
		dataByte,_:=json.Marshal(returnValue)

		toddObject:=temp.ParseToddObject(dataByte)
		retObj=append(retObj, toddObject)
	}

	return retObj, nil
}


// SetObject will insert or update a ToddObject within DynamoDB
func (dynamo *dynamoDB) SetObject(tobj objects.ToddObject) error {

	objJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": tobj.GetLabel(),
		"objectType": "object."+tobj.GetType(),
		"creationDate": time.Now().Unix(),
		"data": tobj,
	})

	// Here, we set the objectType as object.{{objectType}}

	err=WriteData(dynamo,"todd",objJSON)
	if err != nil {
		log.Error("Problem setting object in dynamo")
		return err
	}

	log.Infof("Wrote new Todd Object %v %v to dynamo ", tobj.GetType(), tobj.GetLabel())
	return nil
}

func (dynamo *dynamoDB) DeleteObject(label string, objtype string) error {


	_, err := RemoveData(dynamo,"object."+objtype,label)

	log.Infof("Removed %s object", label)
	if err != nil {
		return err
	}

	return nil
}

// SetGroupMapping will update DynamoDB with the results of a grouping calculation
func (dynamo *dynamoDB) SetGroupMap(groupmap map[string]string) error {

	advJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": "groupmap",
		"objectType": "groupmap",
		"creationDate": time.Now().Unix(),
		"data": groupmap,
		"ttl": TTL.Seconds(),
	})
	if err != nil {
		log.Error("Problem converting GroupMap Advertisement to JSON")
		return err
	}

	err=WriteData(dynamo,"todd",advJSON)
	if err != nil {
		log.Error("Problem setting groupmap in DynamoDB")
		return err
	}

	log.Infof("Groupmap Written in db")

	return nil
}


// GetGroupMap returns a map containing agent-to-group mappings. Agent UUIDs are used for keys
func (dynamo *dynamoDB) GetGroupMap() (map[string]string, error) {

	retMap := make(map[string]string)
	temp,err:=GetData(dynamo,"groupmap","groupmap")
	dynamodbattribute.Unmarshal(temp.Item["data"],&retMap)

	if err != nil {
		log.Info("Error loading Groupmap")
		return nil, err
	}

	return retMap, nil
}


// InitTestRun is responsible for initializing a new test run within the database. This includes creating an entry for the test itself
// using the provided UUID for uniqueness, but also in the case of etcd, a nested entry for each agent participating in the test. Each
// Agent entry will be initially populated with that agent's current group and an initial status, but it will also house the result of
// that agent's testrun data, which will be aggregate dafter all agents have checked back in.
func (dynamo *dynamoDB) InitTestRun(testUUID string, testAgentMap map[string]map[string]string) error {


	data:=testrun{
		Agents: make(map[string]agentTestrun),
	}
	for _, uuidmappings := range testAgentMap {

		// _ is either "targets" or "sources".
		// uuidmappings is a map[string]string that contains uuid (key) to group name (value) mappings for this test.

		for agent, group := range uuidmappings {


			data.Agents[agent]=agentTestrun{
				Group: group,
				Status: "init",
			}

		}

	}

	testrunJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": testUUID,
		"objectType": "testrun",
		"creationDate": time.Now().Unix(),
		"data": data,
	})

	err=WriteData(dynamo,"todd",testrunJSON)
	if err != nil {
		log.Error("Problem setting object in dynamo")
		return err
	}

	return nil
}


// SetAgentTestStatus sets the status for an agent in a particular testrun key.
func (dynamo *dynamoDB) SetAgentTestStatus(testUUID, agentUUID, status string) error {

	temp,err:=GetData(dynamo,"testrun",testUUID)
	var toModify testrun
	dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)
	if err != nil {
		log.Error("Problem getting testrun data")
		return err
	}

	if toModify.Agents==nil{
		toModify.Agents=make(map[string]agentTestrun)
	}

	toModify.Agents[agentUUID]=agentTestrun{
		Testdata: toModify.Agents[agentUUID].Testdata,
		Status: status,
		Group: toModify.Agents[agentUUID].Group,
	}

	testrunJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": testUUID,
		"objectType": "testrun",
		"creationDate": time.Now().Unix(),
		"data": toModify,
	})

	err=WriteData(dynamo,"todd",testrunJSON)
	if err != nil {
		log.Error("Problem setting object in etcd")
		return err
	}

	return nil
}


// SetAgentTestData sets the post-test data for an agent in a particular testrun
func (dynamo *dynamoDB) SetAgentTestData(testUUID, agentUUID, testData string) error {
	temp,err:=GetData(dynamo,"testrun",testUUID)
	if err != nil {
		log.Error("Problem getting testrun data")
		return err
	}

	var toModify testrun
	dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)

	if toModify.Agents==nil{
		log.Error("Problem loading testrun data")
		toModify.Agents=make(map[string]agentTestrun)
	}

	toModify.Agents[agentUUID]=agentTestrun{
		Testdata: testData,
		Status: toModify.Agents[agentUUID].Status,
		Group: toModify.Agents[agentUUID].Group,
	}

	testrunJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": testUUID,
		"objectType": "testrun",
		"creationDate": time.Now().Unix(),
		"data": toModify,
	})

	err=WriteData(dynamo,"todd",testrunJSON)
	if err != nil {
		log.Error("Problem setting object in etcd")
		return err
	}

	return nil
}


// GetTestStatus returns a map containing a list of agent UUIDs that are participating in the provided test, and their status in this test.
func (dynamo *dynamoDB) GetTestStatus(testUUID string) (map[string]string, error) {

	toReturn:=make(map[string]string)

	temp,err:=GetData(dynamo,"testrun",testUUID)
	var toModify testrun
	dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)
	if err != nil {
		log.Error("Problem getting testrun data")
		return nil,err
	}


	if toModify.Agents==nil{
		toModify.Agents=make(map[string]agentTestrun)
	}

	for uuid,agent := range toModify.Agents{
		toReturn[uuid]=agent.Status
	}

	return toReturn, nil
}

func (dynamo *dynamoDB) GetAgentTestData(testUUID, sourceGroup string) (map[string]string, error) {

	toReturn:=make(map[string]string)

	temp,err:=GetData(dynamo,"testrun",testUUID)
	var toModify testrun
	dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)
	if err != nil {
		log.Error("Problem getting testrun data")
		return nil,err
	}

	if toModify.Agents==nil{
		toModify.Agents=make(map[string]agentTestrun)
	}

	for uuid,agent := range toModify.Agents{
		toReturn[uuid]=agent.Testdata
	}

	return toReturn, nil
}


// WriteCleanTestData will write the post-test metrics data that has been cleaned up and
// ready to be displayed or exported to the database
func (dynamo *dynamoDB) WriteCleanTestData(testUUID string, testData string) error {
	log.Infof("Writing Cleanestdata")
	temp,err:=GetData(dynamo,"testrun",testUUID)
	var toModify testrun
	dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)
	if err != nil {
		log.Error("Problem getting testrun data")
		return err
	}

	toModify.Cleandata=testData

	testrunJSON, err := dynamodbattribute.Marshal(map [string]interface{}{
		"UUID": testUUID,
		"objectType": "testrun",
		"creationDate": time.Now().Unix(),
		"data": toModify,
	})

	err=WriteData(dynamo,"todd",testrunJSON)
	if err != nil {
		log.Error("Problem setting object in etcd")
		return err
	}

	return nil
}

// GetCleanTestData will retrieve clean test data from the database
func (dynamo *dynamoDB) GetCleanTestData(testUUID string) (string, error) {
	var toModify testrun

	//Retrying until getting data TODO:Put a limit to the number of loops
	for toModify.Cleandata == "" {
		time.Sleep(time.Second * 2)
		temp,err:=GetData(dynamo,"testrun",testUUID)
		log.Infof("Original object: "+temp.Item["data"].GoString())
		dynamodbattribute.Unmarshal(temp.Item["data"],&toModify)
		if err != nil {
			log.Error("Problem getting testrun data")
			return "",err
		}
	}

	var toReturn string
	json.Unmarshal([]byte(toModify.Cleandata),&toReturn)
	log.Infof("Original data: "+toModify.Cleandata)

	return toModify.Cleandata, nil
}


//This is the generic function to write Data. It is used by all other function which needs to write
func WriteData(dynamo *dynamoDB,tableName string, data *dynamodb.AttributeValue) error {
	//Check if the table exist
	_, err :=  dynamo.dyn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err != nil {
		log.Warn("Table "+tableName+" is not exisiting")
		err2:=CreateTable(dynamo,tableName)
		if err2!= nil {
			log.Warn("Table "+tableName+" had trouble getting created")
			return err2
		}
	}

	_,err=dynamo.dyn.PutItem(&dynamodb.PutItemInput{
		Item:  data.M,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(tableName),
	})

	if err != nil {
		log.Error("Trouble writing data")
		return err
	}

	return nil
}

func CreateTable(dynamo *dynamoDB,tableName string) error {
	_, err2 :=  dynamo.dyn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err2 != nil {
		for {
			dynamo.dyn.CreateTable(&dynamodb.CreateTableInput{
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("objectType"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("UUID"),
						KeyType:       aws.String("RANGE"),
					},
				},
				AttributeDefinitions: []*dynamodb.AttributeDefinition{
					{

						AttributeName: aws.String("UUID"),
						AttributeType: aws.String("S"),
					},
					{

						AttributeName: aws.String("objectType"),
						AttributeType: aws.String("S"),
					},
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
				TableName: aws.String(tableName),
			})



			check, _ :=  dynamo.dyn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})


			if (check.Table != nil && *(check.Table.TableStatus)=="ACTIVE"){
				dynamo.dyn.UpdateTimeToLive(&dynamodb.UpdateTimeToLiveInput{
					TableName: aws.String(tableName),

					TimeToLiveSpecification:&dynamodb.TimeToLiveSpecification{
						AttributeName:aws.String("ttl"),
						Enabled:aws.Bool(true),
					},
				})
				break
			}
			time.Sleep(5 * time.Second)

		}


	}

	return err2
}

//This is the generic function to get Data. It is used by all other function which needs to get some data
func GetData(dynamo *dynamoDB,objectType string, uuid string) (*dynamodb.GetItemOutput,error) {

	err2:=CreateTable(dynamo,"todd")
	if err2!= nil {
		return nil,err2
	}

	result,err:=dynamo.dyn.GetItem(&dynamodb.GetItemInput{
		Key:  map[string]*dynamodb.AttributeValue{
			"objectType": {
				S: aws.String(objectType),
			},
			"UUID": {
				S: aws.String(uuid),
			},
		},
		TableName:aws.String("todd"),
	})

	if err!= nil {
		return nil,err
	}
	return result,nil
}

//This is the generic function to remove Data. It is used by all other function which needs to get some remove
func RemoveData(dynamo *dynamoDB,objectType string, uuid string) (*dynamodb.DeleteItemOutput,error) {

	result, err :=  dynamo.dyn.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"objectType": {
				S: aws.String(objectType),
			},
			"UUID": {
				S: aws.String(uuid),
			},
		},
		TableName: aws.String("todd"),
	})

	if err != nil {
		return nil,err
	}
	return result,nil
}

