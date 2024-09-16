package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	server "github.com/ppreeper/oda/internal/server"
	"github.com/urfave/cli/v2"
)

// Odoo Server Administration Tool
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
	// q := oda.QueryDef{}
	app := &cli.App{
		Name:                 "oda",
		Usage:                "Odoo Server Administration Tool",
		Version:              Commit(),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
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
