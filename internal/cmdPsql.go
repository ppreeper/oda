/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var psqlCmd = &cobra.Command{
	Use:     "psql",
	Short:   "Access the instance database",
	Long:    `Access the instance database`,
	GroupID: "database",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		dbHost := viper.GetString("database.host")
		cwd, _ := GetProject()

		dbuser := GetOdooConf(cwd, "db_user")
		dbpassword := GetOdooConf(cwd, "db_password")
		dbname := GetOdooConf(cwd, "db_name")

		uid, err := IncusGetUid(dbHost, "postgres")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get postgres uid %w", err)
			return
		}

		incusCmd := exec.Command("incus", "exec", dbHost, "--user", uid,
			"--env", "PGPASSWORD="+dbpassword, "-t", "--",
			"psql", "-h", "127.0.0.1", "-U", dbuser, dbname,
		)
		incusCmd.Stdin = os.Stdin
		incusCmd.Stdout = os.Stdout
		incusCmd.Stderr = os.Stderr
		if err := incusCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error instance psql %w", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(psqlCmd)
}
