/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/passhash"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var adminPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Odoo Admin password",
	Long:  `Odoo Admin password`,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}

		var password1, password2 string
		huh.NewInput().
			Title("Please enter  the admin password:").
			Prompt(">").
			EchoMode(huh.EchoModePassword).
			Value(&password1).
			Run()
		huh.NewInput().
			Title("Please verify the admin password:").
			Prompt(">").
			EchoMode(huh.EchoModePassword).
			Value(&password2).
			Run()
		if password1 != password2 {
			fmt.Fprintln(os.Stderr, "passwords entered do not match")
			return
		}
		var confirm bool
		huh.NewConfirm().
			Title("Are you sure you want to change the admin password?").
			Affirmative("yes").
			Negative("no").
			Value(&confirm).
			Run()
		if !confirm {
			fmt.Fprintln(os.Stderr, "password change cancelled")
			return
		}

		// Open Database
		instance, err := GetInstance(viper.GetString("database.host"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error getting instance %w", err)
			return
		}

		cwd, _ := GetProject()

		dbport := viper.GetInt("database.port")
		dbhost := instance.IP4
		dbuser := viper.GetString("database.username")
		dbpassword := viper.GetString("database.password")
		dbname := GetOdooConf(cwd, "db_name")

		if instance.State != "Running" {
			fmt.Fprintln(os.Stderr, "database instance is not running")
			return
		}

		db, err := OpenDatabase(Database{
			Hostname: dbhost,
			Port:     dbport,
			Username: dbuser,
			Password: dbpassword,
			Database: dbname,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening database %w", err)
			return
		}
		defer func() error {
			if err := db.Close(); err != nil {
				fmt.Fprintln(os.Stderr, "error closing database %w", err)
				return nil
			}
			return nil
		}()

		// Write password to database
		passkey, err := passhash.MakePassword(password1, 0, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "password hashing error", err)
		}
		_, err = db.Exec("update res_users set password=$1 where id=2;",
			strings.TrimSpace(string(passkey)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error updating password %w", err)
			return
		}

		fmt.Fprintln(os.Stderr, "admin password changed")
	},
}

func init() {
	adminCmd.AddCommand(adminPasswordCmd)
}
