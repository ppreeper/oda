/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:     "admin",
	Short:   "Admin user management",
	Long:    `Admin user management`,
	GroupID: "user",
}

func init() {
	rootCmd.AddCommand(adminCmd)
}
