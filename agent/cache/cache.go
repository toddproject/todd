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
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // This look strange but is necessary - the sqlite package is used indirectly by database/sql
	"github.com/pkg/errors"

	"github.com/toddproject/todd/config"
)

// AgentCache provides methods for interacting with the on disk cache.
type AgentCache struct {
	// Need similar abstractions to what you did in the tasks package
	db *sql.DB
}

// New returns an initialized instance of AgentCache.
func New(cfg *config.ToDDConfig) (*AgentCache, error) {

	dbLoc := filepath.Join(cfg.OptDir, "agent_cache.db")

	// Clean up any old cache data
	err := os.Remove(dbLoc)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "removing existing DB file")
	}

	// Open connection
	db, err := sql.Open("sqlite3", dbLoc)
	if err != nil {
		return nil, errors.Wrap(err, "opening DB file")
	}

	// Initialize database
	const sqlStmt = `
		CREATE TABLE testruns (id INTEGER NOT NULL PRIMARY KEY, uuid TEXT, testlet TEXT, args TEXT, targets TEXT, results TEXT);
		CREATE TABLE keyvalue (id INTEGER NOT NULL PRIMARY KEY, key TEXT, value TEXT);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, sqlStmt)
	}

	return &AgentCache{db: db}, nil
}

// Close closes the underlying database connection.
func (ac *AgentCache) Close() error {
	return ac.db.Close()
}
