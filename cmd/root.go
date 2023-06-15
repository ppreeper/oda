package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oda",
	Short: "Odoo Administration Tool",
	Long:  `Odoo Administration Tool`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(upgradeCmd)

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)

	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)

	// rootCmd.AddCommand(binCmd)
	// rootCmd.AddCommand(psqlCmd)

	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(resetCmd)
	projectCmd.AddCommand(destroyCmd)
	projectCmd.AddCommand(initprojectCmd)
}

func stringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func getCwd() (cwd string, cdir string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	cdirs := strings.Split(cwd, "/")
	cdir = cdirs[len(cdirs)-1]
	return
}
