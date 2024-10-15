/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var baseUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update base image",
	Long:  `Update base image`,
	Run: func(cmd *cobra.Command, args []string) {
		odooInstances := getBaseImages()

		if len(odooInstances) == 0 {
			fmt.Fprintln(os.Stderr, "no base images found")
			return
		}

		versionOptions := []huh.Option[string]{}
		for _, version := range odooInstances {
			versionOptions = append(versionOptions, huh.NewOption(version, version))
		}

		var (
			version string
			update  bool
		)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Odoo Base Version").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Update Odoo base image packages?").
					Value(&update),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "updating base form error %w", err)
			return
		}
		if !update {
			return
		}

		checkState := GetInstanceState(version)

		switch strings.ToUpper(checkState.Metadata.Status) {
		case "STOPPED":
			SetInstanceState(version, "start")
			for {
				currentState := GetInstanceState(version)
				if strings.ToUpper(currentState.Metadata.Status) == "RUNNING" {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			roleUpdate(version)
			SetInstanceState(version, "stop")
			for {
				currentState := GetInstanceState(version)
				if strings.ToUpper(currentState.Metadata.Status) == "STOPPED" {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
		case "RUNNING":
			roleUpdate(version)
			SetInstanceState(version, "stop")
			for {
				currentState := GetInstanceState(version)
				if strings.ToUpper(currentState.Metadata.Status) == "RUNNING" {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	},
}

func init() {
	baseCmd.AddCommand(baseUpdateCmd)
}
