/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dbRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "database restart",
	Long:  `database restart`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")

		SetInstanceState(dbHost, "stop")
		// WaitForInstance(dbHost, "STOPPED")
		fmt.Fprintln(os.Stderr, dbHost, "stopped")

		SetInstanceState(dbHost, "start")
		// WaitForInstance(dbHost, "RUNNING")
		fmt.Fprintln(os.Stderr, dbHost, "started")
	},
}

func init() {
	dbCmd.AddCommand(dbRestartCmd)
}
