package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(binCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(psqlCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(initprojectCmd)
}

var rootCmd = &cobra.Command{
	Use:   "oda",
	Short: "Oda is an Odoo administration tool",
	Long:  `A Fast and Flexible Odoo administration tool`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init",
	Long:  `init`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_install $ODOODB base,l10n_ca && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install",
	Long:  `install`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_install $ODOODB ${2} && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade",
	Long:  `upgrade`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_upgrade $ODOODB ${2} && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
	},
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup the odoo project",
	Long:  "backup the odoo project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires backup dump filename")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("backup command executed", args[0])
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore the odoo project",
	Long:  "restore the odoo project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires backup dump filename")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("restore command executed", args[0])
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "logs",
	Long:  `logs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// tail -f odoo.log ;;
	},
}

var binCmd = &cobra.Command{
	Use:   "bin",
	Short: "bin",
	Long:  `bin`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// shift && odoo/odoo-bin $@ ;;
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start",
	Long:  `start`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_start $POD ;;
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop",
	Long:  `stop`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_stop $POD ;;
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart",
	Long:  `restart`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// odoo_stop $POD && sleep 2 && odoo_start $POD ;;
	},
}

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "psql",
	Long:  `psql`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// PGPASSWORD=$(gcfg db_pass) psql -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) $(gcfg db_name)
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset",
	Long:  `reset`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// read -r -p "Are you sure? [y/N] " response
		// if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
		// 	odoo_stop $POD
		// 	rm -rf data/* > /dev/null
		// 	PGPASSWORD=$(gcfg db_pass) dropdb -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) -f $(gcfg db_name)
		// fi
	},
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "destroy",
	Long:  `destroy`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// read -r -p "Are you sure you want to destroy everything? [YES/N] " response
		// if [[ "$response" =~ ^(YES)$ ]]; then
		//   read -r -p "Are you **really** sure you want to destroy everything? [YES/N] " response
		//   if [[ "$response" =~ ^(YES)$ ]]; then
		// 	echo "Destroying project"
		// 	odoo_stop $POD
		// 	sudo rm -rf .direnv/ addons/ backups/ conf/ data/ .envrc Pipfile enterprise odoo
		// 	PGPASSWORD=$(gcfg db_pass) dropdb -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) -f  $(gcfg db_name) >/dev/null
		// 	echo "Project has been destroyed"
		//   fi
		// fi
	},
}

var initprojectCmd = &cobra.Command{
	Use:   "initproject",
	Short: "initproject",
	Long:  `initproject`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		// initproject $2 ;;
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseConfig(param string) (string, error) {
	// Get variable from odoo config
	file, err := os.Open("./conf/odoo.conf")
	if err != nil {
		return "cannot find file, make sure you are in the odoo project folder", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	vv := []string{}
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), param) {
			vv = strings.Split(scanner.Text(), "=")
			for i := range vv {
				vv[i] = strings.TrimSpace(vv[i])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "scanner error", err
	}
	return vv[1], nil
}
