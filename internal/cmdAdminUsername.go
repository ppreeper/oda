/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var adminUsernameCmd = &cobra.Command{
	Use:   "username",
	Short: "Odoo Admin username",
	Long:  `Odoo Admin username`,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}

		var user1, user2 string
		huh.NewInput().
			Title("Please enter  the new admin username:").
			Prompt(">").
			Value(&user1).
			Run()
		huh.NewInput().
			Title("Please verify the new admin username:").
			Prompt(">").
			Value(&user2).
			Run()

		if user1 != user2 {
			fmt.Fprintln(os.Stderr, "usernames entered do not match")
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
				return err
			}
			return nil
		}()

		// Write username to database
		_, err = db.Exec("update res_users set login=$1 where id=2;",
			strings.TrimSpace(string(user1)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error updating username %w", err)
			return
		}

		fmt.Fprintln(os.Stderr, "Admin username changed to", user1)
	},
}

func init() {
	adminCmd.AddCommand(adminUsernameCmd)
}
