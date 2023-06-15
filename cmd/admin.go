package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

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

var binCmd = &cobra.Command{
	Use:   "bin",
	Short: "Run an odoo-bin command",
	Long:  `Run an odoo-bin command`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("odoo/odoo-bin ", args)
	},
}

func loggerOut() {
	logfile := parseFile("conf/odoo.conf", "logfile")

	c := exec.Command("tail", "-f", logfile)

	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	// scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}
	}()

	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
}
