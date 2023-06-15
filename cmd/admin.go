package cmd

import (
	"fmt"
	"log"

	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Follow the logs",
	Long:  `Follow the logs`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		logfile := parseFile("conf/odoo.conf", "logfile")

		t, err := tail.TailFile(logfile, tail.Config{Follow: true, ReOpen: true})
		if err != nil {
			log.Fatal(err)
		}
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	},
}

// var binCmd = &cobra.Command{
// 	Use:   "bin",
// 	Short: "Run an odoo-bin command",
// 	Long:  `Run an odoo-bin command`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		cwd, _ := getCwd()

// 		fmt.Println(cwd+"/odoo/odoo-bin", args)
// 		c := exec.Command(cwd+"/odoo/odoo-bin", args...)
// 		if err := c.Start(); err != nil {
// 			fmt.Println(err)
// 			os.Exit(1)
// 		}
// 		c.Process.Release()
// 	},
// }
