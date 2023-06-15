package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
	namesplit := strings.Split(filebase, "_")
	fmt.Println("filestore", filebase, namesplit)
	tPath := path.Join(os.TempDir(), filebase)
	unzipSource(bkpFile, tPath)
	// tFilestore := path.Join(tPath, "filestore")
	dbName := parseFile("conf/odoo.conf", "db_name")
	expDrop(dbName)

	// exp_drop(db)
	// _create_empty_database(db)

	// filestore_path = None
	// with tempfile.TemporaryDirectory() as dump_dir:
	//     if zipfile.is_zipfile(dump_file):
	//         # v8 format
	//         with zipfile.ZipFile(dump_file, 'r') as z:
	//             # only extract known members!
	//             filestore = [
	//                 m for m in z.namelist() if m.startswith('filestore/')
	//             ]
	//             z.extractall(dump_dir, ['dump.sql'] + filestore)

	//             if filestore:
	//                 filestore_path = os.path.join(dump_dir, 'filestore')

	//         pg_cmd = 'psql'
	//         pg_args = ['-q', '-f', os.path.join(dump_dir, 'dump.sql')]

	//     r = subprocess.run(
	//         [find_pg_tool(pg_cmd), '--dbname=' + db, *pg_args],
	//         env=exec_pg_environ(),
	//         stdout=subprocess.DEVNULL,
	//         stderr=subprocess.STDOUT,
	//     )
	//     if r.returncode != 0:
	//         raise Exception("Couldn't restore database")

	//     registry = odoo.modules.registry.Registry.new(db)
	//     with registry.cursor() as cr:
	//         env = odoo.api.Environment(cr, SUPERUSER_ID, {})
	//         if copy:
	//             # if it's a copy of a database, force generation of a new dbuuid
	//             env['ir.config_parameter'].init(force=True)
	//         if neutralize_database:
	//             odoo.modules.neutralize.neutralize_database(cr)

	//         if filestore_path:
	//             filestore_dest = env['ir.attachment']._filestore()
	//             shutil.move(filestore_path, filestore_dest)

	// _logger.info('RESTORE DB: %s', db)
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
