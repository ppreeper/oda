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

var dbPsqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "database psql",
	Long:  `database psql`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")

		uid, err := IncusGetUid(dbHost, "postgres")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get postgres uid %w", err)
			return
		}
		pgCmd := exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--", "psql")
		pgCmd.Stdin = os.Stdin
		pgCmd.Stdout = os.Stdout
		pgCmd.Stderr = os.Stderr
		if err := pgCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to run psql %w", err)
			return
		}
	},
}

func init() {
	dbCmd.AddCommand(dbPsqlCmd)
}
