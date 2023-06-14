package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nxadm/tail"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var res embed.FS

func parseConfig(key string) string {
	// Get variable from odoo config
	file, err := os.Open("./conf/odoo.conf")
	if err != nil {
		return "cannot find file, make sure you are in the odoo project folder"
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	value := ""
	vv := []string{}
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			vv = strings.Split(scanner.Text(), "=")
			for i := range vv {
				vv[i] = strings.TrimSpace(vv[i])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "scanner error"
	}
	if len(vv) == 2 {
		value = vv[1]
	}
	return value
}

func parseEnv(param string) (string, error) {
	// Get variable from envrc
	file, err := os.Open(".envrc")
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

func parseFile(filename, key string) (value string) {
	file, err := os.Open(filename)
	if err != nil {
		return "cannot find file, make sure you are in the odoo project folder"
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	vv := []string{}
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), key) {
			vv = strings.Split(scanner.Text(), "=")
			for i := range vv {
				vv[i] = strings.TrimSpace(vv[i])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "scanner error"
	}
	if len(vv) == 2 {
		value = vv[1]
	}
	return value
}

func pipfile() {
	pipFile, err := template.ParseFS(res, "templates/Pipfile")
	if err != nil {
		panic(err)
	}
	file, _ := os.Create("Pipfile")
	defer file.Close()
	pipFile.Execute(file, nil)
}

func envrc(version string, port int16) {
	envrc, err := template.ParseFS(res, "templates/envrc")
	if err != nil {
		panic(err)
	}
	envrcMap := make(map[string]interface{})
	envrcMap["Version"] = version
	envrcMap["Port"] = port
	file, _ := os.Create(".envrc")
	defer file.Close()
	envrc.Execute(file, envrcMap)
}

func odooconf(version string, port, dbPort int16, dbHost, dbName string) {
	odooConf, err := template.ParseFS(res, "templates/odoo.conf")
	if err != nil {
		panic(err)
	}
	odooConfMap := make(map[string]interface{})
	odooConfMap["Version"] = version
	odooConfMap["Port"] = port
	odooConfMap["DBHost"] = dbHost
	odooConfMap["DBPort"] = dbPort
	odooConfMap["DBName"] = dbName
	file, _ := os.Create("conf/odoo.conf")
	defer file.Close()
	odooConf.Execute(file, odooConfMap)
}

func initProject(version string, port, dbPort int16) {
	paths := []string{"addons", "backups", "conf", "data"}
	epaths := []string{"data", "backups"}
	envrc(version, port)
	cmd := exec.Command("direnv", "allow")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}
	for _, path := range paths {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	for _, path := range epaths {
		err := os.Chmod(path, 0o777)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	dirname, _ := os.UserHomeDir()
	repodir := dirname + "/workspace/repos/" + version + ".0/"
	err := os.Symlink(repodir+"odoo", "odoo")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.Symlink(repodir+"enterprise", "enterprise")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, cdir := getCwd()

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}
	for _, a := range addrs {
		fmt.Println(a)
	}
	ip := getIP()

	odooconf(version, port, dbPort, ip, cdir)
	pipfile()
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

func getIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	addrs, err := interfaces[1].Addrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var ip net.IP
	switch v := addrs[0].(type) {
	case *net.IPAddr:
		fmt.Println("IPAddr", v)
		ip = v.IP
	case *net.IPNet:
		fmt.Println("IPNet", v)
		ip = v.IP
	default:
		break
	}
	return ip.String()
}

func startOdoo() {
	cwd, _ := getCwd()
	odooPort, err := parseEnv("ODOO_PORT")
	if err != nil {
		fmt.Println(err)
		return
	}

	c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--http-port", odooPort)
	if err := c.Start(); err != nil {
		fmt.Println(err)
		return
	}
	c.Process.Release()
}

func stopOdoo() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	c := exec.Command("pkill", "-f", cwd+"/odoo/odoo-bin")
	if err := c.Run(); err != nil {
		fmt.Println(err)
		return
	}
}

func loggerOut() {
	logfile := parseConfig("logfile")

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

func dsn() string {
	dbUser := parseConfig("db_user")
	dbPass := parseConfig("db_pass")
	dbHost := parseConfig("db_host")
	dbPort := parseConfig("db_port")
	dbName := parseConfig("db_name")
	return "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName
}

func manifest(dbName string) (output string) {
	conn, err := pgx.Connect(context.Background(), dsn())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	output = "{" + "\n"
	output += `    "odoo_dump": "1",` + "\n"
	output += `    "db_name": "` + dbName + `",` + "\n"

	versionInfo := parseFile("./odoo/odoo/release.py", "version_info")
	versionInfo = strings.Trim(versionInfo, "(")
	versionInfo = strings.Trim(versionInfo, ")")
	version := strings.Split(versionInfo, ",")
	fmt.Println("versionInfo", versionInfo)
	fmt.Println("version", version)
	output += `    "version": "` + strings.TrimSpace(version[0]) + "." + strings.TrimSpace(version[1]) + `"` + ",\n"
	output += `    "version_info": [` + "\n"
	for k, v := range version {
		suffix := ",\n"
		if k == len(version)-1 {
			suffix = "\n"
		}
		val := strings.TrimSpace(v)
		if strings.ToUpper(val) == "FINAL" {
			val = `"final"`
		}
		output += `        ` + val + suffix
	}

	output += `    ],` + "\n"
	output += `    "major_version": "` + strings.TrimSpace(version[0]) + "." + strings.TrimSpace(version[1]) + `"` + ",\n"

	var pgVersion string
	err = conn.QueryRow(context.Background(), "SHOW server_version").Scan(&pgVersion)
	if err != nil {
		log.Fatal(err)
	}
	output += `    "pg_version": "` + pgVersion + `",` + "\n"

	type Module struct {
		name          string
		latestVersion string
	}
	modules := []Module{}
	rows, err := conn.Query(context.Background(), "SELECT name, latest_version FROM ir_module_module WHERE state = 'installed'")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Fatal(err)
		}
		modules = append(modules, Module{name: values[0].(string), latestVersion: values[1].(string)})
	}
	output += `    "modules": {` + "\n"
	for i, module := range modules {
		suffix := ",\n"
		if i == len(modules)-1 {
			suffix = "\n"
		}
		output += `        "` + module.name + `": "` + module.latestVersion + `"` + suffix
	}
	output += `    }` + "\n"
	output += "}"
	return
}

func dumpDB(dbName string, bkp string) {
	fmt.Println(dbName, bkp)
	folder := "./backups"
	bkpFile := path.Join(folder, bkp+".zip")
	fmt.Println(bkpFile)

	dataDir := parseConfig("data_dir")
	filestore := path.Join(dataDir, "filestore", dbName)
	fmt.Println(filestore)

	tPath := path.Join(os.TempDir(), bkp)
	fmt.Println(tPath)
	tFilestore := path.Join(tPath, "filestore")
	fmt.Println(tFilestore)

	// Filestore
	err := cp.Copy(filestore, tFilestore)
	if err != nil {
		log.Fatal(err)
	}

	// SQL Dump
	c := exec.Command("pg_dump", dsn(), "--no-owner", "--file", path.Join(tPath, "dump.sql"))
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
	// dump_db_manifest(cr)
	manifest := manifest(dbName)
	mjson, err := os.Create(path.Join(tPath, "manifest.json"))
	if err != nil {
		log.Fatal(err)
	}
	defer mjson.Close()
	mjson.WriteString(manifest)

	// write zip file

	// if err := zipSource(tPath, bkpFile); err != nil {
	// 	log.Fatal(err)
	// }
	zipWriter(tPath, bkpFile)

	// def _dump_db(db_name, bkp_name, folder="./backups", backup_format='zip'):
	// """Dump database `db` into file-like object `stream` if stream is None
	// return a file object with the dump """
	// bkp_file = f"{bkp_name}.zip"
	// file_path = os.path.join(folder, bkp_file)
	// with open(file_path, 'wb') as stream:
	//     _logger.info('DUMP DB: %s format %s', db_name, backup_format)

	//     cmd = [find_pg_tool('pg_dump'), '--no-owner', db_name]
	//     env = exec_pg_environ()

	//     if backup_format == 'zip':
	//         with tempfile.TemporaryDirectory() as dump_dir:
	//             filestore = odoo.tools.config.filestore(db_name)
	//             if os.path.exists(filestore):
	//                 shutil.copytree(filestore,
	//                                 os.path.join(dump_dir, 'filestore'))
	//             with open(os.path.join(dump_dir, 'manifest.json'), 'w') as fh:
	//                 db = odoo.sql_db.db_connect(db_name)
	//                 with db.cursor() as cr:
	//                     json.dump(dump_db_manifest(cr), fh, indent=4)
	//             cmd.insert(-1, '--file=' + os.path.join(dump_dir, 'dump.sql'))
	//             subprocess.run(cmd,
	//                            env=env,
	//                            stdout=subprocess.DEVNULL,
	//                            stderr=subprocess.STDOUT,
	//                            check=True)
	//             if stream:
	//                 odoo.tools.osutil.zip_dir(
	//                     dump_dir,
	//                     stream,
	//                     include_dir=False,
	//                     fnct_sort=lambda file_name: file_name != 'dump.sql')
	//             else:
	//                 t = tempfile.TemporaryFile()
	//                 odoo.tools.osutil.zip_dir(
	//                     dump_dir,
	//                     t,
	//                     include_dir=False,
	//                     fnct_sort=lambda file_name: file_name != 'dump.sql')
	//                 t.seek(0)
	//                 return t
	//     else:
	//         cmd.insert(-1, '--format=c')
	//         stdout = subprocess.Popen(cmd,
	//                                   env=env,
	//                                   stdin=subprocess.DEVNULL,
	//                                   stdout=subprocess.PIPE).stdout
	//         if stream:
	//             shutil.copyfileobj(stdout, stream)
	//         else:
	//             return stdout
	//     return file_path
}

func dumpAddon(name string, bkp string) {
	fmt.Println(name, bkp)
}

func main() {
	var dumpFile string

	// Project Commands (Possible Destructive)

	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "project commands for init/destruction",
	}

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Drop database and filestore [CAUTION]",
		Long:  `Drop database and filestore [CAUTION]`,
		Run: func(cmd *cobra.Command, args []string) {
			resetProject()
		},
	}

	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Fully Destroy the project and its files [CAUTION]",
		Long:  `Fully Destroy the project and its files [CAUTION]`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			destroyProject()
		},
	}

	initprojectCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new project",
		Long:  `Create a new project`,
		Args:  cobra.MaximumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			port := int16(8069)
			dbPort := int16(5432)
			if len(args) < 1 {
				fmt.Println("invalid odoo version, valid: <15|16>")
				return
			}

			switch args[0] {
			case "15", "16":
				if len(args) < 1 {
					fmt.Println("require")
					return
				}
				version := string(args[0])
				if len(args) == 2 {
					p, err := strconv.ParseInt(args[1], 10, 16)
					if err != nil {
						fmt.Println("port must be an integer")
						return
					}
					port = int16(p)
				}
				if len(args) == 3 {
					p, err := strconv.ParseInt(args[1], 10, 16)
					if err != nil {
						fmt.Println("port must be an integer")
						return
					}
					d, err := strconv.ParseInt(args[2], 10, 16)
					if err != nil {
						fmt.Println("db_port must be an integer")
						return
					}
					port = int16(p)
					dbPort = int16(d)
				}
				initProject(version, port, dbPort)
			default:
				fmt.Println("invalid odoo version, valid: <15|16>")
			}
		},
	}

	// Odoo Database Commands

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the database",
		Long:  `Initialize the database`,
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := getCwd()

			dbName := parseConfig("db_name")

			c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--no-http", "--stop-after-init", "-d", dbName, "-i", "base,l10n_ca")

			if err := c.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install module(s) (comma seperated list)",
		Long:  `Install module(s) (comma seperated list)`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := getCwd()

			dbName := parseConfig("db_name")

			c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--no-http", "--stop-after-init", "-d", dbName, "-i", args[0])

			if err := c.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade module(s) (comma seperated list)",
		Long:  `Upgrade module(s) (comma seperated list)`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := getCwd()

			dbName := parseConfig("db_name")

			c := exec.Command(cwd+"/odoo/odoo-bin", "-c", cwd+"/conf/odoo.conf", "--no-http", "--stop-after-init", "-d", dbName, "-u", args[0])

			if err := c.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup database and filestore",
		Long:  "Backup database and filestore",
		Run: func(cmd *cobra.Command, args []string) {
			dbName := parseConfig("db_name")
			addonDirs := parseConfig("addons")
			addons := strings.Split(addonDirs, ",")[2:]

			t := time.Now()
			curDate := t.Format("2006_01_02_15_04_05")
			bkpName := curDate + "_" + dbName
			fmt.Printf("backup %s\n", bkpName)
			dumpDB(dbName, bkpName)
			for _, v := range addons {
				addon := strings.TrimPrefix(v, "./")
				fmt.Printf("backup %s_%s\n", bkpName, addon)
				dumpAddon(addon, bkpName)
			}
		},
	}

	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore database and filestore [CAUTION]",
		Long:  "Restore database and filestore [CAUTION]",
		Run: func(cmd *cobra.Command, args []string) {
			if dumpFile == "" {
				fmt.Println("please specify backup file")
				return
			}
			fmt.Println("restore database from", dumpFile)
		},
	}

	restoreAddonsCmd := &cobra.Command{
		Use:   "addons",
		Short: "Restore addons",
		Long:  "Restore addons",
		Run: func(cmd *cobra.Command, args []string) {
			if dumpFile == "" {
				fmt.Println("please specify backup file")
				return
			}
			fmt.Println("restore addons from", dumpFile)
		},
	}

	// Odoo Admin Commands

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Follow the logs",
		Long:  `Follow the logs`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			logfile := parseConfig("logfile")

			t, err := tail.TailFile(logfile, tail.Config{Follow: true, ReOpen: true})
			if err != nil {
				log.Fatal(err)
			}
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
		},
	}

	binCmd := &cobra.Command{
		Use:   "bin",
		Short: "Run an odoo-bin command",
		Long:  `Run an odoo-bin command`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("odoo/odoo-bin ", args)
		},
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the instance",
		Long:  `Start the instance`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			startOdoo()
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the instance",
		Long:  `Stop the instance`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			stopOdoo()
		},
	}

	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the instance",
		Long:  `Restart the instance`,
		Run: func(cmd *cobra.Command, args []string) {
			stopOdoo()
			time.Sleep(2 * time.Second)
			startOdoo()
		},
	}

	psqlCmd := &cobra.Command{
		Use:   "psql",
		Short: "Access the raw database",
		Long:  `Access the raw database`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dbHost := parseConfig("db_host")
			dbPort := parseConfig("db_port")
			dbName := parseConfig("db_name")
			dbUser := parseConfig("db_user")
			dbPass := parseConfig("db_pass")
			// psql -d "host=10.0.30.90 port=5432 dbname=odoo15c user=odoo16 password=odooodoo"
			goexecpath, _ := exec.LookPath("psql")
			fmt.Println(goexecpath, "postgresql://"+dbUser+":"+dbPass+"@"+dbHost+":"+dbPort+"/"+dbName)
			// cmdGoVer := &exec.Cmd{
			// 	Path:   goexecpath,
			// 	Args:   []string{"postgresql://" + dbUser + ":" + dbPass + "@127.0.0.1:" + dbPort + "/" + dbName},
			// 	Stdout: os.Stdout,
			// 	Stderr: os.Stdout,
			// }
			// fmt.Println(cmdGoVer.String())
			// if err := cmdGoVer.Start(); err != nil {
			// 	fmt.Println("error:", err)
			// }
			// cmdGoVer.Process.Release()
			// c := exec.Command(goexecpath, "postgresql://"+dbUser+":"+dbPass+"@"+dbHost+":"+dbPort+"/"+dbName)

			// if err := c.Start(); err != nil {
			// 	fmt.Println(err)
			// 	return
			// }
			// c.Wait()
			// fmt.Println(c.ProcessState)
			// c.Process.Release()

			// reader, err := c.StdoutPipe()
			// if err != nil {
			// 	return
			// }

			// scanner := bufio.NewScanner(reader)
			// go func() {
			// 	for scanner.Scan() {
			// 		line := scanner.Text()
			// 		fmt.Printf("%s\n", line)
			// 	}
			// }()

			// if err := cmdGoVer.Start(); err != nil {
			// 	fmt.Println(err)
			// 	return
			// }
			// if err := cmdGoVer.Wait(); err != nil {
			// 	fmt.Println(err)
			// 	return
			// }
		},
	}

	// Main Root

	rootCmd := &cobra.Command{
		Use:   "oda",
		Short: "Odoo Administration Tool",
		Long:  `Odoo Administration Tool`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(upgradeCmd)

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.AddCommand(restoreAddonsCmd)

	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(binCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(psqlCmd)

	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(resetCmd)
	projectCmd.AddCommand(destroyCmd)
	projectCmd.AddCommand(initprojectCmd)

	restoreCmd.PersistentFlags().StringVarP(&dumpFile, "dump_file", "d", "", "database dump file")
	restoreAddonsCmd.PersistentFlags().StringVarP(&dumpFile, "dump_file", "d", "", "database dump file")

	rootCmd.Execute()

	// var backup, restore, addons bool
	// flag.BoolVar(&backup, "b", false, "backup database")
	// flag.BoolVar(&restore, "r", false, "restore database")
	// flag.BoolVar(&addons, "a", false, "restore addons")
	// var dumpFile, folder string
	// flag.StringVar(&dumpFile, "d", "", "database dump file")
	// flag.StringVar(&folder, "f", "", "addons folder")
	// flag.Parse()

	// fmt.Println("backup", backup, "restore", restore, "addons", addons)
	// fmt.Println("dumpFile", dumpFile, "folder", folder)

	// dbName, err := parseConfig("db_name")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dbName)
	// dirList, err := parseConfig("addons")
	// if err != nil {
	// 	panic(err)
	// }
	// addonDirs := strings.Split(dirList, ",")[2:]
	// fmt.Println(addonDirs)

	// if !backup && !restore && !addons {
	// 	flag.PrintDefaults()
	// 	return
	// }

	// if backup && (restore || addons) {
	// 	fmt.Println("backup or restore cannot run both commands")
	// 	return
	// }

	// if backup {
	// 	//     bkp_name = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}_{db_name}"
	// 	//     print(_dump_db(db_name,bkp_name))
	// 	//     _dump_addons(addons,bkp_name)
	// 	return
	// }

	// if restore && dumpFile == "" {
	// 	fmt.Println("restore command requires a dump file to read")
	// 	return
	// }

	// if restore && dumpFile != "" {
	// 	fmt.Printf("restore from dump file %s\n", dumpFile)
	// 	//     _restore_db(db_name, args.dump_file)
	// 	return
	// }

	// if addons && dumpFile == "" {
	// 	fmt.Println("addons restore command requires a dump file to read")
	// 	return
	// }
	// %Y_%m_%d_%H_%M_%S')}_{db_name}
	// if addons && dumpFile != "" {
	// 	fmt.Printf("addons restore from dump file %s\n", dumpFile)
	// 	//     _restore_addons(args.dump_file) if (args.folder is None or args.folder == "") else _restore_addons(args.dump_file,args.folder)
	// 	return
	// }
}
