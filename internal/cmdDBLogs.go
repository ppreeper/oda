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

var dbLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Follow the database logs",
	Long:  `Follow the database logs`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")

		cmdArgs := []string{"exec", dbHost, "-t", "--", "journalctl", "-f"}

		c := exec.Command("incus", cmdArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error running journalctl: %w", err)
		}
	},
}

func init() {
	dbCmd.AddCommand(dbLogsCmd)
}
