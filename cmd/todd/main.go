/*
    ToDD Client - Primary entrypoint

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"fmt"
	"os"

	cli "github.com/codegangsta/cli"
	capi "github.com/toddproject/todd/api/_old/client"
	api "github.com/toddproject/todd/api/exp"
	expClient "github.com/toddproject/todd/api/exp/client"
)

func main() {

	var clientAPI capi.ClientAPI
	var expApiClient expClient.APIExpClient

	app := cli.NewApp()
	app.Name = "todd"
	app.Version = "v0.1.0"
	app.Usage = "A highly extensible framework for distributed testing on demand"

	var host, port string

	// global level flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "H, host",
			Usage:       "ToDD server hostname",
			Value:       "localhost",
			Destination: &host,
		},
		cli.StringFlag{
			Name:        "P, port",
			Usage:       "ToDD server API port",
			Value:       "8080",
			Destination: &port,
		},
	}

	// TODO(mierdin): This MAY not work. These vars may not execute until after app.Run
	clientAPI.Conf = map[string]string{
		"host": host,
		"port": port,
	}
	expApiClient.Conf = map[string]string{
		"host": host,
		"port": port,
	}

	// ToDD Commands
	// TODO(mierdin): this is quite large. Should consider breaking this up into more manageable chunks
	app.Commands = []cli.Command{

		// "todd agents ..."
		{
			Name:  "agents",
			Usage: "Show ToDD agent information",
			Action: func(c *cli.Context) {
				agents, err := clientAPI.Agents(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
				)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = clientAPI.DisplayAgents(agents, !(c.Args().Get(0) == ""))
				if err != nil {
					fmt.Println("Problem displaying agents (client-side)")
				}
			},
		},

		// "todd create ..."
		{
			Name:  "create",
			Usage: "Create ToDD object (group, testrun, etc.)",
			Action: func(c *cli.Context) {

				err := clientAPI.Create(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
				)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},

		// "todd delete ..."
		{
			Name:  "delete",
			Usage: "Delete ToDD object",
			Action: func(c *cli.Context) {
				err := clientAPI.Delete(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
					c.Args().Get(1),
				)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					fmt.Println("(Are you sure you provided the right object type and/or label?)")
					os.Exit(1)
				}
			},
		},

		// "todd groups ..."
		// TODO(mierdin) need to document usage of c.Args().First()
		{
			Name:    "groups",
			Aliases: []string{"gr"},
			Usage:   "Work with ToDD groups",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List group definitions",
					Action: func(c *cli.Context) {
						err, groups := expApiClient.ListGroups(
							map[string]string{
								"host": host,
								"port": port,
							},
						)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						} else {
							fmt.Println("GOT GROUPS")

							// Convert to interface slice
							// https://github.com/golang/go/wiki/InterfaceSlice
							var resourceSlice []api.ToDDResource = make([]api.ToDDResource, len(groups))
							for i, d := range groups {
								resourceSlice[i] = d
							}

							// Print resources as table to user
							PrintResourcesTable(resourceSlice)

						}
					},
				},
				{
					Name:  "get",
					Usage: "Retrieve a single group definition",
					Action: func(c *cli.Context) {
						err := expApiClient.GetGroup(
							c.Args().First(),
						)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
				{
					Name:  "delete",
					Usage: "Delete a group definition",
					Action: func(c *cli.Context) {
						err := expApiClient.DeleteGroup(
							c.Args().First(),
						)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
				{
					Name:  "create",
					Usage: "Create a new group definition from file",
					Action: func(c *cli.Context) {
						err := expApiClient.CreateGroup(
							c.Args().First(),
						)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
			},
		},

		// "todd run ..."
		{
			Name: "run",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "j",
					Usage: "Output test data for this testrun when finished",
				},
				cli.BoolFlag{
					Name:  "y",
					Usage: "Skip confirmation and run referenced testrun immediately",
				},
				cli.StringFlag{
					Name:  "source-group",
					Usage: "The name of the source group",
				},
				cli.StringFlag{
					Name:  "source-app",
					Usage: "The app to run for this test",
				},
				cli.StringFlag{
					Name:  "source-args",
					Usage: "Arguments to pass to the testlet",
				},
			},
			Usage: "Execute an already uploaded testrun object",
			Action: func(c *cli.Context) {
				err := clientAPI.Run(
					map[string]string{
						"host":        host,
						"port":        port,
						"sourceGroup": c.String("source-group"),
						"sourceApp":   c.String("source-app"),
						"sourceArgs":  c.String("source-args"),
					},
					c.Args().Get(0),
					c.Bool("j"),
					c.Bool("y"),
				)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
