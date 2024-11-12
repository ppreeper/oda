package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

func (o *ODA) Restore(any, move, full bool) error {
	if !IsProject() {
		return nil
	}

	_, project := lib.GetProject()

	var backups []string
	var addons []string

	if any {
		backups, addons = GetOdooBackups("")
	} else {
		backups, addons = GetOdooBackups(project)
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
		fmt.Fprintln(os.Stderr, ui.WarningStyle.Render("restore cancelled"))
		return nil
	}

	if addonFile != "none" {
		fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restore from addon file "+addonFile))
		if err := restoreAddonsTar(addonFile); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("restore addons tar failed", err.Error()))
			return nil
		}
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restore from backup file "+backupFile))
	if err := restoreDBTar(backupFile, move); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("restore db tar failed", err.Error()))
		return nil
	}
	return nil
}

// restoreAddonsTar Restore Odoo DB addons folders
func restoreAddonsTar(addonsFile string) error {
	cwd, _ := lib.GetProject()
	odaConf, _ := config.LoadOdaConfig()
	source := filepath.Join(odaConf.Dirs.Project, "backups", addonsFile)
	dest := filepath.Join(cwd, "addons")
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

// // restoreDBTar Restore Odoo DB from backup
func restoreDBTar(backupFile string, moveDB bool) error {
	cwd, project := lib.GetProject()
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)
	odooConf, _ := config.LoadOdooConfig(cwd)
	source := filepath.Join(odaConf.Dirs.Project, "backups", backupFile)

	dbname := odooConf.DbName
	dbhost := odooConf.DbHost
	dbuser := odooConf.DbUser
	dbpassword := odooConf.DbPassword
	dbtemplate := odooConf.DbTemplate

	dbserver := dbhost
	if dbhost == "localhost" {
		dbserver = project
	}
	uid, err := inc.IncusGetUid(dbserver, "postgres")
	if err != nil {
		return fmt.Errorf("could not get postgres uid %w", err)
	}

	// drop target database
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("drop target database"))
	if err := exec.Command("incus", "exec", dbserver, "--user", uid, "-t", "--",
		"dropdb", "--if-exists", "-U", "postgres", "-f", dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("dropped database "+dbname))

	// create new postgresql database
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("create new postgresql database"))
	if err := exec.Command("incus", "exec", dbserver, "--user", uid, "-t", "--",
		"createdb", "-U", "postgres",
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", dbtemplate,
		"-O", dbuser, dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", dbname, err)
	}
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("created database "+dbname))

	// restore postgresql database
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restore postgresql database"))
	tarpgCmd := exec.Command("tar", "Oaxf", source, "./dump.sql")
	dbhostTarget := dbhost + "." + odaConf.System.Domain
	if dbhost == "localhost" {
		dbInstance, _ := inc.GetInstance(project)
		fmt.Println(dbInstance)
		dbhostTarget = dbInstance.IP4
	}

	pgCmd := exec.Command("incus", "exec", dbserver, "--user", uid,
		"--env", "PGPASSWORD="+dbpassword, "--",
		"psql", "-h", dbhostTarget, "-U", dbuser, "--dbname", dbname, "-q")
	pgCmd.Env = append(pgCmd.Env, "PGPASSWORD="+dbpassword)

	r, w := io.Pipe()
	tarpgCmd.Stdout = w
	pgCmd.Stdin = r

	tarpgCmd.Start()
	pgCmd.Start()
	tarpgCmd.Wait()
	w.Close()
	pgCmd.Wait()
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restored database "+dbname))

	// restore data filestore
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restore postgresql database"))
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
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("restored filestore "+dbname))

	// if not moveDB then reset DBUUID and remove MCode
	if !moveDB {
		fmt.Fprintln(os.Stderr, ui.StepStyle.Render("neutralize the database"))
		if err := dbReset(dbhostTarget, dbname, dbuser, dbpassword); err != nil {
			return fmt.Errorf("db reset failed %w", err)
		}
	}
	// fmt.Println("reset database " + dbname)
	return nil
}

// func dbClone(dbhost, sourceDB, destDB, dbuser, dbpassword string) error {
// 	db, err := OpenDatabase(Database{
// 		Hostname: dbhost,
// 		Username: dbuser,
// 		Password: dbpassword,
// 		Database: "postgres",
// 	})
// 	defer func() error {
// 		if err := db.Close(); err != nil {
// 			return fmt.Errorf("error closing database %w", err)
// 		}
// 		return nil
// 	}()
// 	if err != nil {
// 		return fmt.Errorf("error opening database %w", err)
// 	}

// 	db.Exec("drop database if exists " + destDB)
// 	db.Exec("select pg_terminate_backend (pid) from pg_stat_activity where datname=$1", sourceDB)
// 	db.Exec(fmt.Sprintf("create database %s with template %s owner %s", destDB, sourceDB, dbuser))

// 	return nil
// }

// dbReset Database Reset DBUUID and remove MCode
func dbReset(dbhost, dbname, dbuser, dbpassword string) error {
	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
	})
	defer func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database %w", err)
		}
		return nil
	}()
	if err != nil {
		return fmt.Errorf("error opening database %w", err)
	}

	// -- remove the enterprise code, report.url and web.base.url
	db.Exec("delete from ir_config_parameter where key in ('database.enterprise_code', 'report.url', 'web.base.url.freeze')")

	// reset db uuid
	db.Exec(`update ir_config_parameter set value=(select gen_random_uuid()) where key = 'database.uuid'`)

	// update expiration date
	db.Exec(`insert into ir_config_parameter
			(key,value,create_uid,create_date,write_uid,write_date)
			values
			('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
			current_timestamp,1,current_timestamp)
			on conflict (key)
			do UPDATE set value = (current_date+'3 months'::interval)::timestamp;`)

	// disable bank synchronisation links
	db.Exec(`UPDATE account_online_link SET provider_data = '', client_id = 'duplicate';`)

	// deactivate fetchmail server
	db.Exec("UPDATE fetchmail_server SET active = false;")

	// deactivate mail servers but activate default "localhost" mail server
	db.Exec(`DO $$
			        BEGIN
			            UPDATE ir_mail_server SET active = 'f';
			            IF EXISTS (SELECT 1 FROM ir_module_module WHERE name='mail' and state IN ('installed', 'to upgrade', 'to remove')) THEN
			                UPDATE mail_template SET mail_server_id = NULL;
			            END IF;
			        EXCEPTION
			            WHEN undefined_table OR undefined_column THEN
			        END;
			    $$;`)

	// deactivate crons
	db.Exec("UPDATE ir_cron SET active = 'f';")
	db.Exec(`UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'autovacuum_job' AND module = 'base');`)

	// activating module update notification cron
	db.Exec(`UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'ir_cron_module_update_notification' AND module = 'mail');`)

	// remove platform ir_logging
	db.Exec("DELETE FROM ir_logging WHERE func = 'odoo.sh';")

	// disable prod environment in all delivery carriers
	db.Exec("UPDATE delivery_carrier SET prod_environment = false;")

	// disable delivery carriers from external providers
	db.Exec("UPDATE delivery_carrier SET active = false WHERE delivery_type NOT IN ('fixed', 'base_on_rule');")

	// disable iap account
	db.Exec(`UPDATE iap_account SET account_token = REGEXP_REPLACE(account_token, '(\+.*)?$', '+disabled');`)

	// deactivate mail template
	db.Exec("UPDATE mail_template SET mail_server_id = NULL;")

	// disable generic payment provider
	db.Exec("UPDATE payment_provider SET state = 'disabled' WHERE state NOT IN ('test', 'disabled');")

	// delete domains on websites
	db.Exec("UPDATE website SET domain = NULL;")

	// disable cdn
	db.Exec("UPDATE website SET cdn_activated = false;")

	// delete odoo_ocn.project_id and ocn.uuid
	db.Exec("DELETE FROM ir_config_parameter WHERE key IN ('odoo_ocn.project_id', 'ocn.uuid');")

	// delete Facebook Access Tokens
	db.Exec("UPDATE social_account SET facebook_account_id = NULL, facebook_access_token = NULL;")

	// delete Instagram Access Tokens
	db.Exec("UPDATE social_account SET instagram_account_id = NULL, instagram_facebook_account_id = NULL, instagram_access_token = NULL;")

	// delete LinkedIn Access Tokens
	db.Exec("UPDATE social_account SET linkedin_account_urn = NULL, linkedin_access_token = NULL;")

	// delete Twitter Access Tokens
	db.Exec("UPDATE social_account SET twitter_user_id = NULL, twitter_oauth_token = NULL, twitter_oauth_token_secret = NULL;")

	// delete Youtube Access Tokens
	db.Exec("UPDATE social_account SET youtube_channel_id = NULL, youtube_access_token = NULL, youtube_refresh_token = NULL, youtube_token_expiration_date = NULL, youtube_upload_playlist_id = NULL;")

	// Unset Firebase configuration within website
	db.Exec("UPDATE website SET firebase_enable_push_notifications = false, firebase_use_own_account = false, firebase_project_id = NULL, firebase_web_api_key = NULL, firebase_push_certificate_key = NULL, firebase_sender_id = NULL;")

	// Remove Map Box Token as it's only valid per DB url
	db.Exec("DELETE FROM ir_config_parameter WHERE key = 'web_map.token_map_box';")

	// activate neutralization watermarks banner
	// db.Exec("UPDATE ir_ui_view SET active = true WHERE key = 'web.neutralize_banner';")
	// activate neutralization watermarks ribbon
	// db.Exec("UPDATE ir_ui_view SET active = true WHERE key = 'website.neutralize_ribbon';")

	return nil
}
