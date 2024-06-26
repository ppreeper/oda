package oda

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

func BaseCreatePrompt() error {
	var (
		version string
		create  bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Create Odoo Base Image?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("create base form error %w", err)
	}
	if err := BaseCreate(version); err != nil {
		return fmt.Errorf("create base %s error %w", version, err)
	}
	return nil
}

func BaseDestroyPrompt() error {
	var (
		version string
		destroy bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Destroy Odoo Base Image?").
				Value(&destroy),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("destroy base form error %w", err)
	}
	if err := BaseDestroy(version); err != nil {
		return fmt.Errorf("destroy base %s error %w", version, err)
	}
	return nil
}

func BaseRebuildPrompt() error {
	var (
		version string
		rebuild bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Rebuild Odoo Base Image?").
				Value(&rebuild),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("rebuild base form error %w", err)
	}
	if err := BaseDestroy(version); err != nil {
		return fmt.Errorf("destroy base %s error %w", version, err)
	}
	if err := BaseCreate(version); err != nil {
		return fmt.Errorf("create base %s error %w", version, err)
	}
	return nil
}

func BaseUpdatePrompt() error {
	var (
		version string
		update  bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Update Odoo base image packages?").
				Value(&update),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("updating base form error %w", err)
	}
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)
	container, err := GetContainer(vers)
	if err != nil {
		return fmt.Errorf("get container %s error %w", vers, err)
	}
	switch container.State {
	case "STOPPED":
		if err := IncusStart(vers); err != nil {
			return fmt.Errorf("start container %s error %w", vers, err)
		}
		if err := roleUpdate(vers); err != nil {
			return fmt.Errorf("update container %s error %w", vers, err)
		}
		if err := IncusStop(vers); err != nil {
			return fmt.Errorf("stop container %s error %w", vers, err)
		}
	case "RUNNING":
		if err := roleUpdate(vers); err != nil {
			return fmt.Errorf("update container %s error %w", vers, err)
		}
	}
	return nil
}

////////////////////////

func BaseCreate(version string) error {
	conf := GetConf()
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)
	if err := IncusLaunch(vers, conf.OSImage); err != nil {
		return fmt.Errorf("launch container %s %w", vers, err)
	}
	fmt.Println("launching:", vers)

	fmt.Println("roleUpdate:", vers)
	if err := roleUpdate(vers); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", vers, err)
	}

	fmt.Println("roleBaseline:", vers)
	if err := roleBaseline(vers); err != nil {
		return fmt.Errorf("roleBaseline %s failed %w", vers, err)
	}

	fmt.Println("roleOdooUser:", vers)
	if err := roleOdooUser(vers); err != nil {
		return fmt.Errorf("roleOdooUser %s failed %w", vers, err)
	}

	fmt.Println("roleOdooDirs:", vers)
	if err := roleOdooDirs(vers); err != nil {
		return fmt.Errorf("roleOdooDirs %s failed %w", vers, err)
	}

	fmt.Println("rolePostgresqlRepo:", vers)
	if err := rolePostgresqlRepo(vers); err != nil {
		return fmt.Errorf("rolePostgresqlRepo %s failed %w", vers, err)
	}

	fmt.Println("rolePostgresqlClient:", vers)
	if err := rolePostgresqlClient(vers); err != nil {
		return fmt.Errorf("rolePostgresqlClient %s failed %w", vers, err)
	}

	fmt.Println("roleWkhtmltopdf:", vers)
	if err := roleWkhtmltopdf(vers); err != nil {
		return fmt.Errorf("roleWkhtmltopdf %s failed %w", vers, err)
	}

	fmt.Println("roleOdooBasePackages:", vers)
	if err := roleOdooBasePackages(vers, version); err != nil {
		return fmt.Errorf("roleOdooBasePackages %s failed %w", vers, err)
	}

	fmt.Println("pip3Install:", vers)
	if err := pip3Install(vers, "ebaysdk", "google-auth"); err != nil {
		return fmt.Errorf("pip3Install %s failed %w", vers, err)
	}

	fmt.Println("npmInstall:", vers)
	if err := npmInstall(vers, "rtlcss"); err != nil {
		return fmt.Errorf("npmInstall %s failed %w", vers, err)
	}

	fmt.Println("roleGeoIP2DB:", vers)
	if err := roleGeoIP2DB(vers); err != nil {
		return fmt.Errorf("roleGeoIP2DB %s failed %w", vers, err)
	}

	fmt.Println("papersize:", vers)
	if err := IncusExec(vers, "/usr/sbin/paperconfig", "-p", "letter"); err != nil {
		return fmt.Errorf("papersize %s failed %w", vers, err)
	}

	fmt.Println("roleOdooNode:", vers)
	if err := roleOdooNode(vers); err != nil {
		return fmt.Errorf("roleOdooNode %s failed %w", vers, err)
	}

	fmt.Println("roleOdooService:", vers)
	if err := roleOdooService(vers); err != nil {
		return fmt.Errorf("roleOdooService %s failed %w", vers, err)
	}

	fmt.Println("roleCaddy")
	if err := roleCaddy(vers); err != nil {
		return fmt.Errorf("roleCaddy %s failed %w", vers, err)
	}
	fmt.Println("roleCaddyService")
	if err := roleCaddyService(vers); err != nil {
		return fmt.Errorf("roleCaddyService %s failed %w", vers, err)
	}

	fmt.Println("IncusStop:", vers)
	if err := IncusStop(vers); err != nil {
		return fmt.Errorf("IncusStop %s failed %w", vers, err)
	}

	return nil
}

func BaseDestroy(version string) error {
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)
	if err := IncusDelete(vers); err != nil {
		return fmt.Errorf("destroy container %s failed %w", vers, err)
	}
	fmt.Println("destroying:", vers)
	return nil
}

func aptInstall(name string, pkgs ...string) error {
	pkg := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(name, pkg...); err != nil {
		return fmt.Errorf("apt-get install failed %w", err)
	}
	return nil
}

func pip3Install(name string, pkgs ...string) error {
	pkg := []string{"pip3", "install"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(name, pkg...); err != nil {
		return fmt.Errorf("pip3 install failed %w", err)
	}
	return nil
}

func npmInstall(name string, pkgs ...string) error {
	pkg := []string{"npm", "install", "-g"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(name, pkg...); err != nil {
		return fmt.Errorf("npm install failed %w", err)
	}
	return nil
}

func roleUpdate(name string) error {
	// fmt.Println("apt-get update -y")
	if err := IncusExec(name, "apt-get", "update", "-y"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("apt-get update failed %w", err)
	}

	// fmt.Println("apt-get dist-upgrade -y")
	if err := IncusExec(name, "apt-get", "dist-upgrade", "-y"); err != nil {
		return fmt.Errorf("apt-get dist-upgrade failed %w", err)
	}

	// fmt.Println("apt-get autoremove -y")
	if err := IncusExec(name, "apt-get", "autoremove", "-y"); err != nil {
		return fmt.Errorf("apt-get autoremove failed %w", err)
	}

	// fmt.Println("apt-get autoclean -y")
	if err := IncusExec(name, "apt-get", "autoclean", "-y"); err != nil {
		return fmt.Errorf("apt-get autoclean failed %w", err)
	}
	fmt.Println("update complete")
	return nil
}

func roleBaseline(name string) error {
	if err := aptInstall(name,
		"sudo", "gnupg2", "curl", "wget",
		"apt-utils", "apt-transport-https",
		"git", "lsb-release", "vim",
		"openssh-server",
	); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	f, err := os.Create("/tmp/update.sh")
	if err != nil {
		return fmt.Errorf("create update.sh failed %w", err)
	}
	f.WriteString("#!/bin/bash" + "\n")
	f.WriteString("sudo bash -c \"apt update -y && apt full-upgrade -y && apt autoremove -y && apt autoclean -y\"" + "\n")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/update.sh", name+"/usr/local/bin/update").Run(); err != nil {
		return fmt.Errorf("push update.sh failed %w", err)
	}

	os.Remove("/tmp/update.sh")

	if err := IncusExec(name, "chmod", "+x", "/usr/local/bin/update"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod update failed %w", err)
	}

	return nil
}

func roleOdooUser(name string) error {
	if err := IncusExec(name, "groupadd", "-f", "-g", "1001", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("groupadd odoo failed %w", err)
	}

	if err := IncusExec(name, "useradd", "-ms", "/bin/bash", "-g", "1001", "-u", "1001", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("useradd odoo failed %w", err)
	}

	if err := IncusExec(name, "usermod", "-aG", "sudo", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("usermod odoo failed %w", err)
	}

	f, err := os.Create("/tmp/odoo.sudo")
	if err != nil {
		return fmt.Errorf("create odoo.sudo failed %w", err)
	}
	f.WriteString("odoo ALL=(ALL) NOPASSWD:ALL")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.sudo", name+"/etc/sudoers.d/odoo").Run(); err != nil {
		return fmt.Errorf("push odoo.sudo failed %w", err)
	}

	if err := IncusExec(name, "chown", "root:root", "/etc/sudoers.d/odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	os.Remove("/tmp/odoo.sudo")

	// SSH key
	if err := IncusExec(name, "mkdir", "/home/odoo/.ssh"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("mkdir /home/odoo/.ssh failed %w", err)
	}

	conf := GetConf()
	HOME, _ := os.UserHomeDir()
	sshKey := HOME + "/.ssh/" + conf.SSHKey + ".pub"
	fmt.Println("SSHKey:", sshKey)

	if err := exec.Command("incus", "file", "push", sshKey, name+"/home/odoo/.ssh/authorized_keys").Run(); err != nil {
		return fmt.Errorf("push authorized_keys failed %w", err)
	}

	if err := IncusExec(name, "chmod", "0600", "/home/odoo/.ssh/authorized_keys"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod authorized_keys failed %w", err)
	}

	if err := IncusExec(name, "chown", "-R", "odoo:odoo", "/home/odoo/.ssh"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooDirs(name string) error {
	dirs := []string{"addons", "backups", "conf", "data", "odoo", "enterprise"}

	for _, dir := range dirs {
		if err := IncusExec(name, "mkdir", "-p", "/opt/odoo/"+dir); err != nil {
			fmt.Println(err)
			return fmt.Errorf("mkdir %s failed %w", dir, err)
		}
	}

	if err := IncusExec(name, "chown", "-R", "odoo:odoo", "/opt/odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooBasePackages(name, version string) error {
	packages := []string{}
	switch strings.Split(version, ".")[0] {
	case "15", "16", "17":
		packages = []string{
			"bzip2", "ca-certificates", "curl", "dirmngr", "fonts-liberation",
			"fonts-noto", "fonts-noto-cjk", "fonts-noto-mono", "geoip-database",
			"gnupg", "gsfonts", "inetutils-ping", "libgnutls-dane0", "libgts-bin",
			"libpaper-utils", "locales", "nodejs", "npm", "python3", "python3-babel",
			"python3-chardet", "python3-cryptography", "python3-cups",
			"python3-dateutil", "python3-decorator", "python3-docutils",
			"python3-feedparser", "python3-freezegun", "python3-full", "python3-geoip2",
			"python3-gevent", "python3-greenlet", "python3-html2text", "python3-idna",
			"python3-jinja2", "python3-ldap", "python3-libsass", "python3-lxml",
			"python3-markupsafe", "python3-num2words", "python3-odf",
			"python3-ofxparse", "python3-olefile", "python3-openssl",
			"python3-paramiko", "python3-googleapi", "python3-passlib",
			"python3-pdfminer", "python3-phonenumbers", "python3-pil", "python3-pip",
			"python3-polib", "python3-psutil", "python3-psycopg2", "python3-pydot",
			"python3-pylibdmtx", "python3-pyparsing", "python3-pypdf2",
			"python3-qrcode", "python3-renderpm", "python3-reportlab",
			"python3-reportlab-accel", "python3-requests", "python3-rjsmin",
			"python3-serial", "python3-setuptools", "python3-stdnum", "python3-tz",
			"python3-urllib3", "python3-usb", "python3-vobject", "python3-werkzeug",
			"python3-xlrd", "python3-xlsxwriter", "python3-xlwt", "python3-zeep",
			"shared-mime-info", "unzip", "xz-utils", "zip", "zstd",
		}
	}
	if err := aptInstall(name, packages...); err != nil {
		return fmt.Errorf("apt-get install odoo base packages failed %w", err)
	}
	return nil
}

func roleOdooNode(name string) error {
	if err := IncusExec(name, "wget", "-qO", "/usr/local/bin/oda", "https://raw.githubusercontent.com/ppreeper/oda/main/oda.py"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget oda failed %w", err)
	}
	if err := IncusExec(name, "chmod", "+x", "/usr/local/bin/oda"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod oda failed %w", err)
	}
	return nil
}

func roleOdooService(name string) error {
	f, err := os.Create("/tmp/odoo.service")
	if err != nil {
		return fmt.Errorf("create odoo.service failed %w", err)
	}
	f.WriteString("[Unit]" + "\n")
	f.WriteString("Description=Odoo" + "\n")
	f.WriteString("Requires=network-online.target" + "\n")
	f.WriteString("After=remote-fs.target" + "\n" + "\n")
	f.WriteString("[Service]" + "\n")
	f.WriteString("Type=simple" + "\n")
	f.WriteString("SyslogIdentifier=odoo" + "\n")
	f.WriteString("PermissionsStartOnly=true" + "\n")
	f.WriteString("User=odoo" + "\n")
	f.WriteString("Group=odoo" + "\n")
	f.WriteString("ExecStart=/opt/odoo/odoo/odoo-bin -c /opt/odoo/conf/odoo.conf" + "\n")
	f.WriteString("StandardOutput=journal+console" + "\n")
	f.WriteString("Restart=on-failure" + "\n")
	f.WriteString("RestartSec=10s" + "\n" + "\n")
	f.WriteString("[Install]" + "\n")
	f.WriteString("WantedBy=remote-fs.target" + "\n")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.service", name+"/etc/systemd/system/odoo.service").Run(); err != nil {
		return fmt.Errorf("push odoo.service failed %w", err)
	}

	os.Remove("/tmp/odoo.service")

	if err := IncusExec(name, "systemctl", "daemon-reload"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(name, "systemctl", "enable", "odoo.service"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl enable odoo.service failed %w", err)
	}

	return nil
}

func roleGeoIP2DB(name string) error {
	// # install geolite databases
	geolite := [][]string{
		{"GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"},
		{"GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"},
		{"GeoLite2-Country.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"},
	}

	for _, geo := range geolite {
		fmt.Println("downloading", geo[0])
		if err := IncusExec(name, "wget", "-qO", "/usr/share/GeoIP/"+geo[0], geo[1]); err != nil {
			fmt.Println(err)
			return fmt.Errorf("wget %s failed %w", geo[0], err)
		}
	}

	return nil
}

func rolePostgresqlRepo(name string) error {
	if err := IncusExec(name, "wget", "-qO", "/etc/apt/trusted.gpg.d/pgdg.gpg.asc", "https://www.postgresql.org/media/keys/ACCC4CF8.asc"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget pgdg.gpg.asc failed %w", err)
	}
	f, err := os.Create("/tmp/pgdg.list")
	if err != nil {
		return fmt.Errorf("create pgdg.list failed %w", err)
	}
	f.WriteString("deb [arch=amd64] http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/pgdg.list", name+"/etc/apt/sources.list.d/pgdg.list").Run(); err != nil {
		return fmt.Errorf("push pgdg.list failed %w", err)
	}

	os.Remove("/tmp/pgdg.list")

	if err := roleUpdate(name); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", name, err)
	}

	return nil
}

func rolePostgresqlClient(name string) error {
	if err := IncusExec(name, "apt-get", "install", "-y", "--no-install-recommends", "postgresql-client-15"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("apt-get install postgresql-client-15 failed %w", err)
	}
	return nil
}

func roleWkhtmltopdf(name string) error {
	// wkhtmltopdf
	url := "https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.jammy_amd64.deb"

	if err := IncusExec(name, "wget", "-qO", "wkhtmltox.deb", url); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(name, "apt-get", "install", "-y", "--no-install-recommends", "./wkhtmltox.deb"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("apt-get install wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(name, "rm", "-rf", "wkhtmltox.deb"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("rm wkhtmltox.deb failed %w", err)
	}
	return nil
}

func roleCaddy(name string) error {
	url := "https://caddyserver.com/api/download?os=linux&arch=amd64&p=github.com%2Fcaddy-dns%2Fcloudflare"

	if err := IncusExec(name, "wget", "-qO", "/usr/local/bin/caddy", url); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget caddy failed %w", err)
	}
	if err := IncusExec(name, "chmod", "+x", "/usr/local/bin/caddy"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}

	return nil
}

func roleCaddyService(name string) error {
	f, err := os.Create("/tmp/caddy.service")
	if err != nil {
		return fmt.Errorf("create caddy.service failed %w", err)
	}
	f.WriteString("[Unit]" + "\n")
	f.WriteString("Description=caddy server" + "\n")
	f.WriteString("Requires=network-online.target" + "\n")
	f.WriteString("After=remote-fs.target" + "\n" + "\n")

	f.WriteString("[Service]" + "\n")
	f.WriteString("Type=simple" + "\n")
	f.WriteString("SyslogIdentifier=caddy" + "\n")
	f.WriteString("Restart=on-failure" + "\n")
	f.WriteString("ExecStart=/usr/local/bin/caddy run --config /etc/caddy/Caddyfile" + "\n")
	f.WriteString("KillSignal=SIGTERM" + "\n")
	f.WriteString("StandardOutput=journal+console" + "\n")
	f.WriteString("Restart=on-failure" + "\n")
	f.WriteString("RestartSec=10s" + "\n" + "\n")

	f.WriteString("[Install]" + "\n")
	f.WriteString("WantedBy=multi-user.target" + "\n")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/caddy.service", name+"/etc/systemd/system/caddy.service").Run(); err != nil {
		return fmt.Errorf("push caddy.service failed %w", err)
	}

	os.Remove("/tmp/caddy.service")

	if err := IncusExec(name, "systemctl", "daemon-reload"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(name, "systemctl", "enable", "caddy.service"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl enable caddy.service failed %w", err)
	}

	if err := IncusExec(name, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("mkdir /etc/caddy failed %w", err)
	}

	return nil
}
