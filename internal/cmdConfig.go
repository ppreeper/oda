/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "config commands",
	Long:    `config commands, including environment initialization and development setup`,
	GroupID: "config",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
