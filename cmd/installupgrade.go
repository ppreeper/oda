package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Initialize the database`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		installModule(true, "base,l10n_ca")
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install module(s) (comma seperated list)",
	Long:  `Install module(s) (comma seperated list)`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installModule(true, args[0])
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade module(s) (comma seperated list)",
	Long:  `Upgrade module(s) (comma seperated list)`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installModule(false, args[0])
	},
}

func installModule(install bool, modules string) {
	cwd, _ := getCwd()

	dbName := parseFile("conf/odoo.conf", "db_name")

	iFlag := "-i"
	if !install {
		iFlag = "-u"
	}

	// fmt.Println(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--no-http", "--stop-after-init", "-d", dbName, iFlag, modules)
	c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--no-http", "--stop-after-init", "-d", dbName, iFlag, modules)

	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
