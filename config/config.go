/*
    ToDD Configuration

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package config

import (
	"os"
	"strconv"
)

type ToDDConfig struct {
	OptDir           string
	TestingTimeout   int
	DefaultInterface string
	IPAddrOverride   string
}

func LoadConfigFromEnv() (*ToDDConfig, error) {

	config := ToDDConfig{}

	/*
		REQUIRED
	*/

	// Get configuration parameters from env
	// searchDir := os.Getenv("SYRINGE_LESSONS")
	// if searchDir == "" {
	// 	return nil, errors.New("SYRINGE_LESSONS is a required variable.")
	// } else {
	// 	config.LessonsDir = searchDir
	// }

	/*
		OPTIONAL
	*/
	timeout, err := strconv.Atoi(os.Getenv("TODD_TEST_TIMEOUT"))
	if timeout == 0 || err != nil {
		config.TestingTimeout = 300
	} else {
		config.TestingTimeout = timeout
	}

	optdir := os.Getenv("TODD_OPT_DIR")
	if optdir == "" {
		config.OptDir = "/opt/todd/server"
	} else {
		config.OptDir = optdir
	}

	defif := os.Getenv("TODD_DEFAULT_INTERFACE")
	if defif == "" {
		config.DefaultInterface = "none"
	} else {
		config.DefaultInterface = defif
	}

	ipaddr := os.Getenv("TODD_IPADDR_OVERRIDE")
	if ipaddr == "" {
		config.IPAddrOverride = "none"
	} else {
		config.IPAddrOverride = defif
	}

	return &config, nil

}
