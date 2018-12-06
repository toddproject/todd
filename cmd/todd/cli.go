/*
	Helper functions for the CLI

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"

	api "github.com/toddproject/todd/api/exp"
	pb "github.com/toddproject/todd/api/exp/generated"
)

// PrintResourcesTable takes an API resource and pretty-prints it to a table, regardless of its fields.
// use this for general resources that don't need any special output treatment.
func PrintResourcesTable(resources []api.ToDDResource) error {
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	// print headers for the given resource (using the first one as a sample)
	headers, err := getResourceFields(resources[0])
	if err != nil {
		return err
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// print values for each resource
	for i := range resources {
		values, err := getResourceValues(resources[i])
		if err != nil {
			return err
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	fmt.Fprintln(w)
	w.Flush()

	return nil
}

// getResourceFields retrieves a resource's field names and returns them as a slice of strings
//
// TODO (mierdin): Should consider just adding a field to each resource definition to keep track of the headers that are useful
// in a list format. That way we don't have to do this reflect. print only the headers the object says are useful in a "list" operation
// and print verbosely in a "get" operation, just like st2 does.
func getResourceFields(resource api.ToDDResource) ([]string, error) {
	var retSlice []string
	val := reflect.ValueOf(resource).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		retSlice = append(retSlice, strings.ToUpper(typeField.Name))
	}
	return retSlice, nil
}

// getResourceValues retrieves a resource's field values and returns them as a slice of strings
func getResourceValues(resource api.ToDDResource) ([]string, error) {
	var retSlice []string
	val := reflect.ValueOf(resource).Elem()
	for i := 0; i < val.NumField(); i++ {
		// typeField := val.Type().Field(i)
		valueField := val.Field(i)
		retSlice = append(retSlice, fmt.Sprintf("%v", valueField.Interface()))
	}
	return retSlice, nil
}

// marshalGroupFromFile creates a new Group instance from a file definition
func marshalGroupFromFile(absPath string) (*pb.Group, error) {
	yamlDef, _ := getYAMLDef(absPath)

	var groupObj *pb.Group
	err := yaml.Unmarshal(yamlDef, &groupObj)
	if err != nil {
		return nil, err
	}

	return groupObj, nil
}

// getYAMLDef reads YAML from either stdin or from the filename if stdin is empty
func getYAMLDef(yamlFileName string) ([]byte, error) {
	// If stdin is populated, read from that
	if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}

	// Quit if there's nothing on stdin, and there's no arg either
	if yamlFileName == "" {
		return nil, errors.New("Object definition file not provided - please provide via filename or stdin")
	}

	// Read YAML file
	return ioutil.ReadFile(yamlFileName)
}
