/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize oda setup",
	Long:  `initialize oda setup`,
	Run: func(cmd *cobra.Command, args []string) {
		HOME, _ := os.UserHomeDir()

		CFGDIR, err := os.UserConfigDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		// $HOME/.config/oda/oda.yaml
		odaConf := genOdaConf()

		odaConf.Dirs.Repo = filepath.Join(HOME, "workspace/repos/odoo")
		if _, err := os.Stat(odaConf.Dirs.Repo); os.IsNotExist(err) {
			os.MkdirAll(odaConf.Dirs.Repo, 0o755)
		}

		odaConf.Dirs.Project = filepath.Join(HOME, "workspace/odoo")
		if _, err := os.Stat(odaConf.Dirs.Project); os.IsNotExist(err) {
			os.MkdirAll(odaConf.Dirs.Project, 0o755)
		}

		odaConf.System.SSHKey = filepath.Join(HOME, ".ssh/id_rsa")

		yamlOdaConfOut, err := yaml.Marshal(odaConf)
		if err != nil {
			fmt.Fprintln(os.Stderr, "yaml marshalling error %w", err)
		}
		odaFile := filepath.Join(CFGDIR, "oda", "oda.yaml")
		if _, err := os.Stat(odaFile); os.IsNotExist(err) {
			os.MkdirAll(filepath.Join(CFGDIR, "oda"), 0o750)
			os.WriteFile(odaFile, yamlOdaConfOut, 0o640)
		}
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
}

type OdaDatabase struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type OdaDirs struct {
	Repo    string `json:"repo"`
	Project string `json:"project"`
}
type OdaIncus struct {
	Socket string `json:"socket"`
	Type   string `json:"type"`
	URL    string `json:"url"`
}
type OdaSystem struct {
	Domain string `json:"domain"`
	SSHKey string `json:"ssh_key"`
}
type OdaConf struct {
	Database OdaDatabase `json:"database"`
	Dirs     OdaDirs     `json:"dirs"`
	Incus    OdaIncus    `json:"incus"`
	System   OdaSystem   `json:"system"`
}

func genOdaConf() OdaConf {
	return OdaConf{
		Database: OdaDatabase{
			Host:     "db",
			Port:     5432,
			Username: "odoo",
			Password: "odooodoo",
		},
		Dirs: OdaDirs{
			Repo:    "/home/odoo/workspace/repos/odoo",
			Project: "/home/odoo/workspace/odoo",
		},
		Incus: OdaIncus{
			Socket: "/var/lib/incus/unix.socket",
			Type:   "unix",
			URL:    "http://unix.socket/1.0",
		},
		System: OdaSystem{
			Domain: "local",
			SSHKey: "/home/odoo/.ssh/id_rsa",
		},
	}
}
