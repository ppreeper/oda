package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ppreeper/oda"
	"github.com/urfave/cli/v2"
)

func main() {
	q := oda.QueryDef{}
	app := &cli.App{
		Name:                 "oda",
		Usage:                "Odoo Administration Tool",
		Version:              "0.4.6",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:     "config",
				Usage:    "additional config options",
				Category: "utility",
				Subcommands: []*cli.Command{
					{
						Name:  "vscode",
						Usage: "Setup vscode settings and launch json files",
						Action: func(cCtx *cli.Context) error {
							return oda.ConfigVSCode()
						},
					},
					{
						Name:  "pyright",
						Usage: "Setup pyright settings",
						Action: func(cCtx *cli.Context) error {
							return oda.ConfigPyright()
						},
					},
				},
			},
			{
				Name:     "start",
				Usage:    "Start the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceStart()
				},
			},
			{
				Name:     "stop",
				Usage:    "Stop the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceStop()
				},
			},
			{
				Name:     "restart",
				Usage:    "Restart the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceRestart()
				},
			},
			{
				Name:     "app",
				Usage:    "app management",
				Category: "instance",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "Install module(s)",
						Action: func(cCtx *cli.Context) error {
							modlen := cCtx.Args().Len()
							if modlen == 0 {
								return fmt.Errorf("no modules specified")
							}
							return oda.InstanceAppInstallUpgrade(true, cCtx.Args().Slice()...)
						},
					},
					{
						Name:  "upgrade",
						Usage: "Upgrade module(s)",
						Action: func(cCtx *cli.Context) error {
							modlen := cCtx.Args().Len()
							if modlen == 0 {
								return fmt.Errorf("no modules specified")
							}
							return oda.InstanceAppInstallUpgrade(false, cCtx.Args().Slice()...)
						},
					},
				},
			},
			{
				Name:     "logs",
				Usage:    "Follow the logs",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceLogs()
				},
			},
			{
				Name:     "scaffold",
				Usage:    "Generates an Odoo module skeleton in addons",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no module specified")
					}
					module := cCtx.Args().First()
					return oda.InstanceScaffold(module)
				},
			},
			{
				Name:     "ps",
				Usage:    "List Odoo Instances",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstancePS()
				},
			},
			{
				Name:     "exec",
				Usage:    "Access the shell",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceExec()
				},
			},
			{
				Name:     "psql",
				Usage:    "Access the instance database",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstancePSQL()
				},
			},
			{
				Name:      "query",
				Usage:     "Query an Odoo model",
				UsageText: "oda query <model> [command options]",
				Category:  "instance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "domain",
						Aliases:     []string{"d"},
						Value:       "",
						Usage:       "domain filter",
						Destination: &q.Filter,
					},
					&cli.IntFlag{
						Name:        "offset",
						Aliases:     []string{"o"},
						Value:       0,
						Usage:       "offset",
						Destination: &q.Offset,
					},
					&cli.IntFlag{
						Name:        "limit",
						Aliases:     []string{"l"},
						Value:       0,
						Usage:       "limit records returned",
						Destination: &q.Limit,
					},
					&cli.StringFlag{
						Name:        "fields",
						Aliases:     []string{"f"},
						Value:       "",
						Usage:       "fields to return",
						Destination: &q.Fields,
					},
					&cli.BoolFlag{
						Name:        "count",
						Aliases:     []string{"c"},
						Value:       false,
						Usage:       "count records",
						Destination: &q.Count,
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() == 0 {
						return fmt.Errorf("no model specified")
					}
					q.Model = cCtx.Args().First()
					return oda.InstanceQuery(&q)
				},
			},
			{
				Name:     "db",
				Usage:    "Access postgresql",
				Category: "database",
				Subcommands: []*cli.Command{
					{
						Name:  "psql",
						Usage: "database psql",
						Action: func(cCtx *cli.Context) error {
							return oda.PgdbPgsql()
						},
					},
					{
						Name:  "start",
						Usage: "database start",
						Action: func(cCtx *cli.Context) error {
							return oda.PgdbStart()
						},
					},
					{
						Name:  "stop",
						Usage: "database stop",
						Action: func(cCtx *cli.Context) error {
							return oda.PgdbStop()
						},
					},
					{
						Name:  "restart",
						Usage: "database restart",
						Action: func(cCtx *cli.Context) error {
							return oda.PgdbRestart()
						},
					},
					{
						Name:  "fullreset",
						Usage: "database fullreset",
						Action: func(cCtx *cli.Context) error {
							return oda.PgdbFullReset()
						},
					},
				},
			},
			{
				Name:     "proxy",
				Usage:    "Caddy proxy",
				Category: "proxy",
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "proxy start",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyStart()
						},
					},
					{
						Name:  "stop",
						Usage: "proxy stop",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyStop()
						},
					},
					{
						Name:  "restart",
						Usage: "proxy restart",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyRestart()
						},
					},
					{
						Name:  "generate",
						Usage: "proxy generate",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyGenerate()
						},
					},
				},
			},
			{
				Name:     "backup",
				Usage:    "Backup database filestore and addons",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					return oda.AdminBackup()
				},
			},
			{
				Name:     "restore",
				Usage:    "Restore database and filestore or addons",
				Category: "admin",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "move",
						Value: false,
						Usage: "move server",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if !oda.IsProject() {
						return fmt.Errorf("not in a project directory")
					}
					move := cCtx.Bool("move")
					return oda.AdminRestore(move)
				},
			},
			{
				Name:     "admin",
				Usage:    "Admin user management",
				Category: "instance",
				Subcommands: []*cli.Command{
					{
						Name:  "username",
						Usage: "Odoo Admin username",
						Action: func(cCtx *cli.Context) error {
							return oda.AdminUsername()
						},
					},
					{
						Name:  "password",
						Usage: "Odoo Admin password",
						Action: func(cCtx *cli.Context) error {
							return oda.AdminPassword()
						},
					},
				},
			},
			{
				Name:     "init",
				Usage:    "initialize oda setup",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					return oda.AdminInit()
				},
			},
			{
				Name:     "hostsfile",
				Usage:    "Update /etc/hosts file (Requires root access)",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					return oda.ProjectHostsFile()
				},
			},
			{
				Name:     "project",
				Usage:    "Project level commands [CAUTION]",
				Category: "admin",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "initialize project directory",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectIinit()
						},
					},
					{
						Name:  "branch",
						Usage: "initialize branch of project",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectBranch()
						},
					},
					{
						Name:  "rebuild",
						Usage: "rebuild from another project",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectRebuild()
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
			{
				Name:     "repo",
				Usage:    "Odoo community and enterprise repository management",
				Category: "admin",
				Subcommands: []*cli.Command{
					{
						Name:  "base",
						Usage: "Odoo Source Repository",
						Subcommands: []*cli.Command{
							{
								Name:  "clone",
								Usage: "clone Odoo source repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBaseClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo source repository",
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
								Usage: "clone Odoo branch repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo branch repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchUpdate()
								},
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
