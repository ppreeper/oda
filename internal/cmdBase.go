/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"strings"

	"github.com/spf13/cobra"
)

var baseCmd = &cobra.Command{
	Use:     "base",
	Short:   "Base Image Management",
	Long:    `Base Image Management`,
	GroupID: "image",
}

func init() {
	rootCmd.AddCommand(baseCmd)
}

func getBaseImages() []string {
	versions := GetCurrentOdooRepos()
	var odooVersions []string
	for _, version := range versions {
		odooVersions = append(odooVersions, "odoo-"+strings.ReplaceAll(version, ".", "-"))
	}

	instances, err := GetInstances()
	cobra.CheckErr(err)
	var odooInstances []string
	for _, inst := range instances {
		for _, version := range odooVersions {
			if inst.Name == version {
				odooInstances = append(odooInstances, inst.Name)
			}
		}
	}
	return odooInstances
}
