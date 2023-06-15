package cmd

import (
	"log"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the instance",
	Long:  `Start the instance`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		startOdoo()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the instance",
	Long:  `Stop the instance`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		stopOdoo()
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the instance",
	Long:  `Restart the instance`,
	Run: func(cmd *cobra.Command, args []string) {
		stopOdoo()
		time.Sleep(2 * time.Second)
		startOdoo()
	},
}

func startOdoo() {
	cwd, _ := getCwd()
	odooPort := parseFile(".envrc", "ODOO_PORT")

	c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--http-port", odooPort)
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
	c.Process.Release()
}

func stopOdoo() {
	cwd, _ := getCwd()
	c := exec.Command("pkill", "-f", cwd+"/odoo/odoo-bin")
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
