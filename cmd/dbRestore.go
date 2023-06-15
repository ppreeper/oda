package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database and filestore [CAUTION]",
	Long:  "Restore database and filestore [CAUTION]",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, f := range args {
			filename := filepath.Base(f)
			filebase := strings.TrimSuffix(filename, filepath.Ext(filename))
			namesplit := strings.Split(filebase, "_")
			// fmt.Println(filename, filebase, namesplit)
			if len(namesplit) == 7 {
				restoreDB(f)
			}
			if len(namesplit) == 8 {
				restoreAddon(f)
			}
		}
	},
}

func restoreDB(bkpFile string) {
	filename := filepath.Base(bkpFile)
	filebase := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Create TMP folder to extract files
	tPath := path.Join(os.TempDir(), filebase)
	unzipSource(bkpFile, tPath)

	// Drop and Create Blank Database
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbName := parseFile("conf/odoo.conf", "db_name")
	dropDB(dbName)
	createEmptyDB(dbName)

	// Restore DB
	c := exec.Command("psql", "-q", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-f", filepath.Join(tPath, "dump.sql"), "-w", "-d", dbName)
	c.Env = append(c.Env, "PGPASSWORD="+dbPass)
	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Filestore restore
	dataDir := parseFile("conf/odoo.conf", "data_dir")
	filestore := filepath.Join(dataDir, "filestore", dbName)

	// remove current files
	err := os.RemoveAll(filestore)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// copy from tmp dir to
	tFilestore := filepath.Join(tPath, "filestore")
	err = cp.Copy(tFilestore, filestore)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Delete TMP folder to extract files
	err = os.RemoveAll(tPath)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func restoreAddon(bkpFile string) {
	filename := filepath.Base(bkpFile)
	filebase := strings.TrimSuffix(filename, filepath.Ext(filename))
	namesplit := strings.Split(filebase, "_")
	addonDir := namesplit[7]
	err := os.RemoveAll(addonDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	unzipSource(bkpFile, addonDir)
}
