/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:     "db",
	Short:   "Access postgresql",
	Long:    `Access postgresql`,
	GroupID: "database",
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
