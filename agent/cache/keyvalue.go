/*
    ToDD Agent Cache - working with keyvalues

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	"fmt"

	"database/sql"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3" // This look strange but is necessary - the sqlite package is used indirectly by database/sql
)

// GetKeyValue will retrieve a value from the agent cache using a key string
func (ac AgentCache) GetKeyValue(key string) string {
	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Debugf("Retrieving value of key - %s", key)

	// First, see if the key exists.
	rows, err := db.Query(fmt.Sprintf("select value from keyvalue where key = \"%s\" ", key))
	if err != nil {
		log.Fatal(err)
	}
	value := ""
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&value)
	}
	return value
}

// SetKeyValue sets a KeyValue pair within the agent cache
func (ac AgentCache) SetKeyValue(key, value string) error {

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Debugf("Writing keyvalue pair to agent cache - %s:%s", key, value)

	// First, see if the key exists.
	rows, err := db.Query(fmt.Sprintf("select key, value FROM keyvalue WHERE KEY = \"%s\";", key))
	if err != nil {
		log.Fatal(err)
	}
	rowcount := 0
	defer rows.Close()
	for rows.Next() {
		rowcount++
	}

	if rowcount != 1 {

		// If there is MORE than one row, we should delete the extras first
		// TODO(mierdin): Is this really necessary?
		if rowcount > 1 {
			log.Warn("Extra keyvalue pair detected. Deleting and inserting new record.")
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			stmt, err := tx.Prepare(fmt.Sprintf("DELETE FROM keyvalue WHERE KEY = \"%s\";", key))
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()
			_, err = stmt.Exec()
			if err != nil {
				log.Fatal(err)
			}
		}

		// Begin Insert
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := tx.Prepare("insert into keyvalue(key, value) values(?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(key, value)
		if err != nil {
			log.Fatal(err)
		}
		tx.Commit()

	} else {

		// Begin Update
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err := tx.Prepare(fmt.Sprintf("update keyvalue set value = \"%s\" where key = \"%s\" ", value, key))
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec()
		if err != nil {
			log.Fatal(err)
		}
		tx.Commit()

	}
	return nil
}
