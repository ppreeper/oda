/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:     "ps",
	Short:   "List Odoo Instances",
	Long:    `List Odoo Instances`,
	GroupID: "image",
	Run: func(cmd *cobra.Command, args []string) {
		projects := GetCurrentOdooProjects()
		instances, err := GetInstances()
		if err != nil {
			fmt.Fprintln(os.Stderr, "instances list failed %w", err)
			return
		}

		instanceList := []Instance{}

		for _, instance := range instances {
			for _, project := range projects {
				if instance.Name == project {
					instanceList = append(instanceList, instance)
				}
			}
		}

		maxnameLen := 0
		maxstateLen := 0
		maxipv4Len := 15

		for _, instance := range instanceList {
			if len(instance.Name) > maxnameLen {
				maxnameLen = len(instance.Name)
			}
			if len(instance.State) > maxstateLen {
				maxstateLen = len(instance.State)
			}
		}

		fmt.Fprintf(os.Stderr, "%-*s %-*s %-*s\n",
			maxnameLen+2, "NAME",
			maxstateLen+2, "STATE",
			maxipv4Len+2, "IPV4",
		)
		for _, instance := range instanceList {
			fmt.Fprintf(os.Stderr, "%-*s %-*s %-*s\n",
				maxnameLen+2, instance.Name,
				maxstateLen+2, instance.State,
				maxipv4Len+2, instance.IP4,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
