/*
Copyright © 2024 Peter Preeper

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package internal

import (
	_ "embed"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed commit.txt
var Commit string

var rootCmd = &cobra.Command{
	Use:   "oda",
	Short: "Odoo Client Administration Tool",
	Long:  `Odoo Client Administration Tool`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Commit
	rootCmd.AddGroup(
		&cobra.Group{ID: "app", Title: "App Management"},
		&cobra.Group{ID: "backup", Title: "Backup Management"},
		&cobra.Group{ID: "config", Title: "Config Commands"},
		&cobra.Group{ID: "database", Title: "Database Management"},
		&cobra.Group{ID: "image", Title: "Image Management"},
		&cobra.Group{ID: "instance", Title: "Instance Management"},
		&cobra.Group{ID: "project", Title: "Project Commands"},
		&cobra.Group{ID: "repo", Title: "Repo Management"},
		&cobra.Group{ID: "user", Title: "Admin User Management"},
	)
}

func initConfig() {
	var odaConfigNotExists bool
	sudouser, _ := os.LookupEnv("SUDO_USER")
	if sudouser != "" {
		// Root User
		odauser, err := user.Lookup(sudouser)
		cobra.CheckErr(err)
		cfgdir := path.Join(odauser.HomeDir, ".config", "oda")
		viper.SetConfigType("yaml")
		// Search config in config directory with name "oda.yaml" (without extension).
		viper.SetConfigName("oda")
		viper.AddConfigPath(cfgdir)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "oda.yaml file missing")
			odaConfigNotExists = true
		}
		if odaConfigNotExists {
			fmt.Fprintln(os.Stderr, "oda config file missing, please run: config init")
		}
	} else {
		// Regular user
		cfgdir, err := os.UserConfigDir()
		cobra.CheckErr(err)
		cfgdir = filepath.Join(cfgdir, "oda")
		viper.SetConfigType("yaml")
		// Search config in config directory with name "oda.yaml" (without extension).
		viper.SetConfigName("oda")
		viper.AddConfigPath(cfgdir)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "oda.yaml file missing")
			odaConfigNotExists = true
		}
		if odaConfigNotExists {
			fmt.Fprintln(os.Stderr, "oda config file missing, please run: config init")
		}
	}

	// Search config in current directory with name ".oda.yaml" (without extension).
	viper.SetConfigName(".oda")
	viper.AddConfigPath(".")
	viper.MergeInConfig() // nolint:errcheck // ignore error
}
