/*
    ToDD Client API Calls for "todd create"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/toddproject/todd/server/objects"
)

// Create is responsible for pushing a ToDD object to the server for eventual storage in whatever database is being used
// It will send a ToddObject rendered as JSON to the "createobject" method of the ToDD API
func (capi ClientAPI) Create(conf map[string]string, yamlFileName string) error {

	// Pull YAML from either stdin or from the filename if stdin is empty
	yamlDef, err := getYAMLDef(yamlFileName)
	if err != nil {
		return err
	}

	// Unmarshal YAML file into a BaseObject so we can peek into the metadata
	var baseobj objects.BaseObject
	err = yaml.Unmarshal(yamlDef, &baseobj)
	if err != nil {
		return errors.New("YAML file not in correct format")
	}

	// finalobj represents the object being created, regardless of type.
	// ToddObject is an interface that satisfies all ToDD objects
	var finalobj objects.ToddObject

	switch baseobj.Type {
	case "group":
		var groupObj objects.GroupObject
		err = yaml.Unmarshal(yamlDef, &groupObj)
		if err != nil {
			return errors.New("Group YAML object not in correct format")
		}
		finalobj = groupObj
	case "testrun":
		var testrunObj objects.TestRunObject
		err = yaml.Unmarshal(yamlDef, &testrunObj)
		if err != nil {
			return errors.New("Testrun YAML object not in correct format")
		}

		if testrunObj.Spec.TargetType == "group" {

			// We need to do a quick conversion because JSON does not support non-string
			// keys, and would reject this during Marshal if we don't.
			stringifiedMap := make(map[string]string)
			for k, v := range testrunObj.Spec.Target.(map[interface{}]interface{}) {
				stringifiedMap[k.(string)] = v.(string)
			}
			testrunObj.Spec.Target = stringifiedMap

		}

		finalobj = testrunObj

	default:
		return errors.New("Invalid object type provided")
	}

	// Marshal the final object into JSON
	jsonByte, err := json.Marshal(finalobj)
	if err != nil {
		return errors.New("Problem marshalling the final object into JSON")
	}

	// Construct API request, and send POST to server for this object
	url := fmt.Sprintf("http://%s:%s/v1/object/create", conf["host"], conf["port"])

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Print a regular OK message if object was created successfully - else print the HTTP status code
	if resp.Status == "200 OK" {
		fmt.Println("[OK]")
	} else {
		return errors.New(resp.Status)
	}

	return nil
}

// getYAMLDef reads YAML from either stdin or from the filename if stdin is empty
func getYAMLDef(yamlFileName string) ([]byte, error) {
	// If stdin is populated, read from that
	if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}

	// Quit if there's nothing on stdin, and there's no arg either
	if yamlFileName == "" {
		return nil, errors.New("Object definition file not provided - please provide via filename or stdin")
	}

	// Read YAML file
	return ioutil.ReadFile(yamlFileName)
}
