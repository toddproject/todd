/*
   ToDD API - manages todd objects

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/server/objects"
)

// ListObjects will query the database layer for a slice of objects based on the parameters provided in the URL.
// The client may ask for a specific type of object by using the "/all" notation after the object keyword in the URL,
// or this may be a specific type, such as "/group".
func (tapi ToDDApi) ListObjects(w http.ResponseWriter, r *http.Request) {

	// See if the client is trying to list all objects
	if r.URL.String() == "/v1/object/all" {
		log.Warn("/v1/object/all function currently unsupported")
		return
	}

	// Derive specific type from URL
	objType := strings.Split(r.URL.String(), "/")[3]
	objectList, err := tapi.tdb.GetObjects(objType)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
		return
	}

	response, err := json.MarshalIndent(objectList, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(response))
}

// CreateObject will decode a JSON object into a proper ToddObject instance, and send that to the
// database layer to be written persistently.
func (tapi ToDDApi) CreateObject(w http.ResponseWriter, r *http.Request) {

	// Defer the closing of the body
	defer r.Body.Close()

	// Read the content into a byte array
	// (we're doing this so we can access the JSON contents more than once)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// Marshal API data into BaseObject
	var baseobj objects.BaseObject
	err = json.Unmarshal(body, &baseobj)
	if err != nil {
		panic(err)
	}

	// Generate a more specific Todd Object based on the JSON data
	finalobj := baseobj.ParseToddObject(body)

	err = tapi.tdb.SetObject(finalobj)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
		return
	}
}

// DeleteObject will decode a JSON object from a client request to determine the type and label of the
// object that needs to be deleted. Then, it will send this information to the database layer to delete this object.
func (tapi ToDDApi) DeleteObject(w http.ResponseWriter, r *http.Request) {

	deleteInfo := make(map[string]string)

	err := json.NewDecoder(r.Body).Decode(&deleteInfo)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
	}

	err = tapi.tdb.DeleteObject(deleteInfo["label"], deleteInfo["type"])
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
	}
}
