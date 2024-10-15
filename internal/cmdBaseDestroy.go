/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var baseDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy base image",
	Long:  `Destroy base image`,
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
			destroy bool
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Odoo Base Version").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Destroy Odoo Base Image?").
					Value(&destroy),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "destroy base form error %w", err)
			return
		}
		if destroy {
			SetInstanceState(version, "stop")
			// WaitForInstance(version, "stopped")
			DeleteInstance(version)
			fmt.Fprintln(os.Stderr, "destroying:", version)
		}
	},
}

func init() {
	baseCmd.AddCommand(baseDestroyCmd)
}
