/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Project level commands",
	Long:    `Project level commands`,
	GroupID: "project",
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
