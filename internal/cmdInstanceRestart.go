/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart the instance",
	Long:    `Restart the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()

		instance, err := GetInstance(project)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo instance", err)
			return
		}
		if instance == (Instance{}) {
			fmt.Fprintln(os.Stderr, "no odoo instance, please launch one first")
			return
		}
		switch strings.ToUpper(instance.State) {
		case "RUNNING":
			SetInstanceState(project, "restart")
		case "STOPPED":
			SetInstanceState(project, "start")
		}

		instanceStatus := GetInstanceState(project)
		fmt.Fprintln(os.Stderr, project, instanceStatus.Metadata.Status)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
