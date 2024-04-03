package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	oda "github.com/ppreeper/oda/internal"
	"github.com/urfave/cli/v2"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		Revision := ""
		LastCommit := time.Time{}
		// DirtyBuild := false
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				Revision = string([]rune(kv.Value)[:7])
			case "vcs.time":
				LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
			}
		}
		return fmt.Sprintf("%d%02d%02d-%s", LastCommit.Year(), LastCommit.Month(), LastCommit.Day(), Revision)
	}
	return ""
}

func main() {
	q := oda.QueryDef{}
	app := &cli.App{
		Name:                 "oda",
		Usage:                "Odoo Administration Tool",
		Version:              Commit(),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:     "admin",
				Usage:    "Admin user management",
				Category: "admin",
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
				Name:     "app",
				Usage:    "app management",
				Category: "app",
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
				Name:     "base",
				Usage:    "Base Image Management",
				Category: "base",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create base image",
						Action: func(cCtx *cli.Context) error {
							return oda.BaseCreatePrompt()
						},
					},
					{
						Name:  "destroy",
						Usage: "Destroy base image",
						Action: func(cCtx *cli.Context) error {
							return oda.BaseDestroyPrompt()
						},
					},
					{
						Name:  "rebuild",
						Usage: "Rebuild base image",
						Action: func(cCtx *cli.Context) error {
							return oda.BaseRebuildPrompt()
						},
					},
					{
						Name:  "update",
						Usage: "Update base image packages",
						Action: func(cCtx *cli.Context) error {
							return oda.BaseUpdatePrompt()
						},
					},
				},
			},
			{
				Name:     "config",
				Usage:    "additional config options",
				Category: "config",
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
				Name:     "db",
				Usage:    "Access postgresql",
				Category: "db",
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
							// incus start db
							return oda.PgdbFullReset()
						},
					},
					{
						Name:  "logs",
						Usage: "Follow the logs",
						Action: func(cCtx *cli.Context) error {
							return oda.DBLogs()
						},
					},
				},
			},
			{
				Name:     "project",
				Usage:    "Project level commands [CAUTION]",
				Category: "project",
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
				Category: "repo",
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
			{
				Name:  "create",
				Usage: "Create the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceCreate()
				},
			},
			{
				Name:  "destroy",
				Usage: "Destroy the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceDestroy()
				},
			},
			{
				Name:  "rebuild",
				Usage: "Rebuild the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceRebuild()
				},
			},
			{
				Name:  "start",
				Usage: "Start the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceStart()
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceStop()
				},
			},
			{
				Name:  "restart",
				Usage: "Restart the instance",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceRestart()
				},
			},
			{
				Name:  "ps",
				Usage: "List Odoo Instances",
				Action: func(cCtx *cli.Context) error {
					return oda.InstancePS()
				},
			},
			{
				Name:  "logs",
				Usage: "Follow the logs",
				Action: func(cCtx *cli.Context) error {
					return oda.InstanceLogs()
				},
			},
			{
				Name:  "exec",
				Usage: "Access the shell",
				Action: func(cCtx *cli.Context) error {
					username := "odoo"
					modlen := cCtx.Args().Len()
					if modlen != 0 {
						username = cCtx.Args().First()
					}
					return oda.InstanceExec(username)
				},
			},
			// //////////////////////////////////////////////
			{
				Name:  "psql",
				Usage: "Access the instance database",
				Action: func(cCtx *cli.Context) error {
					return oda.InstancePSQL()
				},
			},
			{
				Name:  "scaffold",
				Usage: "Generates an Odoo module skeleton in addons",
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
				Name:      "query",
				Usage:     "Query an Odoo model",
				UsageText: "oda query <model> [command options]",
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
					&cli.StringFlag{
						Name:        "username",
						Aliases:     []string{"u"},
						Value:       "admin",
						Usage:       "username",
						Destination: &q.Username,
					},
					&cli.StringFlag{
						Name:        "password",
						Aliases:     []string{"p"},
						Value:       "admin",
						Usage:       "password",
						Destination: &q.Password,
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
				Name:  "backup",
				Usage: "Backup database filestore and addons",
				Action: func(cCtx *cli.Context) error {
					return oda.AdminBackup()
				},
			},
			{
				Name:  "restore",
				Usage: "Restore database and filestore or addons",
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
				Name:  "init",
				Usage: "initialize oda setup",
				Action: func(cCtx *cli.Context) error {
					return oda.AdminInit()
				},
			},
			{
				Name:  "hostsfile",
				Usage: "Update /etc/hosts file (Requires root access)",
				Action: func(cCtx *cli.Context) error {
					return oda.ProjectHostsFile()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
