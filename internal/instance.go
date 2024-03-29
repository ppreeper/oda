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

func InstanceCreate() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	conf := GetConf()
	_, project := GetProject()
	if err := IncusLaunch(project, conf.OSImage); err != nil {
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "update", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("apt-get update")

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "install", "-y", "wget", "git").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("apt-get install pre-requisites")

	// REPO="https://github.com/wkhtmltopdf/packaging"
	// vers=$(git ls-remote --tags ${REPO} | grep "refs/tags.*[0-9]$" | awk '{print $2}' | sed 's/refs\/tags\///' | sort -V | uniq | tail -1)
	// VC=$(grep ^VERSION_CODENAME /etc/os-release | awk -F'=' '{print $2}')
	// UC=$(grep ^UBUNTU_CODENAME /etc/os-release | awk -F'=' '{print $2}')
	// CN=''
	// [ -n "$UC" ] && CN=$UC || CN=$VC
	// FN="wkhtmltox_${vers}.${CN}_amd64.deb"0

	if err := exec.Command("incus", "exec", project, "-t", "--", "groupadd", "-f", "-g", "1001", "odoo").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "useradd", "-ms", "/bin/bash", "-g", "1001", "-u", "1001", "odoo").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// # mkdir mount directories
	if err := exec.Command("incus", "exec", project, "-t", "--",
		"mkdir", "-p", "/opt/odoo/{addons,conf,data,backups,odoo,enterprise}",
	).Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--",
		"chown", "odoo:odoo", "/opt/odoo",
	).Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// sudo bash -c "groupadd -f -g 1001 odoo"
	// if [ -f $(grep "^odoo:" /etc/passwd) ]; then
	//   sudo bash -c "useradd -ms /usr/sbin/nologin -g 1001 -u 1001 odoo"
	// fi

	// # PostgreSQL Repo
	// sudo wget -qO /etc/apt/trusted.gpg.d/pgdg.gpg.asc https://www.postgresql.org/media/keys/ACCC4CF8.asc
	// echo "deb http://apt.postgresql.org/pub/repos/apt/ ${CN}-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list
	// sudo bash -c "apt-get update -y"

	// # postgresql
	// sudo apt-get install -y --no-install-recommends postgresql-client-15

	// # install wkhtmltopdf
	// wget -qc ${REPO}/releases/download/${vers}/${FN} -O ${HOME}/wkhtmltox.deb
	// sudo apt-get install -y --no-install-recommends ${HOME}/wkhtmltox.deb
	// rm -rf ${HOME}/wkhtmltox.deb
	// sudo bash -c "apt-get update -y"

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt", "upgrade", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "install", "-y", "--no-install-recommends",
		"bzip2",
		"ca-certificates",
		"curl",
		"dirmngr",
		"fonts-liberation",
		"fonts-noto",
		"fonts-noto-cjk",
		"fonts-noto-mono",
		"geoip-database",
		"gnupg",
		"gsfonts",
		"inetutils-ping",
		"libgnutls-dane0",
		"libgts-bin",
		"libpaper-utils",
		"locales",
		"nodejs",
		"npm",
		"python3",
		"python3-babel",
		"python3-chardet",
		"python3-cryptography",
		"python3-cups",
		"python3-dateutil",
		"python3-decorator",
		"python3-docutils",
		"python3-feedparser",
		"python3-freezegun",
		"python3-geoip2",
		"python3-gevent",
		"python3-greenlet",
		"python3-html2text",
		"python3-idna",
		"python3-jinja2",
		"python3-ldap",
		"python3-libsass",
		"python3-lxml",
		"python3-markupsafe",
		"python3-num2words",
		"python3-ofxparse",
		"python3-olefile",
		"python3-openssl",
		"python3-paramiko",
		"python3-passlib",
		"python3-pdfminer",
		"python3-phonenumbers",
		"python3-pil",
		"python3-pip",
		"python3-polib",
		"python3-psutil",
		"python3-psycopg2",
		"python3-pydot",
		"python3-pylibdmtx",
		"python3-pyparsing",
		"python3-pypdf2",
		"python3-pytzdata",
		"python3-qrcode",
		"python3-renderpm",
		"python3-reportlab",
		"python3-reportlab-accel",
		"python3-requests",
		"python3-rjsmin",
		"python3-serial",
		"python3-setuptools",
		"python3-stdnum",
		"python3-urllib3",
		"python3-usb",
		"python3-vobject",
		"python3-werkzeug",
		"python3-xlrd",
		"python3-xlsxwriter",
		"python3-xlwt",
		"python3-zeep",
		"shared-mime-info",
		"unzip",
		"xz-utils",
		"zip").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// # install geolite databases
	geolite := [][]string{
		{"GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"},
		{"GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"},
		{"GeoLite2-Country.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"},
	}

	for _, geo := range geolite {
		fmt.Println("downloading", geo[0])
		if err := exec.Command("incus", "exec", project, "-t", "--",
			"wget", "-qO", "/usr/share/GeoIP/"+geo[0], geo[1],
		).Run(); err != nil {
			fmt.Println(err)
			return err
		}
	}

	// # install additional python libraries
	if err := exec.Command("incus", "exec", project, "-t", "--",
		"sudo", "pip3", "install", "--break-system-packages", "ebaysdk", "google-auth",
	).Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// # install additional node libraries
	if err := exec.Command("incus", "exec", project, "-t", "--",
		"npm", "install", "-g", "rtlcss",
	).Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// # update system
	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "update", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "dist-upgrade", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "autoremove", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", project, "-t", "--", "apt-get", "autoclean", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := IncusRestart(project); err != nil {
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

func InstanceStart() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	fmt.Println("Start the instance")
	conf := GetConf()
	cwd, project := GetProject()
	dirs := GetDirs()
	version := GetVersion()

	fmt.Println(cwd, project, dirs, version)

	// pods, _ := GetPods(true)
	// for _, pod := range pods {
	// 	if strings.Contains(pod.Name, project) && strings.HasPrefix(pod.Status, "Up") {
	// 		// Check to see if the instance is running
	// 		fmt.Println(project+".local", "is already running")
	// 		return nil
	// 	}
	// 	if strings.Contains(pod.Name, project) &&
	// 		(strings.HasPrefix(pod.Status, "Created") || strings.HasPrefix(pod.Status, "Exited")) {
	// 		// Check to see if the instance is in invalid state
	// 		// Remove if in invalid state
	// 		InstanceStop()
	// 	}
	// }
	fmt.Println("incus", "launch", "images:"+conf.OSImage, project)

	// out, err := exec.Command("podman",
	// 	"run", "--rm", "-d",
	// 	"--publish-all",
	// 	"--userns", "keep-id",
	// 	"--name", project+".local",
	// 	"-v", cwd+"/conf:/opt/odoo/conf:ro",
	// 	"-v", cwd+"/data:/opt/odoo/data",
	// 	"-v", cwd+"/addons:/opt/odoo/addons",
	// 	"-v", filepath.Join(dirs.Repo, version, "odoo")+":/opt/odoo/odoo:ro",
	// 	"-v", filepath.Join(dirs.Repo, version, "enterprise")+":/opt/odoo/enterprise:ro",
	// 	"-v", filepath.Join(dirs.Project, "backups")+":/opt/odoo/backups",
	// 	"ghcr.io/ppreeper/odoobase:"+version,
	// ).Output()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(project+".local", "started", string(out[0:12]))
	// ProxyGenerate()
	// ProxyStop()
	// ProxyStart()
	return nil
}

func InstanceStop() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	// if err := exec.Command("podman", "stop", project+".local").Run(); err != nil {
	// 	fmt.Println("stopping: ", err)
	// }
	// if err := exec.Command("podman", "rm", project+".local").Run(); err != nil {
	// 	fmt.Println("removing: ", err)
	// }
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
	// watch for nested creation

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

	incusCmd := exec.Command("incus", "exec", project, "-t", "--", "/bin/bash")
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
	cwd, _ := GetProject()

	// dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")

	conf := GetConf()
	uid, err := PgdbGetuid(conf.DBH)
	if err != nil {
		fmt.Println(err)
		return err
	}

	pgCmd := exec.Command("incus", "exec", conf.DBH,
		"--env", "PGPASSWORD="+dbpassword,
		"--user", uid, "-t", "--",
		"psql", "-h", "127.0.0.1", "-U", dbuser, dbname)
	pgCmd.Stdin = os.Stdin
	pgCmd.Stdout = os.Stdout
	pgCmd.Stderr = os.Stderr
	if err := pgCmd.Run(); err != nil {
		fmt.Println(err)
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
