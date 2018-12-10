/*
    Host Resources - crypto functions

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func GetFileSHA256(filename string) string {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Error generating hash for file %s", filename)
		panic(fmt.Sprintf("Error generating hash for file %s", filename)) // TODO: Use log.Panicf?
	}
	hasher.Write(s)

	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}
