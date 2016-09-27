/*
    ToDD Agent Cache

    This package manages the local agent cache. This is a handy
    way of storing some locally significant values on the agent.
    As an example, testrun data is stored here so that the agent
    is able to run tests and cache the post-test metrics in this
    cache until the server acknowledges that it received this data.

    It's also useful for storing key/value type data, such as what
    group the agent is in, it's agent UUID, etc. The Init() function
    will delete and recreate this cache, so no data in this cache
    will persist between restarts of the agent.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	"database/sql"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3" // This look strange but is necessary - the sqlite package is used indirectly by database/sql

	"github.com/toddproject/todd/config"
)

func NewAgentCache(cfg config.Config) *AgentCache {
	var ac AgentCache
	ac.db_loc = fmt.Sprintf("%s/agent_cache.db", cfg.LocalResources.OptDir)
	return &ac
}

type AgentCache struct {
	// Need similar abstractions to what you did in the tasks package
	db_loc string
}

// TODO(mierdin): Handling errors in this package?

// Init will set up the sqlite database to serve as this agent cache.
func (ac AgentCache) Init() {

	// Clean up any old cache data
	os.Remove(ac.db_loc)

	// Open connection
	db, err := sql.Open("sqlite3", ac.db_loc)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize database
	sqlStmt := `
    create table testruns (id integer not null primary key, uuid text, testlet text, args text, targets text, results text);
    delete from testruns;
    create table keyvalue (id integer not null primary key, key text, value text);
    delete from keyvalue;
    `

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Errorf("%q: %s\n", err, sqlStmt)
		return
	}
}
