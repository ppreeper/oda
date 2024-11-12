/*
Copyright © 2024 Peter Preeper <ppreeper@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"embed"
	// _ "embed"
	"fmt"
	"log"
	"os"

	"github.com/ppreeper/oda/internal"
	"github.com/ppreeper/oda/ui"
	"github.com/urfave/cli/v2"
)

//go:generate sh -c "printf '%s (%s)' $(git tag -l | sort -V | tail -1) $(date +%Y%m%d)-$(git rev-parse --short HEAD)"
//go:embed commit.txt
var commit string

//go:embed templates/*
var templates embed.FS

func main() {
	oda := internal.NewODA("oda", "Odoo Client Administration Tool", commit, templates)

	var restoreAny, restoreMove, restoreNeutralize bool

	app := &cli.App{
		Name:                 oda.Name,
		Usage:                oda.Usage,
		Version:              oda.Version,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			// ####################################
			// Admin User Management
			// admin       Admin user management
			{
				Name:     "admin",
				Usage:    "Admin user management",
				Category: "Admin User Management",
				Subcommands: []*cli.Command{
					{
						Name:  "updateuser",
						Usage: "Odoo Update User",
						Action: func(cCtx *cli.Context) error {
							return oda.UpdateUser()
						},
					},
				},
			},
			// ####################################
			// App Management
			//   install     Install module(s)
			{
				Name:     "install",
				Usage:    "Install module(s)",
				Category: "App Management",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no modules specified")
					}
					// calls into "incus exec <project> odoo-bin install <module>"
					return oda.InstanceAppInstallUpgrade(true, cCtx.Args().Slice()...)
				},
			},
			//   upgrade     Upgrade module(s)
			{
				Name:     "upgrade",
				Usage:    "Upgrade module(s)",
				Category: "App Management",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no modules specified")
					}
					// calls into "incus exec <project> odoo-bin upgrade <module>"
					return oda.InstanceAppInstallUpgrade(false, cCtx.Args().Slice()...)
				},
			},
			//   scaffold    Generates an Odoo module skeleton in addons
			{
				Name:     "scaffold",
				Usage:    "Scaffold module",
				Category: "App Management",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no modules specified")
					}
					// calls into "incus exec <project> odoo-bin scaffold <module>"
					return oda.Scaffold(cCtx.Args().First())
				},
			},
			// ####################################
			// Backup Management
			//   backup      Backup database filestore and addons
			{
				Name:     "backup",
				Usage:    "Backup database filestore and addons",
				Category: "Backup Management",
				Action: func(cCtx *cli.Context) error {
					// calls into "incus exec <project> odas backup"
					return oda.Backup()
				},
			},
			//   restore     Restore database and filestore or addons
			{
				Name:     "restore",
				Usage:    "Restore database and filestore or addons",
				Category: "Backup Management",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "any",
						Value:       false,
						Usage:       "lookup any backup",
						Destination: &restoreAny,
					},
					&cli.BoolFlag{
						Name:        "move",
						Value:       false,
						Usage:       "move database",
						Destination: &restoreMove,
					},
					&cli.BoolFlag{
						Name:        "neutralize",
						Value:       true,
						Usage:       "neutralize database",
						Destination: &restoreNeutralize,
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("move") && cCtx.Bool("neutralize") {
						return fmt.Errorf("cannot move and neutralize at the same time")
					}
					return oda.Restore(
						cCtx.Bool("any"),
						cCtx.Bool("move"),
						cCtx.Bool("neutralize"),
					)
				},
			},
			// ####################################
			// Config Commands (Requires root access)
			//   config      config commands
			{
				Name:     "config",
				Usage:    "config commands",
				Category: "Config Commands",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "initialize oda setup",
						Action: func(cCtx *cli.Context) error {
							// setup oda configuration
							return oda.ConfigInit()
						},
					},
					{
						Name:  "pyright",
						Usage: "Setup pyright settings",
						Action: func(cCtx *cli.Context) error {
							// update pyright config
							// in project directory
							return oda.ConfigPyright()
						},
					},
					{
						Name:  "vscode",
						Usage: "Setup vscode settings.json and launch.json",
						Action: func(cCtx *cli.Context) error {
							// add settings.json and launch.json to .vscode folder
							// in project directory
							return oda.ConfigVSCode()
						},
					},
				},
			},
			//   hostsfile   Update /etc/hosts file (Requires root access)
			{
				Name:     "hosts",
				Usage:    "Update /etc/hosts file (Requires root access)",
				Category: "Config Commands",
				Action: func(cCtx *cli.Context) error {
					return oda.HostsUpdate(cCtx.Args().First())
				},
			},
			// ####################################
			// Database Management
			//   db          Access postgresql
			{
				Name:     "db",
				Usage:    "Access postgresql",
				Category: "Database Management",
				Subcommands: []*cli.Command{
					{
						Name:  "fullreset",
						Usage: "database full reset",
						Action: func(cCtx *cli.Context) error {
							return oda.DBFullReset()
						},
					},
					{
						Name:  "exec",
						Usage: "database exec (root)",
						Action: func(cCtx *cli.Context) error {
							return oda.DBEXEC()
						},
					},
					{
						Name:  "psql",
						Usage: "database psql",
						Action: func(cCtx *cli.Context) error {
							return oda.DBPSQL()
						},
					},
					{
						Name:  "logs",
						Usage: "follow the database logs",
						Action: func(cCtx *cli.Context) error {
							return oda.DBLogs()
						},
					},
					{
						Name:  "start",
						Usage: "start database",
						Action: func(cCtx *cli.Context) error {
							// incus starts db instance
							return oda.DBStart()
						},
					},
					{
						Name:  "stop",
						Usage: "stop database",
						Action: func(cCtx *cli.Context) error {
							// incus stops db instance
							return oda.DBStop()
						},
					},
					{
						Name:  "restart",
						Usage: "restart database",
						Action: func(cCtx *cli.Context) error {
							// incus restarts db instance
							return oda.DBRestart()
						},
					},
				},
			},
			//   psql        Access the instance database
			{
				Name:     "psql",
				Usage:    "Access the instance database",
				Category: "Database Management",
				Action: func(cCtx *cli.Context) error {
					return oda.OdooPSQL()
				},
			},
			//   query       Query an Odoo model
			{
				Name:     "query",
				Usage:    "Query an Odoo model, make sure the flags are set before the model",
				Category: "Database Management",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "domain",
						Aliases:     []string{"d"},
						Value:       "",
						Usage:       "domain filter",
						Destination: &oda.Q.Filter,
					},
					&cli.IntFlag{
						Name:        "offset",
						Aliases:     []string{"o"},
						Value:       0,
						Usage:       "offset",
						Destination: &oda.Q.Offset,
					},
					&cli.IntFlag{
						Name:        "limit",
						Aliases:     []string{"l"},
						Value:       0,
						Usage:       "limit records returned",
						Destination: &oda.Q.Limit,
					},
					&cli.StringFlag{
						Name:        "fields",
						Aliases:     []string{"f"},
						Value:       "",
						Usage:       "fields to return",
						Destination: &oda.Q.Fields,
					},
					&cli.BoolFlag{
						Name:        "count",
						Aliases:     []string{"c"},
						Value:       false,
						Usage:       "count records",
						Destination: &oda.Q.Count,
					},
					&cli.StringFlag{
						Name:        "username",
						Aliases:     []string{"u"},
						Value:       "admin",
						Usage:       "username",
						Destination: &oda.Q.Username,
					},
					&cli.StringFlag{
						Name:        "password",
						Aliases:     []string{"p"},
						Value:       "admin",
						Usage:       "password",
						Destination: &oda.Q.Password,
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() == 0 {
						fmt.Fprintln(os.Stderr, ui.WarningStyle.Render("no model specified"))
						return nil
					}
					oda.Q.Model = cCtx.Args().First()
					// calls from outside of container to the container
					return oda.Query()
				},
			},
			// ####################################
			// Image Management
			{
				Name:     "base",
				Usage:    "Base Image Management",
				Category: "Image Management",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create Base Instance",
						Action: func(cCtx *cli.Context) error {
							// builds base image for odoo version (15,16,17,18)
							return oda.BaseCreate()
						},
					},
					{
						Name:  "update",
						Usage: "Update Base Instance",
						Action: func(cCtx *cli.Context) error {
							// update base image for odoo version (15,16,17,18)
							return oda.BaseUpdate()
						},
					},
					{
						Name:  "destroy",
						Usage: "Build Base Instance",
						Action: func(cCtx *cli.Context) error {
							// deletes base image for odoo version (15,16,17,18)
							return oda.BaseDestroy()
						},
					},
				},
			},
			// ####################################
			// Instance Management
			//   create      Create the instance
			{
				Name:     "instance",
				Usage:    "Instance Management",
				Category: "Image Management",
				Subcommands: []*cli.Command{
					{
						Name:     "create",
						Usage:    "Create the instance",
						Category: "Instance Management",
						Action: func(cCtx *cli.Context) error {
							// full instance create
							// create instance
							return oda.OdooCreate()
						},
					},
					//   destroy     Destroy the instance
					{
						Name:     "destroy",
						Usage:    "Destroy the instance",
						Category: "Instance Management",
						Action: func(cCtx *cli.Context) error {
							return oda.OdooDestroy()
						},
					},
				},
			},
			//   exec        Access the shell
			{
				Name:     "exec",
				Usage:    "Access the shell",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					username := "odoo"
					modlen := cCtx.Args().Len()
					if modlen == 1 {
						username = cCtx.Args().First()
					}
					return oda.OdooExec(username)
				},
			},
			//   restart     Restart the instance
			{
				Name:     "restart",
				Usage:    "Restart the instance",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					return oda.OdooRestart()
				},
			},
			//   start       Start the instance
			{
				Name:     "start",
				Usage:    "start the instance",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					return oda.OdooStart()
				},
			},
			//   stop        Stop the instance
			{
				Name:     "stop",
				Usage:    "Stop the instance",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					return oda.OdooStop()
				},
			},
			{
				Name:     "ps",
				Usage:    "List Odoo Instances",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					// calls "incus ps"
					return oda.OdooPS()
				},
			},
			//   logs        Follow the logs
			{
				Name:     "logs",
				Usage:    "Follow the logs",
				Category: "Container Management",
				Action: func(cCtx *cli.Context) error {
					return oda.OdooLogs()
				},
			},
			// ####################################
			// Project Commands
			//   project     Project level commands
			{
				Name:     "project",
				Usage:    "Project level commands",
				Category: "Project Commands",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "initialize project directory",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectInit()
						},
					},
					{
						Name:  "reset",
						Usage: "reset project dir and db",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectReset()
						},
					},
				},
			},
			// ####################################
			// Repo Management
			{
				Name:     "repo",
				Usage:    "Odoo community and enterprise repository management",
				Category: "Repo Management",
				Subcommands: []*cli.Command{
					{
						Name:  "base",
						Usage: "Odoo Source Repository",
						Subcommands: []*cli.Command{
							{
								Name:  "clone",
								Usage: "clone Odoo Source Repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBaseClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo Source Repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBaseUpdate()
								},
							},
						},
					},
					{
						Name:  "branch",
						Usage: "Odoo Source Branch",
						Subcommands: []*cli.Command{
							{
								Name:  "clone",
								Usage: "clone Odoo Branch Repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo Branch Repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchUpdate()
								},
							},
						},
					},
				},
			},
			// ####################################
			// welcome
			{
				Name:  "welcome",
				Usage: "Welcome message",
				Action: func(cCtx *cli.Context) error {
					return oda.Welcome()
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
