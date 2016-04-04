/*
   This file contains mock infrastructure for the api package

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"io"
	//"net/http"
	"net/http/httptest"
	//"testing"
	//"time"

	"github.com/Mierdin/todd/config"
)

var (
	server    *httptest.Server
	reader    io.Reader //Ignore this for now
	agentsUrl string
)

func init() {
	cfg := config.Config{
		API: config.API{
			Host: "127.0.0.1",
			Port: "8080",
		},
		AMQP: config.AMQP{
			User:     "",
			Password: "",
			Host:     "",
			Port:     "",
		},
		Comms: config.Comms{
			Plugin: "",
		},
		Assets: config.Assets{
			IP:   "",
			Port: "",
		},
		DB: config.DB{
			IP:     "",
			Port:   "",
			Plugin: "",
		},
		TSDB: config.TSDB{
			IP:     "",
			Port:   "",
			Plugin: "",
		},
		Testing: config.Testing{
			Timeout: 30,
		},
		Grouping: config.Grouping{
			Interval: 10, //
		},
		LocalResources: config.LocalResources{
			DefaultInterface: "",
			OptDir:           "",
			IPAddrOverride:   "",
		},
	}

	tapi := ToDDApi{
		cfg: cfg,
	}

	server = httptest.NewServer(handlers(&tapi)) //Creating new server with the user handlers
	// need to close this

	agentsUrl = fmt.Sprintf("%s/v1/agent", server.URL) //Grab the address for the API endpoint

}

// // TestGetAgents is a WIP - example of an integration test we can write to test our API.
// func TestGetAgents(t *testing.T) {

// 	fmt.Println("--------------------")
// 	fmt.Println(agentsUrl)
// 	fmt.Println("--------------------")

// 	time.Sleep(time.Second * 1000)

// 	// userJson := `{"username": "dennis", "balance": 200}`
// 	// reader = strings.NewReader(userJson) //Convert string to reader

// 	request, err := http.NewRequest("GET", agentsUrl, nil) //Create request with JSON body

// 	res, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		t.Error(err) //Something is wrong while sending request
// 	}

// 	if res.StatusCode > 299 {
// 		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
// 	}
// }
