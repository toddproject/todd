/*
    ToDD Client API Calls for "todd run"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

// Run is responsible for activating an existing testrun object
func (capi ClientApi) Run(conf map[string]string, testrunName string, displayReport, skipConfirm bool) error {

	sourceGroup := conf["sourceGroup"]
	sourceApp := conf["sourceApp"]
	sourceArgs := conf["sourceArgs"]

	// If no subarg was provided, do nothing
	if testrunName == "" {
		return errors.New("Please provide testrun object name to run.")
	}

	if !skipConfirm {
		fmt.Printf("Activate testrun %q? (y/n):", testrunName)
		var userResponse string
		_, err := fmt.Scanln(&userResponse)
		if err != nil {
			return err
		}
		if userResponse != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// anonymous struct to hold our testRun info
	testRunInfo := struct {
		TestRunName string `json:"testRunName"`
		SourceGroup string `json:"sourceGroup"`
		SourceApp   string `json:"sourceApp"`
		SourceArgs  string `json:"sourceArgs"`
	}{
		testrunName,
		sourceGroup,
		sourceApp,
		sourceArgs,
	}

	// Marshal the final object into JSON
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(testRunInfo)
	if err != nil {
		return err
	}
	// Construct API request, and send POST to server for this object
	url := fmt.Sprintf("http://%s:%s/v1/testrun/run", conf["host"], conf["port"])

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Print a regular OK message if object was written successfully - else print the HTTP status code
	if resp.Status != "200 OK" {
		return errors.New(resp.Status)
	}

	serverResponse, _ := ioutil.ReadAll(resp.Body)

	switch string(serverResponse) {
	case "notfound":
		return errors.New("ERROR - Specified testrun object not found.")
	case "invalidtopology":
		return errors.New("ERROR - Not enough agents are in the groups specified by the testrun")
	case "failure":
		return errors.New("ERROR - some kind of error was encountered on the server. Test was not run.")
	}

	testUUID := string(serverResponse)

	fmt.Print("\nRUNNING TEST: ", testUUID)
	fmt.Print("\n\n")

	fmt.Println("(Please be patient while the test finishes...)")

	err = listenForTestStatus(conf)
	if err != nil {
		fmt.Printf("Problem subscribing to testrun updates stream: %s\n", err)
		fmt.Println("Will now watch the testrun metrics API for 45 seconds to see if we get a result that way. Please wait...")
	}

	// Poll for results
	timeout := time.After(45 * time.Second)
	data, err := getRunResult(conf, testUUID)
	for err != nil {
		select {
		case <-timeout:
			return errors.New("Failed to retrieve test data after 45 seconds. Something must be wrong - quitting.")
		default:
			time.Sleep(1 * time.Second)
			data, err = getRunResult(conf, testUUID)
		}
	}

	fmt.Printf("\n\nDone.\n")

	// display it to the user if desired
	if sourceGroup != "" || displayReport {
		var buf bytes.Buffer
		err := json.Indent(&buf, data, "", "  ")
		if err != nil {
			fmt.Printf("error %q: %v\n", string(data), err)
		}
		buf.WriteTo(os.Stdout)
		fmt.Println()
	}

	return nil
}

var errNoTestResult = errors.New("No test result")

// getRunResult collects the results of test run from the server's REST API
func getRunResult(conf map[string]string, testUUID string) ([]byte, error) {
	// Go back and get our testrun data
	url := fmt.Sprintf("http://%s:%s/v1/testdata?testUuid=%s", conf["host"], conf["port"], testUUID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, errNoTestResult
	}

	return ioutil.ReadAll(resp.Body)
}

// listenForTestStatus connects to the server's test event stream and prints the progression
//
// This blocks until all agents have finished or an error occurs.
func listenForTestStatus(conf map[string]string) error {
	retries := 0
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:8081", conf["host"]))

	// If the call to net.Dial produces an error, this loop will execute the call again
	// until "retries" reaches it's configured limit
	for err != nil {
		if retries > 5 {
			return errors.New("Failed to subscribe to test event stream after several retries.")
		}

		retries++
		time.Sleep(1 * time.Second)
		fmt.Println("Failed to subscribe to test event stream. Retrying...")
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:8081", conf["host"]))
	}
	defer conn.Close()

	var recordCount int
	firstMessage := false
	for {

		// listen for reply
		message, err := bufio.NewReader(conn).ReadString('\n')
		// If an error is raised, it's probably because the server killed the connection
		if err != nil {
			// TODO(mierdin): This doesn't really tell us if the connection died because of an error or not
			return errors.New("Disconnected from testrun status stream")
		}

		var statuses map[string]string
		err = json.Unmarshal([]byte(message), &statuses)
		if err != nil {
			return fmt.Errorf("Invalid status from server %q: %v", message, err)
		}

		if !firstMessage {
			recordCount = len(statuses)
			firstMessage = true
		}

		init, ready, testing, finished := 0, 0, 0, 0
		for _, status := range statuses {

			switch status {
			case "init":
				init++
			case "ready":
				ready++
			case "testing":
				testing++
			case "finished":
				finished++
			default:
				return errors.New("Invalid status received.")
			}
		}

		// Print the status line (note the \r which keeps the same line in place on the terminal)
		fmt.Printf(
			"\r %[1]s INIT: (%[3]d/%[2]d)  READY: (%[4]d/%[2]d)  TESTING: (%[5]d/%[2]d)  FINISHED: (%[6]d/%[2]d)",
			time.Now(),
			recordCount,
			init,
			ready,
			testing,
			finished,
		)

		if finished == recordCount {
			break
		}

		// Send an ack back to the server to let it know we're alive
		fmt.Fprintf(conn, "ack\n")
	}

	return nil
}
