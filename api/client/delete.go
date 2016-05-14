/*
    ToDD Client API Calls for "todd delete"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Delete will send a request to remove an existing ToDD Object
func (capi ClientApi) Delete(conf map[string]string, objType, objLabel string) {

	// If insufficient subargs were provided, error out
	if objType == "" || objLabel == "" {
		fmt.Println("Error, need to provide type and label (Ex. 'todd delete group datacenter')")
		os.Exit(1)
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
		panic(err)
	}

	// Construct API request, and send POST to server for this object
	var url string
	url = fmt.Sprintf("http://%s:%s/v1/object/delete", conf["host"], conf["port"])

	var jsonByte = []byte(json_str)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print a regular OK message if object was written successfully - else print some debug info
	if resp.Status == "200 OK" {
		fmt.Println("[OK]")
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}

}
