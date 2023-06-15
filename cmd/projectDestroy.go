package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Drop database and filestore [CAUTION]",
	Long:  `Drop database and filestore [CAUTION]`,
	Run: func(cmd *cobra.Command, args []string) {
		resetProject()
	},
}

func resetProject() {
	conf1 := stringPrompt("Are you sure you want to reset everything? [YES/N]")
	conf2 := stringPrompt("Are you **really** sure you want to reset everything? [YES/N]")
	if conf1 == "YES" && conf2 == "YES" {
		// TODO: stop odoo
		dirs, err := os.ReadDir("data")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, d := range dirs {
			name := "data/" + d.Name()
			if d.IsDir() {
				err := os.RemoveAll(name)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				err := os.Remove(name)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
		// TODO: Drop Database
		// 	PGPASSWORD=$(gcfg db_pass) dropdb -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) -f $(gcfg db_name)
	} else {
		fmt.Println("Project not reset")
	}
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Fully Destroy the project and its files [CAUTION]",
	Long:  `Fully Destroy the project and its files [CAUTION]`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		destroyProject()
	},
}

func destroyProject() {
	conf1 := stringPrompt("Are you sure you want to destroy everything? [YES/N]")
	conf2 := stringPrompt("Are you **really** sure you want to destroy everything? [YES/N]")
	if conf1 == "YES" && conf2 == "YES" {
		// TODO: stop odoo
		fmt.Println("Destroying project")
		dirs, err := os.ReadDir(".")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, d := range dirs {
			if d.IsDir() {
				err := os.RemoveAll(d.Name())
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				err := os.Remove(d.Name())
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}

		// TODO: Drop Database
		fmt.Println("Project has been destroyed")
	} else {
		fmt.Println("Project not destroyed")
	}
}
