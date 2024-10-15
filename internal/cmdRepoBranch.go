/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"github.com/spf13/cobra"
)

var repoBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Odoo Source Branch",
	Long:  `Odoo Source Branch`,
}

func init() {
	repoCmd.AddCommand(repoBranchCmd)
}
