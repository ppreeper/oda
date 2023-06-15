package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup database and filestore",
	Long:  "Backup database and filestore",
	Run: func(cmd *cobra.Command, args []string) {
		dbName := parseFile("conf/odoo.conf", "db_name")
		addonDirs := parseFile("conf/odoo.conf", "addons")
		addons := strings.Split(addonDirs, ",")[2:]

		t := time.Now()
		curDate := t.Format("2006_01_02_15_04_05")
		bkpName := curDate + "_" + dbName
		dumpDB("backups", dbName, bkpName)
		dumpAddons("backups", addons, bkpName)
	},
}

func dumpDB(folder string, dbName string, bkp string) {
	bkpFile := path.Join(folder, bkp+".zip")
	fmt.Println(bkpFile)

	dataDir := parseFile("conf/odoo.conf", "data_dir")
	filestore := path.Join(dataDir, "filestore", dbName)

	tPath := path.Join(os.TempDir(), bkp)
	tFilestore := path.Join(tPath, "filestore")

	// Filestore backup
	err := cp.Copy(filestore, tFilestore)
	if err != nil {
		log.Fatal(err)
	}
	// SQL Dump
	c := exec.Command("pg_dump", dsn(), "--no-owner", "--file", path.Join(tPath, "dump.sql"))
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
	// manifest
	manifest := dumpManifest(dbName)
	jsonString, _ := json.MarshalIndent(manifest, "", "  ")

	mjson, err := os.Create(path.Join(tPath, "manifest.json"))
	if err != nil {
		log.Fatal(err)
	}
	defer mjson.Close()
	mjson.Write(jsonString)

	// zip files
	err = zipSource(tPath, bkpFile)
	if err != nil {
		log.Fatal(err)
	}

	// remove tmp folder
	err = os.RemoveAll(tPath)
	if err != nil {
		log.Fatal(err)
	}
}

func dumpAddons(folder string, addons []string, bkp string) {
	for _, addon := range addons {
		base := filepath.Base(addon)
		bkpFile := path.Join(folder, bkp+"_"+base+".zip")
		fmt.Println(bkpFile)
		// zip files
		err := zipSource(addon, bkpFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func dumpManifest(dbName string) map[string]interface{} {
	conn, err := pgx.Connect(context.Background(), dsn())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	mm := make(map[string]interface{})
	mm["odoo_dump"] = "1"
	mm["db_name"] = dbName

	versionInfo := parseFile("./odoo/odoo/release.py", "version_info")
	versionInfo = strings.Trim(versionInfo, "(")
	versionInfo = strings.Trim(versionInfo, ")")
	version := strings.Split(versionInfo, ",")
	vv := []any{}
	for _, v := range version {
		i, err := strconv.Atoi(strings.Trim(strings.TrimSpace(v), "'"))
		if err != nil {
			vv = append(vv, strings.Trim(strings.TrimSpace(v), "'"))
		} else {
			vv = append(vv, i)
		}
	}

	mm["version"] = strings.TrimSpace(version[0]) + "." + strings.TrimSpace(version[1])
	mm["major_version"] = strings.TrimSpace(version[0]) + "." + strings.TrimSpace(version[1])
	mm["version_info"] = vv

	var pgVersion string
	err = conn.QueryRow(context.Background(), "SHOW server_version").Scan(&pgVersion)
	if err != nil {
		log.Fatal(err)
	}
	mm["pg_version"] = pgVersion

	modules := make(map[string]interface{})
	rows, err := conn.Query(context.Background(), "SELECT name, latest_version FROM ir_module_module WHERE state = 'installed'")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal(err)
		}
		modules[values[0].(string)] = values[1].(string)
	}
	mm["modules"] = modules
	return mm
}
