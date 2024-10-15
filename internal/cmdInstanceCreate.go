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

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create the instance",
	Long:    `Create the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		version := fmt.Sprintf("%0.1f", viper.GetFloat64("version"))
		odooimage := "odoo-" + strings.ReplaceAll(version, ".", "-")
		CopyInstance(odooimage, project)
		time.Sleep(5 * time.Second)
		if err := IncusIdmap(project); err != nil {
			fmt.Fprintln(os.Stderr, "error idmap", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
