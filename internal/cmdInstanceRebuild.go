/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rebuildCmd = &cobra.Command{
	Use:     "rebuild",
	Short:   "Rebuild the instance",
	Long:    `Rebuild the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		confim := AreYouSure("rebuild the " + project + " instance")
		if !confim {
			fmt.Fprintln(os.Stderr, "rebuild the "+project+" instance canceled")
			return
		}

		instance, err := GetInstance(project)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo instance", err)
			return
		}
		switch strings.ToUpper(instance.State) {
		case "RUNNING":
			SetInstanceState(project, "stop")
		}

		// Delete the instance
		DeleteInstance(project)
		fmt.Fprintln(os.Stderr, "Image "+project+" deleted")

		time.Sleep(5 * time.Second)

		// Create the instance
		version := fmt.Sprintf("%0.1f", viper.GetFloat64("version"))
		odooimage := "odoo-" + strings.ReplaceAll(version, ".", "-")
		CopyInstance(odooimage, project)
		fmt.Fprintln(os.Stderr, "Image "+project+" restarted")
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
