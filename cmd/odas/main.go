package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	server "github.com/ppreeper/oda/internal/server"
	"github.com/urfave/cli/v2"
)

//go:generate sh -c "printf '%s (%s)' $(git tag -l --contains HEAD) $(date +%Y%m%d)-$(git rev-parse --short HEAD)" > commit.txt
//go:embed commit.txt
var Commit string

func main() {
	// q := oda.QueryDef{}
	app := &cli.App{
		Name:                 "odas",
		Usage:                "Odoo Server Administration Tool",
		Version:              Commit,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:     "caddy",
				Usage:    "update caddyfile",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					domain := ""
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						domain = "local"
					} else {
						domain = cCtx.Args().First()
					}
					return server.CaddyfileUpdate(domain)
				},
			},
			{
				Name:     "hosts",
				Usage:    "update hosts file",
				Category: "admin",
				Action: func(cCtx *cli.Context) error {
					domain := ""
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						domain = "local"
					} else {
						domain = cCtx.Args().First()
					}
					return server.HostsUpdate(domain)
				},
			},
			// App Management
			{
				Name:     "install",
				Usage:    "Install module(s)",
				Category: "app management",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no modules specified")
					}
					return server.InstanceAppInstallUpgrade(true, cCtx.Args().Slice()...)
				},
			},
			{
				Name:     "upgrade",
				Usage:    "Upgrade module(s)",
				Category: "app management",
				Action: func(cCtx *cli.Context) error {
					modlen := cCtx.Args().Len()
					if modlen == 0 {
						return fmt.Errorf("no modules specified")
					}
					return server.InstanceAppInstallUpgrade(false, cCtx.Args().Slice()...)
				},
			},
			// Backup Restore
			{
				Name:     "backup",
				Usage:    "Backup database filestore and addons",
				Category: "backup",
				Action: func(cCtx *cli.Context) error {
					return server.AdminBackup()
				},
			},
			{
				Name:     "restore",
				Usage:    "Restore database and filestore or addons",
				Category: "backup",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "any",
						Value: false,
						Usage: "any backup",
					},
					&cli.BoolFlag{
						Name:  "move",
						Value: false,
						Usage: "move server",
					},
					&cli.BoolFlag{
						Name:  "neutralize",
						Value: false,
						Usage: "fully neutralize the server",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("move") && cCtx.Bool("neutralize") {
						return fmt.Errorf("cannot move and neutralize at the same time")
					}
					return server.AdminRestore(
						cCtx.Bool("any"),
						cCtx.Bool("move"),
						cCtx.Bool("neutralize"),
					)
				},
			},
			{
				Name:     "trim",
				Usage:    "Trim database backups",
				Category: "backup",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "limit",
						Value: 10,
						Usage: "number of backups to keep",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return server.Trim(cCtx.Int("limit"), false)
				},
			},
			{
				Name:     "trimall",
				Usage:    "Trim all database backups",
				Category: "backup",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "limit",
						Value: 10,
						Usage: "number of backups to keep",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return server.Trim(cCtx.Int("limit"), true)
				},
			},
			// Control
			{
				Name:     "start",
				Usage:    "Start the instance",
				Category: "control",
				Action: func(cCtx *cli.Context) error {
					return server.ServiceStart()
				},
			},
			{
				Name:     "stop",
				Usage:    "Stop the instance",
				Category: "control",
				Action: func(cCtx *cli.Context) error {
					return server.ServiceStop()
				},
			},
			{
				Name:     "restart",
				Usage:    "Restart the instance",
				Category: "control",
				Action: func(cCtx *cli.Context) error {
					return server.ServiceRestart()
				},
			},
			// General
			{
				Name:     "logs",
				Usage:    "Follow the logs",
				Category: "general",
				Action: func(cCtx *cli.Context) error {
					return server.InstanceLogs()
				},
			},
			// User Management
			{
				Name:     "admin",
				Usage:    "Admin user management",
				Category: "user management",
				Subcommands: []*cli.Command{
					{
						Name:  "username",
						Usage: "Odoo Admin username",
						Action: func(cCtx *cli.Context) error {
							return server.AdminUsername()
						},
					},
					{
						Name:  "password",
						Usage: "Odoo Admin password",
						Action: func(cCtx *cli.Context) error {
							return server.AdminPassword()
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
