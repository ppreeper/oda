package oda

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/odoojrpc"
)

func InstanceStart() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	fmt.Println("Start the instance")
	cwd, project := GetProject()
	dirs := GetDirs()
	version := GetVersion()

	pods, _ := GetPods(true)
	for _, pod := range pods {
		if strings.Contains(pod.Name, project) && strings.HasPrefix(pod.Status, "Up") {
			// Check to see if the instance is running
			fmt.Println(project+".local", "is already running")
			return nil
		}
		if strings.Contains(pod.Name, project) &&
			(strings.HasPrefix(pod.Status, "Created") || strings.HasPrefix(pod.Status, "Exited")) {
			// Check to see if the instance is in invalid state
			// Remove if in invalid state
			InstanceStop()
		}
	}

	out, err := exec.Command("podman",
		"run", "--rm", "-d",
		"--publish-all",
		"--userns", "keep-id",
		"--name", project+".local",
		"-v", cwd+"/conf:/opt/odoo/conf:ro",
		"-v", cwd+"/data:/opt/odoo/data",
		"-v", cwd+"/addons:/opt/odoo/addons",
		"-v", filepath.Join(dirs.Repo, version, "odoo")+":/opt/odoo/odoo:ro",
		"-v", filepath.Join(dirs.Repo, version, "enterprise")+":/opt/odoo/enterprise:ro",
		"-v", filepath.Join(dirs.Project, "backups")+":/opt/odoo/backups",
		"ghcr.io/ppreeper/odoobase:"+version,
	).Output()
	if err != nil {
		return err
	}
	fmt.Println(project+".local", "started", string(out[0:12]))
	ProxyGenerate()
	ProxyStop()
	ProxyStart()
	return nil
}

func InstanceStop() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	if err := exec.Command("podman", "stop", project+".local").Run(); err != nil {
		fmt.Println("stopping: ", err)
	}
	if err := exec.Command("podman", "rm", project+".local").Run(); err != nil {
		fmt.Println("removing: ", err)
	}
	fmt.Println(project+".local", "stopped")
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

func InstanceAppInstallUpgrade(install bool, modules ...string) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}

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

	_, project := GetProject()
	podCmd := exec.Command("podman",
		"exec", "-it",
		project+".local",
		"odoo/odoo-bin",
		"--no-http", "--stop-after-init",
		iu, modList,
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstanceScaffold(module string) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	podCmd := exec.Command("podman",
		"exec", "-it",
		project+".local",
		"odoo/odoo-bin",
		"scaffold",
		module,
		filepath.Join("/opt/odoo/addons/", module),
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstancePS() error {
	out, err := exec.Command("podman", "ps", "--format", "{{.Image}};{{.Names}}").Output()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	pods := strings.Split(string(out), "\n")
	conf := GetConf()
	for _, pod := range pods {
		if strings.Contains(pod, conf.Odoobase) {
			podSplit := strings.Split(pod, ";")
			fmt.Println(podSplit[1])
		}
	}
	return nil
}

func InstanceExec() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	podCmd := exec.Command("podman",
		"exec", "-it",
		project+".local",
		"/bin/bash",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstanceLogs() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	podCmd := exec.Command("podman",
		"logs", "-f",
		project+".local",
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
	cwd, project := GetProject()

	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

	podCmd := exec.Command("podman",
		"exec", "-it",
		"-e", "PGPASSWORD="+dbpassword,
		project+".local",
		"psql", "-h", dbhost,
		"-U", dbuser,
		dbname,
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func InstanceQuery(q *QueryDef) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	cwd, project := GetProject()

	dbname := GetOdooConf(cwd, "db_name")

	oc := odoojrpc.NewOdoo().
		WithHostname(project + ".local").
		WithPort(443).
		WithDatabase(dbname).
		WithUsername("admin").
		WithPassword("admin").
		WithSchema("https")

	err := oc.Login()
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
	fmt.Println("user1", user1, "user2", user2)

	if user1 != user2 {
		return fmt.Errorf("usernames entered do not match")
	}

	cwd, _ := GetProject()

	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

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

	db.MustExec("update res_users set login=$1 where id=2;",
		strings.TrimSpace(string(user1)))

	fmt.Println("Admin username changed to", user1)
	return nil
}

func AdminPassword() error {
	fmt.Println("admin password")
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
	passkey, err := exec.Command("podman",
		"exec", "-it", project+".local", "oda_db.py", "-p", password1).Output()
	if err != nil {
		return fmt.Errorf("error generating password %w", err)
	}
	// salt := []byte(randStr(22))
	// iteration := 350_000
	// key := pbkdf2.Key([]byte(password1), salt, iteration, sha512.Size, sha512.New)
	// pass := base64.RawStdEncoding.EncodeToString(key)

	// password := fmt.Sprintf("\\$pbkdf2-sha512\\$%d\\$%s\\$%s", iteration, string(salt), pass)

	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

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

	db.MustExec("update res_users set password=$1 where id=2;",
		strings.TrimSpace(string(passkey)))
	return nil
}

// n is the length of random string we want to generate
// func randStr(n int) string {
// 	charset := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
// 	b := make([]byte, n)
// 	for i := range b {
// 		// randomly select 1 character from given charset
// 		b[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return string(b)
// }

type Pod struct {
	Image  string
	Name   string
	Ports  map[string]string
	Status string
}

func GetPods(all bool) ([]Pod, error) {
	podCmd := []string{"ps", "--format", "{{.Image}};{{.Names}};{{.Ports}};{{.Status}}"}
	if all {
		podCmd = append(podCmd, "-a")
	}
	out, err := exec.Command("podman", podCmd...).Output()
	if err != nil {
		return nil, err
	}
	podList := strings.Split(string(out), "\n")
	slices.Sort(podList)
	pods := []Pod{}
	for _, pod := range podList {

		podSplit := strings.Split(pod, ";")
		if len(podSplit) != 4 {
			continue
		}

		aPod := Pod{
			Image:  podSplit[0],
			Name:   podSplit[1],
			Ports:  make(map[string]string),
			Status: podSplit[3],
		}

		ports := strings.Split(podSplit[2], ",")
		if len(ports) == 1 && ports[0] == "" {
			continue
		}
		for _, port := range ports {
			portSplit := strings.Split(port, "->")
			source := strings.Split(portSplit[0], ":")
			dest := strings.Split(portSplit[1], "/")
			aPod.Ports[dest[0]] = source[1]
		}
		pods = append(pods, aPod)
	}
	return pods, nil
}
