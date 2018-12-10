/*
    ToDD Persistence Layer

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package persistence

import (
	log "github.com/sirupsen/logrus"
	"github.com/toddproject/todd/config"

	"github.com/dgraph-io/badger"
)

func NewPersistence(cfg *config.ToDDConfig) (*Persistence, error) {

	var p Persistence

	opts := badger.DefaultOptions
	opts.Dir = cfg.OptDir
	opts.ValueDir = cfg.OptDir
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	p.db = db

	return &p, nil
}

type Persistence struct {
	config *config.ToDDConfig
	db     *badger.DB
}
