package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "oda",
		Usage:                "Odoo Administration Tool",
		Version:              "0.4.4",
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
							return configVSCode()
						},
					},
					{
						Name:  "pyright",
						Usage: "Setup pyright settings",
						Action: func(cCtx *cli.Context) error {
							return configPyright()
						},
					},
				},
			},
			{
				Name:     "start",
				Usage:    "Start the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instanceStart()
				},
			},
			{
				Name:     "stop",
				Usage:    "Stop the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instanceStop()
				},
			},
			{
				Name:     "restart",
				Usage:    "Restart the instance",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instanceRestart()
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
							return instanceAppInstallUpgrade(true, cCtx.Args().Slice()...)
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
							return instanceAppInstallUpgrade(false, cCtx.Args().Slice()...)
						},
					},
				},
			},
			{
				Name:     "logs",
				Usage:    "Follow the logs",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instanceLogs()
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
					return instanceScaffold(module)
				},
			},
			{
				Name:     "ps",
				Usage:    "List Odoo Instances",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instancePS()
				},
			},
			{
				Name:     "exec",
				Usage:    "Access the shell",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instanceExec()
				},
			},
			{
				Name:     "psql",
				Usage:    "Access the instance database",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					return instancePSQL()
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
							return pgdbPgsql()
						},
					},
					{
						Name:  "start",
						Usage: "database start",
						Action: func(cCtx *cli.Context) error {
							return pgdbStart()
						},
					},
					{
						Name:  "stop",
						Usage: "database stop",
						Action: func(cCtx *cli.Context) error {
							return pgdbStop()
						},
					},
					{
						Name:  "restart",
						Usage: "database restart",
						Action: func(cCtx *cli.Context) error {
							return pgdbRestart()
						},
					},
					{
						Name:  "fullreset",
						Usage: "database fullreset",
						Action: func(ctx *cli.Context) error {
							return pgdbFullReset()
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
							return proxyStart()
						},
					},
					{
						Name:  "stop",
						Usage: "proxy stop",
						Action: func(cCtx *cli.Context) error {
							return proxyStop()
						},
					},
					{
						Name:  "restart",
						Usage: "proxy restart",
						Action: func(cCtx *cli.Context) error {
							return proxyRestart()
						},
					},
					{
						Name:  "generate",
						Usage: "proxy generate",
						Action: func(cCtx *cli.Context) error {
							return proxyGenerate()
						},
					},
				},
			},
			{
				Name:     "backup",
				Usage:    "Backup database filestore and addons",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					return adminBackup()
				},
			},
			{
				Name:     "restore",
				Usage:    "Restore database and filestore or addons",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					return adminRestore()
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
							return adminUsername()
						},
					},
					{
						Name:  "password",
						Usage: "Odoo Admin password",
						Action: func(cCtx *cli.Context) error {
							return adminPassword()
						},
					},
				},
			},
			{
				Name:     "init",
				Usage:    "initialize oda setup",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					return adminInit()
				},
			},
			{
				Name:     "hostsfile",
				Usage:    "Update /etc/hosts file (Requires root access)",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					return projectHostsFile()
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
							return projectIinit()
						},
					},
					{
						Name:  "branch",
						Usage: "initialize branch of project",
						Action: func(cCtx *cli.Context) error {
							return projectBranch()
						},
					},
					{
						Name:  "rebuild",
						Usage: "rebuild from another project",
						Action: func(cCtx *cli.Context) error {
							return projectRebuild()
						},
					},
					{
						Name:  "reset",
						Usage: "reset project dir and db",
						Action: func(cCtx *cli.Context) error {
							return projectReset()
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
									return repoBaseClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo source repository",
								Action: func(cCtx *cli.Context) error {
									return repoBaseUpdate()
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
									return repoBranchClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo branch repository",
								Action: func(cCtx *cli.Context) error {
									return repoBranchUpdate()
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
