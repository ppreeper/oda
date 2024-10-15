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

var dbStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "database stop",
	Long:  `database start`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")
		_, err := GetInstance(dbHost)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get instance %w", err)
			return
		}
		SetInstanceState(dbHost, "stop")
		// WaitForInstance(dbHost, "STOPPED")
		fmt.Fprintln(os.Stderr, dbHost, "stopped")
	},
}

func init() {
	dbCmd.AddCommand(dbStopCmd)
}
