/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var baseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create base image",
	Long:  `Create base image`,
	Run: func(cmd *cobra.Command, args []string) {
		repoVersionsComparison := GetCurrentOdooRepos()
		versions := GetCurrentOdooRepos()
		odooInstances := getBaseImages()

		for _, repo := range repoVersionsComparison {
			repoVersion := "odoo-" + strings.ReplaceAll(repo, ".", "-")
			for _, odooInstance := range odooInstances {
				if repoVersion == odooInstance {
					versions = removeValue(versions, repo)
				}
			}
		}

		if len(versions) == 0 {
			fmt.Fprintln(os.Stderr, "no versions to create, rebuild if necessary")
			return
		}

		versionOptions := []huh.Option[string]{}
		for _, version := range versions {
			versionOptions = append(versionOptions, huh.NewOption(version, version))
		}

		var (
			version string
			create  bool
		)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Odoo Base Version").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Create Odoo Base Image?").
					Value(&create),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "create base form error %w", err)
			return
		}
		if create {
			if err := BaseCreate(version); err != nil {
				fmt.Fprintf(os.Stderr, "create base %s error %v\n", version, err)
				return
			}
		}
	},
}

func init() {
	baseCmd.AddCommand(baseCreateCmd)
}

func GetOdooConfig(version string) OdooConfig {
	for _, config := range OdooConfigs {
		if config.Version == version {
			return config
		}
	}
	return OdooConfig{}
}

func BaseCreate(version string) error {
	fmt.Println("Creating base image for Odoo version", version)
	config := GetOdooConfig(version)

	CreateInstance(config.InstanceName, config.Image)

	WaitForInstance(config.InstanceName, "RUNNING")

	SetInstanceState(config.InstanceName, "start")

	roleUpdateScript(config.InstanceName)

	// roleGupScript(config.InstanceName)
	rolePreeperRepo(config.InstanceName)

	roleUpdate(config.InstanceName)

	fmt.Fprintln(os.Stderr, "add common system packages to", config.InstanceName)
	if err := aptInstall(config.InstanceName, config.BaselinePackages...); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	roleOdooUser(config.InstanceName)

	roleOdooDirs(config.InstanceName)

	rolePostgresqlRepo(config.InstanceName)

	rolePostgresqlClient(config.InstanceName, OdooDatabase.Version)

	roleWkhtmltopdf(config.InstanceName)

	fmt.Fprintln(os.Stderr, "add odoo dependencies to", config.InstanceName)
	if err := aptInstall(config.InstanceName, config.Odoobase...); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	npmInstall(config.InstanceName, "rtlcss")

	roleGeoIP2DB(config.InstanceName)

	rolePaperSize(config.InstanceName)

	roleOdooService(config.InstanceName)

	roleCaddy(config.InstanceName)

	roleCaddyService(config.InstanceName)

	SetInstanceState(config.InstanceName, "stop")

	return nil
}
