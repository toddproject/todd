/*
    ToDD Client API Calls for "todd delete"

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
	"net/http"
)

// Delete will send a request to remove an existing ToDD Object
func (capi ClientApi) Delete(conf map[string]string, objType, objLabel string) error {

	// If insufficient subargs were provided, error out
	if objType == "" || objLabel == "" {
		return errors.New("Error, need to provide type and label (Ex. 'todd delete group datacenter')")
	}

	// anonymous struct to hold our delete info
	deleteinfo := struct {
		Label string `json:"label"`
		Type  string `json:"type"`
	}{
		objLabel,
		objType,
	}

	// Marshal deleteinfo into JSON
	json_str, err := json.Marshal(deleteinfo)
	if err != nil {
		return err
	}

	// Construct API request, and send POST to server for this object
	var url string
	url = fmt.Sprintf("http://%s:%s/v1/object/delete", conf["host"], conf["port"])

	var jsonByte = []byte(json_str)
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

	// Print a regular OK message if object was written successfully - else print the HTTP status code
	if resp.Status == "200 OK" {
		fmt.Println("[OK]")
	} else {
		return errors.New(resp.Status)
	}

	return nil
}
