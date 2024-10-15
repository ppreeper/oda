/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var repoBaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Odoo Source Repository",
	Long:  `Odoo Source Repository`,
}

func init() {
	repoCmd.AddCommand(repoBaseCmd)
}
