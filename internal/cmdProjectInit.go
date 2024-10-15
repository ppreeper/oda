/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize project directory",
	Long:  `initialize project directory`,
	Run: func(cmd *cobra.Command, args []string) {
		projects := GetCurrentOdooProjects()
		versions := GetCurrentOdooRepos()

		versionOptions := []huh.Option[string]{}
		for _, version := range versions {
			versionOptions = append(versionOptions, huh.NewOption(version, version))
		}

		var (
			name    string
			edition string
			version string
			create  bool
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project Name").
					Value(&name).
					Validate(func(str string) error {
						// check if project already exists
						if existsIn(projects, str) {
							return fmt.Errorf("project %s already exists", str)
						}
						if str == "" {
							return fmt.Errorf("project name is required")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Title("Odoo Edition").
					Options(
						huh.NewOption("Community", "community"),
						huh.NewOption("Enterprise", "enterprise").Selected(true),
					).
					Value(&edition),

				huh.NewSelect[string]().
					Title("Odoo Branch").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Create Project?").
					Value(&create),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "project init form error %w", err)
			return
		}
		if !create {
			return
		}

		if err := projectSetup(name, edition, version); err != nil {
			fmt.Fprintln(os.Stderr, "project setup failed %w", err)
			return
		}
	},
}

func init() {
	projectCmd.AddCommand(projectInitCmd)
}

// projectSetup Project Config Setup
func projectSetup(projectName, edition, version string) error {
	projectDir := filepath.Join(viper.GetString("dirs.project"), projectName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("cannot create project directory %w", err)
	}
	for _, pdir := range []string{"addons", "conf", "data"} {
		projectSubDir := filepath.Join(projectDir, pdir)
		if err := os.MkdirAll(projectSubDir, 0o755); err != nil {
			return fmt.Errorf("cannot create project subdirectory %s %w", pdir, err)
		}
	}

	// odoo.conf
	odooConfFile := filepath.Join(projectDir, "conf", "odoo.conf")
	writeOdooConf(odooConfFile, projectName, edition)

	// .env
	envFile := filepath.Join(projectDir, ".env")
	if err := os.WriteFile(envFile, []byte("ODOO_V="+version), 0o644); err != nil {
		return fmt.Errorf("cannot create project .env file %w", err)
	}

	// .oda.yaml
	odaYamlFile := filepath.Join(projectDir, ".oda.yaml")
	if err := os.WriteFile(odaYamlFile, []byte("version: "+version), 0o644); err != nil {
		return fmt.Errorf("cannot create project .oda.yaml file %w", err)
	}

	return nil
}

// writeOdooConf Write Odoo Configfile
func writeOdooConf(file, projectName, edition string) error {
	// fmt.Println("writeOdooConf", file, projectName, edition)
	projectName = strings.ReplaceAll(projectName, "-", "_")
	dbname := projectName + "_" + viper.GetString("system.domain")
	fo, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("cannot create odoo.conf file %w", err)
	}
	defer fo.Close()

	data := map[string]string{
		"db_host":        viper.GetString("database.host"),
		"db_port":        viper.GetString("database.port"),
		"db_user":        viper.GetString("database.username"),
		"db_password":    viper.GetString("database.password"),
		"db_name":        dbname,
		"enterprise_dir": "",
	}
	if edition == "enterprise" {
		data["enterprise_dir"] = "/opt/odoo/enterprise,"
	}
	t, err := template.ParseFS(embedFS, "templates/odoo.conf")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)
	return nil
}
