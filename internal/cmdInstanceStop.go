/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the instance",
	Long:    `Stop the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		SetInstanceState(project, "stop")
		instanceStatus := GetInstanceState(project)
		fmt.Fprintln(os.Stderr, project, instanceStatus.Metadata.Status)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
