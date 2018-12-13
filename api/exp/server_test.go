package api

import (
	"log"

	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/persistence"
)

var cfg = config.ToDDConfig{
	OptDir:  "/tmp/toddtests",
	ApiPort: 50099,
}

func init() {

	p, err := persistence.NewPersistence(&cfg)
	if err != nil {
		log.Fatalf("Error setting up database: %v\n", err)
	}

	go StartAPI(&cfg, p)
}
