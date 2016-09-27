/*
    ToDD Agent Cache - working with testruns

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3" // This looks strange but is necessary - the sqlite package is used indirectly by database/sql

	"github.com/toddproject/todd/agent/defs"
)

// InsertTestRun places a new test run into the agent cache
func (ac AgentCache) InsertTestRun(tr defs.TestRun) error {

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Error(err)
		return errors.New("Error accessing sqlite cache")
	}
	defer db.Close()

	// Begin Insert
	tx, err := db.Begin()
	if err != nil {
		log.Error(err)
		return errors.New("Error beginning new InsertTestRun action")
	}
	stmt, err := tx.Prepare("insert into testruns(uuid, testlet, args, targets) values(?, ?, ?, ?)")
	if err != nil {
		log.Error(err)
		return errors.New("Error preparing new InsertTestRun action")
	}
	defer stmt.Close()

	// Marshal our string slices to be stored in the database
	json_targets, err := json.Marshal(tr.Targets)
	if err != nil {
		log.Error(err)
		return errors.New("Error marshaling testrun data into JSON")
	}
	json_args, err := json.Marshal(tr.Args)
	if err != nil {
		log.Error(err)
		return errors.New("Error marshaling testrun data into JSON")
	}

	_, err = stmt.Exec(tr.Uuid, tr.Testlet, string(json_args), string(json_targets))
	if err != nil {
		log.Error(err)
		return errors.New("Error executing new testrun insert")
	}

	tx.Commit()

	log.Info("Inserted new testrun into agent cache - ", tr.Uuid)

	return nil
}

// GetTestRun retrieves a testrun from the agent cache by its UUID
func (ac AgentCache) GetTestRun(uuid string) (defs.TestRun, error) {

	var tr defs.TestRun

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Error(err)
		return tr, errors.New("Error accessing sqlite cache")
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select testlet, args, targets from testruns where uuid = \"%s\" ", uuid))
	if err != nil {
		log.Error(err)
		return tr, errors.New("Error creating query for selecting testrun")
	}

	defer rows.Close()
	for rows.Next() {

		// TODO(mierdin): This may be unnecessary - rows.Scan() might allow you to pass this in as a byteslice. Experiment with this
		var args_json, targets_json string

		rows.Scan(&tr.Testlet, &args_json, &targets_json)
		err = json.Unmarshal([]byte(args_json), &tr.Args)
		if err != nil {
			log.Error(err)
			return tr, errors.New("Error unmarshaling testrun data from JSON")
		}
		err = json.Unmarshal([]byte(targets_json), &tr.Targets)
		if err != nil {
			log.Error(err)
			return tr, errors.New("Error unmarshaling testrun data from JSON")
		}
	}

	tr.Uuid = uuid

	log.Info("Found testrun ", tr.Uuid, " running testlet ", tr.Testlet)

	return tr, nil
}

// UpdateTestRunData will update an existing testrun entry in the agent cache with the post-test
// metrics dataset that corresponds to that testrun (by testrun UUID)
func (ac AgentCache) UpdateTestRunData(uuid string, testData string) error {

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Error(err)
		return errors.New("Error accessing sqlite cache for testrun update")
	}
	defer db.Close()

	// Begin Update
	tx, err := db.Begin()
	if err != nil {
		log.Error(err)
		return errors.New("Error beginning new UpdateTestRunData action")
	}

	stmt, err := tx.Prepare(fmt.Sprintf("update testruns set results = '%s' where uuid = '%s' ", testData, uuid))
	if err != nil {
		log.Error(err)
		return errors.New("Error preparing new UpdateTestRunData action")
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		log.Error(err)
		return errors.New("Error executing new UpdateTestRunData action")
	}
	tx.Commit()

	log.Infof("Inserted test data for %s into cache", uuid)

	return nil
}

// DeleteTestRun will remove an entire testrun entry from teh agent cache by UUID
func (ac AgentCache) DeleteTestRun(uuid string) error {

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Error(err)
		return errors.New("Error accessing sqlite cache for DeleteTestRun")
	}
	defer db.Close()

	// Begin Update
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return errors.New("Error beginning new DeleteTestRun action")
	}

	stmt, err := tx.Prepare(fmt.Sprintf("delete from testruns where uuid = \"%s\" ", uuid))
	if err != nil {
		log.Fatal(err)
		return errors.New("Error preparing new DeleteTestRun action")
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
		return errors.New("Error executing new DeleteTestRun action")
	}
	tx.Commit()

	return nil
}

// GetFinishedTestRuns returns a map of test UUIDS (keys) and the corresponding post-test metric data for those UUIDs (values)
// The metric data is stored as a string containing JSON text, so this is what's placed into this map (meaning JSON parsing is
// not performed in this function)
func (ac AgentCache) GetFinishedTestRuns() (map[string]string, error) {

	retmap := make(map[string]string)

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Error(err)
		return retmap, errors.New("Error accessing sqlite cache for finished testruns")
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprint("select uuid, results from testruns where results != \"\" "))
	if err != nil {
		log.Error(err)
		return retmap, errors.New("Error creating query for selecting finished testruns")
	}

	defer rows.Close()
	for rows.Next() {
		var uuid, testdata string
		rows.Scan(&uuid, &testdata)
		log.Debug("Found ripe testrun: ", uuid)
		retmap[uuid] = testdata
	}

	return retmap, nil
}
