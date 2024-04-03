package oda

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/odoojrpc"
)

func InstanceCreate() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	version := GetVersion()
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)

	if err := IncusCopy(vers, project); err != nil {
		return err
	}
	IncusStop(project)
	if err := IncusIdmap(project); err != nil {
		return err
	}
	return nil
}

func InstanceDestroy() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	confim := AreYouSure("destroy the " + project + " instance")
	if !confim {
		return fmt.Errorf("destroying the " + project + " instance canceled")
	}
	if err := IncusDelete(project); err != nil {
		return err
	}
	return nil
}

func InstanceRebuild() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	confim := AreYouSure("rebuild the " + project + " instance")
	if !confim {
		return fmt.Errorf("rebuild the " + project + " instance canceled")
	}
	if err := IncusDelete(project); err != nil {
		return err
	}
	fmt.Println("Image " + project + " deleted")
	if err := InstanceCreate(); err != nil {
		return err
	}
	return nil
}

func InstanceStart() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	conf := GetConf()
	_, project := GetProject()

	_, err := GetContainer(project)
	if err != nil {
		InstanceCreate()
	}

	if err := IncusStart(project); err != nil {
		fmt.Println(err)
		return err
	}

	if err := IncusHosts(project, GetConf().Domain); err != nil {
		fmt.Println(err)
		return err
	}

	if err := IncusCaddyfile(project, GetConf().Domain); err != nil {
		fmt.Println("IncusCaddyfile", err)
		return err
	}

	if err := InstanceMounts(project); err != nil {
		fmt.Println("InstanceMounts", err)
		return err
	}
	time.Sleep(2 * time.Second)
	if err := SSHConfigGenerate(project); err != nil {
		fmt.Println("SSHConfigGenerate", err)
		return err
	}
	if err = exec.Command("sshconfig").Run(); err != nil {
		fmt.Println("sshconfig", err)
		return err
	}

	// ProxyGenerate()
	// ProxyStop()
	// ProxyStart()
	fmt.Println(project + "." + conf.Domain + " started")

	return nil
}

func InstanceStop() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	conf := GetConf()
	_, project := GetProject()
	IncusStop(project)
	fmt.Println(project + "." + conf.Domain + " stopped")
	return nil
}

func InstanceRestart() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	if err := InstanceStop(); err != nil {
		return err
	}
	if err := InstanceStart(); err != nil {
		return err
	}
	return nil
}

func InstanceMounts(project string) error {
	conf := GetConf()
	cwd, _ := GetProject()
	version := GetVersion()

	IncusMount(project, "odoo", conf.Repo+"/"+version+"/odoo", "/opt/odoo/odoo")
	IncusMount(project, "enterprise", conf.Repo+"/"+version+"/enterprise", "/opt/odoo/enterprise")
	IncusMount(project, "backups", conf.Project+"/backups", "/opt/odoo/backups")
	IncusMount(project, "addons", cwd+"/addons", "/opt/odoo/addons")
	IncusMount(project, "conf", cwd+"/conf", "/opt/odoo/conf")
	IncusMount(project, "data", cwd+"/data", "/opt/odoo/data")

	return nil
}

func InstanceAppInstallUpgrade(install bool, modules ...string) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()

	iu := "-u"

	if install {
		iu = "-i"
	}

	mods := []string{}
	for _, mod := range modules {
		mm := strings.Split(mod, ",")
		mods = append(mods, mm...)
	}

	modList := strings.Join(mods, ",")

	uid, err := IncusGetUid(project, "odoo")
	if err != nil {
		fmt.Println(err)
		return err
	}

	incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"odoo/odoo-bin",
		"--no-http", "--stop-after-init",
		"-c", "/opt/odoo/conf/odoo.conf",
		iu, modList,
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		return err
	}

	return nil
}

func InstanceScaffold(module string) error {
	// watch for nested creation

	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	uid, err := IncusGetUid(project, "odoo")
	if err != nil {
		fmt.Println(err)
		return err
	}

	incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"odoo/odoo-bin",
		"scaffold", module, "/opt/odoo/addons",
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstancePS() error {
	projects := GetCurrentOdooProjects()
	containers, err := GetContainers()
	if err != nil {
		return fmt.Errorf("containers list failed %w", err)
	}

	containerList := []Container{}

	for _, container := range containers {
		for _, project := range projects {
			if container.Name == project {
				containerList = append(containerList, container)
			}
		}
	}

	maxnameLen := 0
	maxstateLen := 0
	maxipv4Len := 15

	for _, container := range containerList {
		if len(container.Name) > maxnameLen {
			maxnameLen = len(container.Name)
		}
		if len(container.State) > maxstateLen {
			maxstateLen = len(container.State)
		}
	}

	fmt.Printf("%-*s %-*s %-*s\n",
		maxnameLen+2, "NAME",
		maxstateLen+2, "STATE",
		maxipv4Len+2, "IPV4",
	)
	for _, container := range containerList {
		fmt.Printf("%-*s %-*s %-*s\n",
			maxnameLen+2, container.Name,
			maxstateLen+2, container.State,
			maxipv4Len+2, container.IP4,
		)
	}
	return nil
}

func InstanceExec(username string) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()

	uid, err := IncusGetUid(project, username)
	if err != nil {
		fmt.Println(err)
		return err
	}

	incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"/bin/bash",
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstanceLogs() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}

	_, project := GetProject()
	podCmd := exec.Command("incus",
		"exec", project, "-t", "--",
		"tail", "-f", "/var/log/syslog",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstancePSQL() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	conf := GetConf()
	cwd, _ := GetProject()

	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

	uid, err := IncusGetUid(conf.DBHost, "postgres")
	if err != nil {
		fmt.Println(err)
		return err
	}

	incusCmd := exec.Command("incus", "exec", conf.DBHost, "--user", uid,
		"--env", "PGPASSWORD="+dbpassword, "-t", "--",
		"psql", "-h", "127.0.0.1", "-U", dbuser, dbname,
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstanceQuery(q *QueryDef) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	// conf := GetConf()
	cwd, project := GetProject()
	container, err := GetContainer(project)
	if err != nil {
		return fmt.Errorf("error getting container %w", err)
	}

	dbname := GetOdooConf(cwd, "db_name")

	oc := odoojrpc.NewOdoo().
		WithHostname(container.IP4).
		WithPort(8069).
		WithDatabase(dbname).
		WithUsername(q.Username).
		WithPassword(q.Password).
		WithSchema("http")

	err = oc.Login()
	if err != nil {
		return err
	}

	umdl := strings.Replace(q.Model, "_", ".", -1)

	fields := parseFields(q.Fields)
	if q.Count {
		fields = []string{"id"}
	}

	filtp, err := parseFilter(q.Filter)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
	}

	rr, err := oc.SearchRead(umdl, filtp, q.Offset, q.Limit, fields)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
	}
	if q.Count {
		fmt.Println("records:", len(rr))
	} else {
		jsonStr, err := json.MarshalIndent(rr, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err.Error())
		}
		fmt.Println(string(jsonStr))
	}

	return nil
}

func AdminUsername() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	var user1, user2 string
	huh.NewInput().
		Title("Please enter  the new admin username:").
		Prompt(">").
		Value(&user1).
		Run()
	huh.NewInput().
		Title("Please verify the new admin username:").
		Prompt(">").
		Value(&user2).
		Run()

	if user1 != user2 {
		return fmt.Errorf("usernames entered do not match")
	}

	conf := GetConf()
	cwd, _ := GetProject()

	container, err := GetContainer(conf.DBHost)
	if err != nil {
		return fmt.Errorf("error getting container %w", err)
	}

	dbhost := container.IP4
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Port:     5432,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
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

	_, err = db.Exec("update res_users set login=$1 where id=2;",
		strings.TrimSpace(string(user1)))
	if err != nil {
		return fmt.Errorf("error updating username %w", err)
	}

	fmt.Println("Admin username changed to", user1)
	return nil
}

func AdminPassword() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	var password1, password2 string
	huh.NewInput().
		Title("Please enter  the admin password:").
		Prompt(">").
		Password(true).
		Value(&password1).
		Run()
	huh.NewInput().
		Title("Please verify the admin password:").
		Prompt(">").
		Password(true).
		Value(&password2).
		Run()
	if password1 != password2 {
		return fmt.Errorf("passwords entered do not match")
	}
	var confirm bool
	huh.NewConfirm().
		Title("Are you sure you want to change the admin password?").
		Affirmative("yes").
		Negative("no").
		Value(&confirm).
		Run()
	if !confirm {
		return fmt.Errorf("password change cancelled")
	}
	cwd, project := GetProject()

	uid, err := IncusGetUid(project, "odoo")
	if err != nil {
		fmt.Println(err)
		return err
	}

	passkey, err := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"oda", "password", password1,
	).Output()
	if err != nil {
		return fmt.Errorf("error generating password %w", err)
	}

	conf := GetConf()

	container, err := GetContainer(conf.DBHost)
	if err != nil {
		return fmt.Errorf("error getting container %w", err)
	}

	dbhost := container.IP4
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
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

	_, err = db.Exec("update res_users set password=$1 where id=2;",
		strings.TrimSpace(string(passkey)))
	if err != nil {
		return fmt.Errorf("error updating password %w", err)
	}
	fmt.Println("admin password changed")

	return nil
}

func SSHConfigGenerate(project string) error {
	HOME, _ := os.UserHomeDir()
	conf := GetConf()
	container, err := GetContainer(project)
	if err != nil {
		return fmt.Errorf("error getting container %w", err)
	}
	// priority;host;hostname;user;identityfile;port
	sshconfig := fmt.Sprintf("%d;%s.%s;%s;%s;%s;%d", 10, project, conf.Domain, container.IP4, "odoo", conf.SSHKey, 22)
	sshconfigCSV := filepath.Join(HOME, ".ssh", "sshconfig.csv")
	// READ config
	hosts, err := os.Open(sshconfigCSV)
	if err != nil {
		return fmt.Errorf("hosts file read failed %w", err)
	}
	defer hosts.Close()
	hostlines := []string{}
	scanner := bufio.NewScanner(hosts)
	for scanner.Scan() {
		hostlines = append(hostlines, scanner.Text())
	}
	headerLine := hostlines[0]
	// Remove old lines
	newHostlines := []string{}
	for _, hostline := range hostlines[1:] {
		hostlineSplit := strings.Split(hostline, ";")
		if hostlineSplit[1] == project+"."+conf.Domain {
			continue
		}
		newHostlines = append(newHostlines, hostline)
	}
	// Add new lines
	newHostlines = append(newHostlines, sshconfig)
	// WRITE config
	fo, err := os.Create(sshconfigCSV)
	if err != nil {
		return err
	}
	defer fo.Close()
	fo.WriteString(headerLine + "\n")
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	return nil
}
