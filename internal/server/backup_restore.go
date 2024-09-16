package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

func AdminBackup() error {
	// Get the current date and time
	currentTime := time.Now()
	// Format the time as a string
	timeString := currentTime.Format("2006_01_02_15_04_05")
	// main database and filestore
	dumpDBTar(timeString)
	// 	addons
	dumpAddonsTar(timeString)
	return nil
}

func dumpAddonsTar(bkp_prefix string) error {
	db_name := GetOdooConf("", "db_name")
	addons_path := GetOdooConf("", "addons_path")
	addons := strings.Split(addons_path, ",")[2:]
	for _, addon := range addons {
		folder := strings.Replace(addon, "/opt/odoo/", "", 1)
		dirlist, _ := os.ReadDir(addon)
		if len(dirlist) != 0 {
			tar_cmd := "tar"
			bkp_file := fmt.Sprintf("%s__%s__%s.tar.zst", bkp_prefix, db_name, folder)
			file_path := filepath.Join("/opt/odoo/backups", bkp_file)
			tar_args := []string{"ahcf", file_path, "-C", addon, "."}
			cmd := exec.Command(tar_cmd, tar_args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("addons backup %s failed: %w", bkp_file, err)
			}
			fmt.Println("addons:", file_path)
		}
	}
	return nil
}

func dumpDBTar(bkp_prefix string) error {
	db_host := GetOdooConf("", "db_host")
	db_port, _ := strconv.Atoi(GetOdooConf("", "db_port"))
	db_name := GetOdooConf("", "db_name")
	db_user := GetOdooConf("", "db_user")
	db_password := GetOdooConf("", "db_password")
	data_dir := GetOdooConf("", "data_dir")
	bkp_file := fmt.Sprintf("%s__%s.tar.zst", bkp_prefix, db_name)
	dump_dir := filepath.Join("/opt/odoo/backups", fmt.Sprintf("%s__%s", bkp_prefix, db_name))
	file_path := filepath.Join("/opt/odoo/backups", bkp_file)

	// create dump_dir
	if err := os.MkdirAll(dump_dir, 0o755); err != nil {
		return fmt.Errorf("directory already exists: %w", err)
	}

	// postgresql database
	pg_cmd := exec.Command("pg_dump",
		"-h", db_host,
		"-p", fmt.Sprintf("%d", db_port),
		"-U", db_user,
		"--no-owner",
		"--file", filepath.Join(dump_dir, "dump.sql"),
		db_name,
	)
	pg_cmd.Env = append(pg_cmd.Env, "PGPASSWORD="+db_password)
	pg_cmd.Stdin = os.Stdin
	pg_cmd.Stdout = os.Stdout
	pg_cmd.Stderr = os.Stderr
	if err := pg_cmd.Run(); err != nil {
		return fmt.Errorf("could not backup postgresql database %s: %w", db_name, err)
	}

	// filestore
	filestore := filepath.Join(data_dir, "filestore", db_name)
	filestore_back := filepath.Join(dump_dir, "filestore")
	if _, err := os.Stat(filestore); err == nil {
		if err := os.Symlink(filestore, filestore_back); err != nil {
			return fmt.Errorf("symlink failed: %w", err)
		}
	}

	// create tar archive
	tar_cmd := exec.Command("tar",
		"achf", file_path, "-C", dump_dir, ".",
	)
	tar_cmd.Stdin = os.Stdin
	tar_cmd.Stdout = os.Stdout
	tar_cmd.Stderr = os.Stderr
	if err := tar_cmd.Run(); err != nil {
		return fmt.Errorf("could not backup database %s: %w", db_name, err)
	}

	// cleanup dump_dir
	if err := os.RemoveAll(dump_dir); err != nil {
		return fmt.Errorf("could not cleanup dump_dir: %w", err)
	}

	fmt.Println("odoo:", file_path)

	return nil
}

// AdminRestore Restore from backup file
func AdminRestore(any, move, full bool) error {
	fmt.Println("AdminRestore", "any:", any, "move:", move, "full:", full)
	var backups []string
	var addons []string
	db_name := GetOdooConf("", "db_name")

	if any {
		backups, addons = GetOdooBackups("")
	} else {
		backups, addons = GetOdooBackups(db_name)
	}

	backupOptions := []huh.Option[string]{}
	for _, backup := range backups {
		backupOptions = append(backupOptions, huh.NewOption(backup, backup))
	}
	addonOptions := []huh.Option[string]{}
	addonOptions = append(addonOptions, huh.NewOption("None", "none"))
	for _, addon := range addons {
		addonOptions = append(addonOptions, huh.NewOption(addon, addon))
	}

	var (
		backupFile string
		addonFile  string
		confirm    bool
	)

	huh.NewSelect[string]().
		Title("Odoo Backup File").
		Options(backupOptions...).
		Value(&backupFile).
		Run()

	huh.NewSelect[string]().
		Title("Odoo Addon File").
		Options(addonOptions...).
		Value(&addonFile).
		Run()

	huh.NewConfirm().
		Title("Restore Project?").
		Value(&confirm).
		Run()

	if !confirm {
		fmt.Println("restore cancelled")
		return nil
	}

	if addonFile != "none" {
		fmt.Println("restore from addon file " + addonFile)
		if err := restoreAddonsTar(addonFile); err != nil {
			return fmt.Errorf("restore addons tar failed %w", err)
		}
	}

	fmt.Println("restore from backup file " + backupFile)
	if err := restoreDBTar(backupFile, move, full); err != nil {
		return fmt.Errorf("restore db tar failed %w", err)
	}

	return nil
}

// restoreAddonsTar Restore Odoo DB addons folders
func restoreAddonsTar(addonsFile string) error {
	root_dir := filepath.Join("/", "opt", "odoo")
	source := filepath.Join(root_dir, "backups", addonsFile)
	dest := filepath.Join(root_dir, "addons")

	if err := RemoveContents(dest); err != nil {
		return fmt.Errorf("remove contents failed: %w", err)
	}
	cmd := exec.Command("tar",
		"axf", source, "-C", dest, ".",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract addon files failed: %w", err)
	}
	return nil
}

// restoreDBTar Restore Odoo DB from backup
func restoreDBTar(backupFile string, moveDB bool, neutralize bool) error {
	// Stop Odoo Service
	ServiceStop()

	cwd := filepath.Join("/", "opt", "odoo")
	source := filepath.Join(cwd, "backups", backupFile)

	dbname := GetOdooConf(cwd, "db_name")
	dbhost := GetOdooConf(cwd, "db_host")
	dbport := GetOdooConf(cwd, "db_port")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbtemplate := GetOdooConf(cwd, "db_template")
	port, _ := strconv.Atoi(dbport)

	odb := OdooDB{
		Hostname: dbhost,
		Port:     dbport,
		Database: dbname,
		Username: dbuser,
		Password: dbpassword,
		Template: dbtemplate,
	}

	// drop target database
	if err := odb.DropDatabase(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}

	// create new postgresql database
	if err := odb.CreateDatabase(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", dbname, err)
	}

	// restore postgresql database
	if err := odb.RestoreDatabase(source); err != nil {
		return fmt.Errorf("could not restore postgresql database %s error: %w", dbname, err)
	}

	// restore data filestore
	// fmt.Println("restore postgresql database")
	data := filepath.Join(cwd, "data")
	if err := RemoveContents(data); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}
	filestore := filepath.Join(data, "filestore", dbname)
	if err := os.MkdirAll(filestore, 0o755); err != nil {
		return fmt.Errorf("filestore directory creation failed %w", err)
	}
	tarCmd := exec.Command("tar",
		"axf", source, "-C", filestore, "--strip-components=2", "./filestore",
	)
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("filestore restore failed %w", err)
	}
	// fmt.Println("restored filestore " + dbname)

	// if not moveDB then reset DBUUID and remove MCode
	if !moveDB {
		fmt.Println("neutralize the database")
		db, err := OpenDatabase(Database{
			Hostname: dbhost,
			Port:     port,
			Database: dbname,
			Username: dbuser,
			Password: dbpassword,
		})
		if err != nil {
			return fmt.Errorf("error opening database %w", err)
		}
		defer func() error {
			if err := db.Close(); err != nil {
				return fmt.Errorf("error closing database %w", err)
			}
			return nil
		}()

		db.RemoveEnterpriseCode()
		db.ChangeDBUUID()
		db.UpdateDatabaseExpirationDate()
		db.DisableBankSync()
		db.DisableFetchmail()
		db.DeactivateMailServers()
		db.DeactivateCrons()
		db.ActivateModuleUpdateNotificationCron()
		db.RemoveIRLogging()
		db.DisableProdDeliveryCarriers()
		db.DisableDeliveryCarriers()
		db.DisableIAPAccount()
		db.DisableMailTemplate()
		db.DisablePaymentGeneric()
		db.DeleteWebsiteDomains()
		db.DisableCDN()
		db.DeleteOCNProjectUUID()
		db.UnsetFirebase()
		db.RemoveMapBoxToken()

		// Social Media
		db.RemoveFacebookTokens()
		db.RemoveInstagramTokens()
		db.RemoveLinkedInTokens()
		db.RemoveTwitterTokens()
		db.RemoveYoutubeTokens()

		if neutralize {
			db.ActivateNeutralizationWatermarks()
		}
	}
	return nil
}

// Trim database backups
func Trim(limit int, all bool) error {
	db_name := GetOdooConf("", "db_name")

	// # Get all backup files
	backups, addons := GetOdooBackups("")

	// # Group backup files by database name
	bkpFiles := make(map[string][]string)
	for _, k := range backups {
		fname := strings.Split(k, "__")
		dname := strings.TrimRight(fname[1], ".tar.zst")
		curFiles := bkpFiles[dname]
		bkpFiles[dname] = append(curFiles, k)
	}
	backupKeys := make([]string, 0, len(bkpFiles))
	for k := range bkpFiles {
		backupKeys = append(backupKeys, k)
	}
	slices.Sort(backupKeys)

	rmbkp := []string{}
	for _, k := range backupKeys {
		if len(bkpFiles[k]) > limit {
			rmbkp = append(rmbkp, bkpFiles[k][:len(bkpFiles[k])-limit]...)
		}
	}

	// # Group addon files by database name
	addonFiles := make(map[string][]string)
	for _, k := range addons {
		fname := strings.Split(k, "__")
		dname := fname[1]
		curFiles := addonFiles[dname]
		addonFiles[dname] = append(curFiles, k)
	}
	addonKeys := make([]string, 0, len(addonFiles))
	for k := range addonFiles {
		addonKeys = append(addonKeys, k)
	}
	slices.Sort(addonKeys)

	rmaddons := []string{}
	for _, k := range addonKeys {
		if len(addonFiles[k]) > limit {
			rmaddons = append(rmaddons, addonFiles[k][:len(addonFiles[k])-limit]...)
		}
	}

	// Join rmbackups and rmaddons
	rmlist := []string{}
	if all {
		rmlist = append(rmlist, rmbkp...)
		rmlist = append(rmlist, rmaddons...)
	} else {
		for _, k := range rmbkp {
			if strings.Contains(k, db_name) {
				rmlist = append(rmlist, k)
			}
		}
		for _, k := range rmaddons {
			if strings.Contains(k, db_name) {
				rmlist = append(rmlist, k)
			}
		}
	}

	for _, r := range rmlist {
		backupFile := filepath.Join("/", "opt", "odoo", "backups", r)
		// fmt.Println("rm", "-f", backupFile)
		os.Remove(backupFile)
	}

	return nil
}
