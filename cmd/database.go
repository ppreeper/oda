package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func dsn() string {
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	dbName := parseFile("conf/odoo.conf", "db_name")
	return "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName
}

func dsnLite() string {
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	return "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort
}

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Access the raw database",
	Long:  `Access the raw database`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("psql", dsn())
		c := exec.Command("psql", dsn())
		if err := c.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(1)
	},
}

func dropDB(dbName string) {
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")

	// db drop
	c := exec.Command("dropdb", "--force", "--if-exists", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-w", dbName)
	c.Env = append(c.Env, "PGPASSWORD="+dbPass)
	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createEmptyDB(dbName string) {
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbTemplate := parseFile("conf/odoo.conf", "db_template")

	// db create
	c := exec.Command("createdb", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-O", dbUser, "-T", dbTemplate, "--lc-collate", "C", "-E", "UTF8", "-w", dbName)
	c.Env = append(c.Env, "PGPASSWORD="+dbPass)
	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
