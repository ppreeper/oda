/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"bufio"
	"fmt"
	"os"
	"slices"

	"github.com/ppreeper/str"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var hostsfileCmd = &cobra.Command{
	Use:     "hostsfile",
	Short:   "Update /etc/hosts file (Requires root access)",
	Long:    `Update /etc/hosts file (Requires root access)`,
	GroupID: "config",
	Run: func(cmd *cobra.Command, args []string) {
		sudouser, _ := os.LookupEnv("SUDO_USER")
		if sudouser == "" {
			fmt.Fprintln(os.Stderr, "not allowed: this requires root access")
			return
		}

		hosts, err := os.Open("/etc/hosts")
		if err != nil {
			fmt.Fprintln(os.Stderr, "hosts file read failed %w", err)
			return
		}
		defer hosts.Close()

		hostlines := []string{}
		scanner := bufio.NewScanner(hosts)
		for scanner.Scan() {
			hostlines = append(hostlines, scanner.Text())
		}
		begin := slices.Index(hostlines, "#ODABEGIN")
		end := slices.Index(hostlines, "#ODAEND")

		if begin > end {
			fmt.Fprintln(os.Stderr, "host file out of order, edit /etc/hosts manually")
			return
		}

		projects := GetCurrentOdooProjects()

		instances, err := GetInstances()
		if err != nil {
			fmt.Fprintln(os.Stderr, "instances list failed %w", err)
			return
		}

		projectLines := []string{}

		for _, instance := range instances {
			for _, project := range projects {
				if instance.Name == project {
					projectLines = append(projectLines,
						str.RightLen(instance.IP4, " ", 16)+" "+instance.Name+"."+viper.GetString("system.domain"))
				}
			}
		}
		instance, err := GetInstance(viper.GetString("database.host"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "instance %s not found %v\n", viper.GetString("database.host"), err)
			return
		}
		projectLines = append(projectLines,
			str.RightLen(instance.IP4, " ", 16)+" "+viper.GetString("database.host")+"."+viper.GetString("system.domain"))

		newHostlines := []string{}
		if begin == -1 && end == -1 {
			newHostlines = append(newHostlines, hostlines...)
			newHostlines = append(newHostlines, "#ODABEGIN")
			newHostlines = append(newHostlines, projectLines...)
			newHostlines = append(newHostlines, "#ODAEND")
		} else {
			newHostlines = append(newHostlines, hostlines[:begin+1]...)
			newHostlines = append(newHostlines, projectLines...)
			newHostlines = append(newHostlines, hostlines[end:]...)
		}

		fo, err := os.Create("/etc/hosts")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error write /etc/hosts file failed %w", err)
			return
		}
		defer fo.Close()
		for _, hostline := range newHostlines {
			fo.WriteString(hostline + "\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(hostsfileCmd)
}
