/*
    ToDD Client - Primary entrypoint

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package main

import (
	"fmt"
	"os"

	capi "github.com/Mierdin/todd/api/client"
	cli "github.com/codegangsta/cli"
)

func main() {

	var clientapi capi.ClientApi

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

	// ToDD Commands
	app.Commands = []cli.Command{

		// "todd agents ..."
		{
			Name:  "agents",
			Usage: "Show ToDD agent information",
			Action: func(c *cli.Context) {
				agents, err := clientapi.Agents(
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
				clientapi.DisplayAgents(agents, !(c.Args().Get(0) == ""))
			},
		},

		// "todd create ..."
		{
			Name:  "create",
			Usage: "Create ToDD object (group, testrun, etc.)",
			Action: func(c *cli.Context) {

				err := clientapi.Create(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
				)
				if err != nil {
					fmt.Println("Unable to create object on ToDD server.")
					os.Exit(1)
				}
			},
		},

		// "todd delete ..."
		{
			Name:  "delete",
			Usage: "Delete ToDD object",
			Action: func(c *cli.Context) {
				err := clientapi.Delete(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
					c.Args().Get(1),
				)
				if err != nil {
					fmt.Println("ERROR - Are you sure you provided the right object type and/or label?")
					os.Exit(1)
				}
			},
		},

		// "todd groups ..."
		{
			Name:  "groups",
			Usage: "Show current agent-to-group mappings",
			Action: func(c *cli.Context) {
				err := clientapi.Groups(
					map[string]string{
						"host": host,
						"port": port,
					},
				)
				if err != nil {
					fmt.Println("ERROR - Unable to obtain current group mapping")
					os.Exit(1)
				}
			},
		},

		// "todd objects ..."
		{
			Name:  "objects",
			Usage: "Show information about installed group objects",
			Action: func(c *cli.Context) {
				err := clientapi.Objects(
					map[string]string{
						"host": host,
						"port": port,
					},
					c.Args().Get(0),
				)
				if err != nil {
					fmt.Println("ERROR - Unable to retrieve object")
					os.Exit(1)
				}
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
				err := clientapi.Run(
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
					fmt.Println("ERROR - Problem running testrun")
					os.Exit(1)
				}

			},
		},
	}

	app.Run(os.Args)
}
