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
	"path/filepath"

	cli "github.com/codegangsta/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "todd"

	// TODO(mierdin): autogen like in syringe
	app.Version = "v0.1.0"
	app.Usage = "The Distributed Network-Service-Level Assertion Engine"

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
	// clientAPI.Conf = map[string]string{
	// 	"host": host,
	// 	"port": port,
	// }
	// expApiClient.Conf = map[string]string{
	// 	"host": host,
	// 	"port": port,
	// }

	// ToDD Commands
	// TODO(mierdin): this is quite large. Should consider breaking this up into more manageable chunks
	app.Commands = []cli.Command{

		// "todd group ..."
		// TODO(mierdin) need to document usage of c.Args().First()
		{
			Name:    "group",
			Aliases: []string{"gr", "groups"},
			Usage:   "Work with ToDD groups",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List group definitions",
					Action: func(c *cli.Context) {
						groups, err := ListGroups()
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						} else {

							// Convert to interface slice
							// https://github.com/golang/go/wiki/InterfaceSlice
							// var resourceSlice []api.ToDDResource = make([]api.ToDDResource, len(groups))
							var resourceSlice []interface{}
							for _, d := range groups {
								// resourceSlice[i] = d
								resourceSlice = append(resourceSlice, d)
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
						err := GetGroup(
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
						err := DeleteGroup(
							c.Args().First(),
						)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},

				// TODO(mierdin): Optionally, would be cool if for all "create" CLI functions, you
				// could allow the user to not specify a path arg, in which case they'd be sent to
				// a wizard to assemble it themselves
				{
					Name:      "create",
					Usage:     "Create a new group definition from file",
					ArgsUsage: "<PATH>",
					Action: func(c *cli.Context) {
						absPath, err := filepath.Abs(c.Args().First())
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						group, err := marshalGroupFromFile(absPath)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						err = CreateGroup(group)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
			},
		},
		{
			Name:    "agent",
			Aliases: []string{"ag", "agents"},
			Usage:   "Work with ToDD agents",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List registered agents",
					Action: func(c *cli.Context) {
						agents, err := ListAgents()
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						} else {

							// Convert to interface slice
							// https://github.com/golang/go/wiki/InterfaceSlice
							// var resourceSlice []api.ToDDResource = make([]api.ToDDResource, len(agents))
							var resourceSlice []interface{}
							for _, d := range agents {
								// resourceSlice[i] = d
								resourceSlice = append(resourceSlice, d)
							}

							// Print resources as table to user
							PrintResourcesTable(resourceSlice)

						}
					},
				},
				// {
				// 	Name:  "get",
				// 	Usage: "Retrieve a single group definition",
				// 	Action: func(c *cli.Context) {
				// 		err := GetGroup(
				// 			c.Args().First(),
				// 		)
				// 		if err != nil {
				// 			fmt.Println(err)
				// 			os.Exit(1)
				// 		}
				// 	},
				// },
			},
		},
	}

	app.Run(os.Args)
}
