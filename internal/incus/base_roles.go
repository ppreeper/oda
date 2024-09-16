package oda

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func roleCaddy(instanceName string) error {
	fmt.Println("roleCaddy")
	url := "https://caddyserver.com/api/download?os=linux&arch=amd64&p=github.com%2Fcaddy-dns%2Fcloudflare"

	if err := IncusExec(instanceName, "wget", "-qO", "/usr/local/bin/caddy", url); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget caddy failed %w", err)
	}
	if err := IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/caddy"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}
	if err := IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}

	return nil
}

func roleCaddyService(instanceName string) error {
	fmt.Println("roleCaddyService:", instanceName)
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

	if err := exec.Command("incus", "file", "push", "/tmp/caddy.service", instanceName+"/etc/systemd/system/caddy.service").Run(); err != nil {
		return fmt.Errorf("push caddy.service failed %w", err)
	}

	os.Remove("/tmp/caddy.service")

	if err := IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(instanceName, "systemctl", "enable", "caddy.service"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl enable caddy.service failed %w", err)
	}

	if err := IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("mkdir /etc/caddy failed %w", err)
	}

	return nil
}

func rolePostgresqlRepo(instanceName string) error {
	fmt.Println("rolePostgresqlRepo:", instanceName)
	if err := IncusExec(instanceName, "wget", "-qO", "/etc/apt/trusted.gpg.d/pgdg.gpg.asc", "https://www.postgresql.org/media/keys/ACCC4CF8.asc"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget pgdg.gpg.asc failed %w", err)
	}
	f, err := os.Create("/tmp/pgdg.list")
	if err != nil {
		return fmt.Errorf("create pgdg.list failed %w", err)
	}
	f.WriteString("deb [arch=amd64] http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/pgdg.list", instanceName+"/etc/apt/sources.list.d/pgdg.list").Run(); err != nil {
		return fmt.Errorf("push pgdg.list failed %w", err)
	}

	os.Remove("/tmp/pgdg.list")

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	return nil
}

func rolePostgresqlClient(instanceName string) error {
	fmt.Println("rolePostgresqlClient:", instanceName)
	conf := GetConf()
	if err := IncusExec(instanceName, "apt-get", "install", "-y", "--no-install-recommends", "postgresql-client-"+conf.DBVersion); err != nil {
		fmt.Println(err)
		return fmt.Errorf("apt-get install postgresql-client-%s failed %w", conf.DBVersion, err)
	}
	return nil
}

func roleWkhtmltopdf(instanceName string) error {
	fmt.Println("roleWkhtmltopdf:", instanceName)
	// wkhtmltopdf
	url := "https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.jammy_amd64.deb"

	if err := IncusExec(instanceName, "wget", "-qO", "wkhtmltox.deb", url); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(instanceName, "apt-get", "install", "-y", "--no-install-recommends", "./wkhtmltox.deb"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("apt-get install wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(instanceName, "rm", "-rf", "wkhtmltox.deb"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("rm wkhtmltox.deb failed %w", err)
	}
	return nil
}

func roleOdooService(instanceName string) error {
	fmt.Println("roleOdooService:", instanceName)
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

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.service", instanceName+"/etc/systemd/system/odoo.service").Run(); err != nil {
		return fmt.Errorf("push odoo.service failed %w", err)
	}

	os.Remove("/tmp/odoo.service")

	if err := IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(instanceName, "systemctl", "enable", "odoo.service"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("systemctl enable odoo.service failed %w", err)
	}

	return nil
}

func roleGeoIP2DB(instanceName string) error {
	fmt.Println("roleGeoIP2DB:", instanceName)
	// install geolite databases
	geolite := [][]string{
		{"GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"},
		{"GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"},
		{"GeoLite2-Country.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"},
	}

	for _, geo := range geolite {
		fmt.Println("downloading", geo[0])
		if err := IncusExec(instanceName, "wget", "-qO", "/usr/share/GeoIP/"+geo[0], geo[1]); err != nil {
			fmt.Println(err)
			return fmt.Errorf("wget %s failed %w", geo[0], err)
		}
	}

	return nil
}

func rolePaperSize(instanceName string) error {
	fmt.Println("papersize:", instanceName)
	if err := IncusExec(instanceName, "/usr/sbin/paperconfig", "-p", "letter"); err != nil {
		return fmt.Errorf("papersize %s failed %w", instanceName, err)
	}

	return nil
}

func roleOdooUser(instanceName string) error {
	fmt.Println("roleOdooUser:", instanceName)
	if err := IncusExec(instanceName, "groupadd", "-f", "-g", "1001", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("groupadd odoo failed %w", err)
	}

	if err := IncusExec(instanceName, "useradd", "-ms", "/bin/bash", "-g", "1001", "-u", "1001", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("useradd odoo failed %w", err)
	}

	if err := IncusExec(instanceName, "usermod", "-aG", "sudo", "odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("usermod odoo failed %w", err)
	}

	f, err := os.Create("/tmp/odoo.sudo")
	if err != nil {
		return fmt.Errorf("create odoo.sudo failed %w", err)
	}
	f.WriteString("odoo ALL=(ALL) NOPASSWD:ALL")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.sudo", instanceName+"/etc/sudoers.d/odoo").Run(); err != nil {
		return fmt.Errorf("push odoo.sudo failed %w", err)
	}

	if err := IncusExec(instanceName, "chown", "root:root", "/etc/sudoers.d/odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	os.Remove("/tmp/odoo.sudo")

	// SSH key
	if err := IncusExec(instanceName, "mkdir", "/home/odoo/.ssh"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("mkdir /home/odoo/.ssh failed %w", err)
	}

	conf := GetConf()
	HOME, _ := os.UserHomeDir()
	sshKey := HOME + "/.ssh/" + conf.SSHKey + ".pub"
	fmt.Println("SSHKey:", sshKey)

	if err := exec.Command("incus", "file", "push", sshKey, instanceName+"/home/odoo/.ssh/authorized_keys").Run(); err != nil {
		return fmt.Errorf("push authorized_keys failed %w", err)
	}

	if err := IncusExec(instanceName, "chmod", "0600", "/home/odoo/.ssh/authorized_keys"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod authorized_keys failed %w", err)
	}

	if err := IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/home/odoo/.ssh"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooDirs(instanceName string) error {
	fmt.Println("roleOdooDirs:", instanceName)
	dirs := []string{"addons", "backups", "conf", "data", "odoo", "enterprise"}

	for _, dir := range dirs {
		if err := IncusExec(instanceName, "mkdir", "-p", "/opt/odoo/"+dir); err != nil {
			fmt.Println(err)
			return fmt.Errorf("mkdir %s failed %w", dir, err)
		}
	}

	if err := IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/opt/odoo"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooBasePackages(instanceName, version string) error {
	fmt.Println("roleOdooBasePackages:", instanceName)
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
	if err := aptInstall(instanceName, packages...); err != nil {
		return fmt.Errorf("apt-get install odoo base packages failed %w", err)
	}
	return nil
}

func roleOdooNode(instanceName string) error {
	fmt.Println("roleOdooNode:", instanceName)
	if err := IncusExec(instanceName, "wget", "-qO", "/usr/local/bin/oda", "https://raw.githubusercontent.com/ppreeper/oda/main/oda.py"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("wget oda failed %w", err)
	}
	if err := IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/oda"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod oda failed %w", err)
	}
	return nil
}

func roleBaseline(instanceName string) error {
	fmt.Println("roleBaseline:", instanceName)
	if err := aptInstall(instanceName,
		"sudo", "gnupg2", "curl", "wget",
		"apt-utils", "apt-transport-https",
		"git", "lsb-release", "vim",
		"openssh-server",
	); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	return nil
}

func roleUpdate(instanceName string) error {
	fmt.Println("roleUpdate:", instanceName)
	if err := IncusExec(instanceName, "update"); err != nil {
		return fmt.Errorf("roleUpdate: update failed %w", err)
	}

	fmt.Println("update complete")
	return nil
}

func roleUpdateScript(instanceName string) error {
	f, err := os.Create("/tmp/update.sh")
	if err != nil {
		return fmt.Errorf("create update.sh failed %w", err)
	}
	f.WriteString("#!/bin/bash" + "\n")
	f.WriteString("sudo bash -c \"apt update -y && apt full-upgrade -y && apt autoremove -y && apt autoclean -y\"" + "\n")
	f.Close()

	if err := exec.Command("incus", "file", "push", "/tmp/update.sh", instanceName+"/usr/local/bin/update").Run(); err != nil {
		return fmt.Errorf("push update.sh failed %w", err)
	}

	os.Remove("/tmp/update.sh")

	if err := IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/update"); err != nil {
		fmt.Println(err)
		return fmt.Errorf("chmod update failed %w", err)
	}
	return nil
}
