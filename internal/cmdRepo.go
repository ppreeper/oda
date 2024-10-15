/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Odoo community and enterprise repository management",
	Long:    `Odoo community and enterprise repository management`,
	GroupID: "repo",
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
