/*
    ToDD Client API Calls for "todd objects"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"

	// TODO need to fix this, I think it's part of stdlib now so you could kill all vendored deps for this

	"github.com/toddproject/todd/server/objects"
)

// Objects will query ToDD for all objects, with the type requested in the sub-arguments, and then display a list of those
// objects to the user.
func (capi ClientAPI) Objects(conf map[string]string, objType string) error {

	// If no subarg was provided, instruct the user to provide the object type
	if objType == "" {
		return errors.New("Please provide the object type")
	}

	url := fmt.Sprintf("http://%s:%s/v1/object/%s", conf["host"], conf["port"], objType)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parsedObjects := objects.ParseToddObjects(body)

	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "LABEL\tTYPE\tSPEC\t")

	for i := range parsedObjects {

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\n",
			parsedObjects[i].GetLabel(),
			parsedObjects[i].GetType(),
			parsedObjects[i].GetSpec(),
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil

}
