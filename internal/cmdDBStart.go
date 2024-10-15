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

var dbStartCmd = &cobra.Command{
	Use:   "start",
	Short: "database start",
	Long:  `database start`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")
		_, err := GetInstance(dbHost)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get instance %w", err)
		}
		SetInstanceState(dbHost, "start")
		// WaitForInstance(dbHost, "RUNNING")
		fmt.Fprintln(os.Stderr, dbHost, "started")
	},
}

func init() {
	dbCmd.AddCommand(dbStartCmd)
}
