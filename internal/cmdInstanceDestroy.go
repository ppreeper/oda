/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Short:   "Destroy the instance",
	Long:    `Destroy the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		confim := AreYouSure("destroy the " + project + " instance")
		if !confim {
			fmt.Fprintln(os.Stderr, "destroying the "+project+" instance canceled")
			return
		}
		SetInstanceState(project, "stop")
		DeleteInstance(project)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
