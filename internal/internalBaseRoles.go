package internal

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func aptInstall(instanceName string, pkgs ...string) error {
	// fmt.Fprintln(os.Stderr, "apt-get install packages to", instanceName)
	pkg := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("apt-get install failed %w", err)
	}
	return nil
}

func npmInstall(instanceName string, pkgs ...string) error {
	// fmt.Fprintln(os.Stderr, "npm install packages to", instanceName)
	pkg := []string{"npm", "install", "-g"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("npm install failed %w", err)
	}
	return nil
}

func roleCaddy(instanceName string) error {
	fmt.Fprintln(os.Stderr, "install caddy to", instanceName)
	url := "https://caddyserver.com/api/download?os=linux&arch=amd64&p=github.com%2Fcaddy-dns%2Fcloudflare"

	if err := IncusExec(instanceName, "wget", "-qO", "/usr/local/bin/caddy", url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("wget caddy failed %w", err)
	}
	if err := IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}
	if err := IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}

	return nil
}

func roleCaddyService(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add caddy.service systemd file to", instanceName)
	fo, err := os.Create("/tmp/caddy.service")
	if err != nil {
		return fmt.Errorf("create caddy.service failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/caddy.service")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)

	if err := exec.Command("incus", "file", "push", "/tmp/caddy.service", instanceName+"/etc/systemd/system/caddy.service").Run(); err != nil {
		return fmt.Errorf("push caddy.service failed %w", err)
	}

	os.Remove("/tmp/caddy.service")

	if err := IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(instanceName, "systemctl", "enable", "caddy.service"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl enable caddy.service failed %w", err)
	}

	if err := IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("mkdir /etc/caddy failed %w", err)
	}

	return nil
}

func rolePostgresqlRepo(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add postgresql repo to", instanceName)

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	if err := aptInstall(instanceName, "postgresql-common", "apt-transport-https", "ca-certificates"); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	if err := IncusExec(instanceName, "/usr/share/postgresql-common/pgdg/apt.postgresql.org.sh", "-y"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt.postgresql.org.sh failed %w", err)
	}

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	return nil
}

func rolePostgresqlClient(instanceName string, dbVersion string) error {
	fmt.Fprintln(os.Stderr, "add postgresql-client to", instanceName)
	if err := IncusExec(instanceName, "apt-get", "install", "-y", "--no-install-recommends", "postgresql-client-"+dbVersion); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt-get install postgresql-client-%s failed %w", dbVersion, err)
	}
	return nil
}

func rolePreeperRepo(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add preeper.org repo:", instanceName)

	fo, err := os.Create("/tmp/preeper.list")
	if err != nil {
		return fmt.Errorf("create preeper.list failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/preeper.list")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)

	if err := exec.Command("incus", "file", "push", "/tmp/preeper.list", instanceName+"/etc/apt/sources.list.d/preeper.list").Run(); err != nil {
		return fmt.Errorf("push preeper.list failed %w", err)
	}

	os.Remove("/tmp//preeper.list")

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	return nil
}

func roleWkhtmltopdf(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add wkthmltox to", instanceName)
	// wkhtmltopdf
	url := "https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.jammy_amd64.deb"

	if err := IncusExec(instanceName, "wget", "-qO", "wkhtmltox.deb", url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("wget wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(instanceName, "apt-get", "install", "-y", "--no-install-recommends", "./wkhtmltox.deb"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt-get install wkhtmltox.deb failed %w", err)
	}

	if err := IncusExec(instanceName, "rm", "-rf", "wkhtmltox.deb"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("rm wkhtmltox.deb failed %w", err)
	}
	return nil
}

func roleOdooService(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add odoo.service systemd file to", instanceName)
	fo, err := os.Create("/tmp/odoo.service")
	if err != nil {
		return fmt.Errorf("create odoo.service failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/odoo.service")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.service", instanceName+"/etc/systemd/system/odoo.service").Run(); err != nil {
		return fmt.Errorf("push odoo.service failed %w", err)
	}

	os.Remove("/tmp/odoo.service")

	if err := IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := IncusExec(instanceName, "systemctl", "enable", "odoo.service"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl enable odoo.service failed %w", err)
	}

	return nil
}

func roleGeoIP2DB(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add geoip2 databases to", instanceName)
	// install geolite databases
	geolite := [][]string{
		{"GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"},
		{"GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"},
		{"GeoLite2-Country.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"},
	}

	for _, geo := range geolite {
		fmt.Fprintln(os.Stderr, "downloading", geo[0])
		if err := IncusExec(instanceName, "wget", "-qO", "/usr/share/GeoIP/"+geo[0], geo[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return fmt.Errorf("wget %s failed %w", geo[0], err)
		}
	}

	return nil
}

func rolePaperSize(instanceName string) error {
	fmt.Fprintln(os.Stderr, "papersize:", instanceName)
	if err := IncusExec(instanceName, "/usr/sbin/paperconfig", "-p", "letter"); err != nil {
		return fmt.Errorf("papersize %s failed %w", instanceName, err)
	}

	return nil
}

func roleOdooUser(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add odoo user to", instanceName)
	if err := IncusExec(instanceName, "groupadd", "-f", "-g", "1001", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("groupadd odoo failed %w", err)
	}

	if err := IncusExec(instanceName, "useradd", "-ms", "/bin/bash", "-g", "1001", "-u", "1001", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("useradd odoo failed %w", err)
	}

	if err := IncusExec(instanceName, "usermod", "-aG", "sudo", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	os.Remove("/tmp/odoo.sudo")

	// SSH key
	if err := IncusExec(instanceName, "mkdir", "/home/odoo/.ssh"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("mkdir /home/odoo/.ssh failed %w", err)
	}

	sshKey := viper.GetString("system.sshkey") + ".pub"
	fmt.Fprintln(os.Stderr, "SSHKey:", sshKey)

	if err := exec.Command("incus", "file", "push", sshKey, instanceName+"/home/odoo/.ssh/authorized_keys").Run(); err != nil {
		return fmt.Errorf("push authorized_keys failed %w", err)
	}

	if err := IncusExec(instanceName, "chmod", "0600", "/home/odoo/.ssh/authorized_keys"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod authorized_keys failed %w", err)
	}

	if err := IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/home/odoo/.ssh"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooDirs(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add odoo app directories to", instanceName)
	dirs := []string{"addons", "backups", "conf", "data", "odoo", "enterprise", "design-themes", "industry"}

	for _, dir := range dirs {
		if err := IncusExec(instanceName, "mkdir", "-p", "/opt/odoo/"+dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return fmt.Errorf("mkdir %s failed %w", dir, err)
		}
	}

	if err := IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/opt/odoo"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleUpdate(instanceName string) error {
	fmt.Fprintln(os.Stderr, "update packages", instanceName)
	if err := IncusExec(instanceName, "update"); err != nil {
		return fmt.Errorf("roleUpdate: update failed %w", err)
	}

	fmt.Fprintln(os.Stderr, "update complete")
	return nil
}

func roleUpdateScript(instanceName string) error {
	fmt.Fprintln(os.Stderr, "add update script to", instanceName)
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
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod update failed %w", err)
	}
	return nil
}

// func roleGupScript(instanceName string) error {
// 	gup := `https://raw.githubusercontent.com/ppreeper/gup/main/gup`
// 	if err := IncusExec(instanceName, "wget", "-qO", "/usr/local/bin/gup", gup); err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return fmt.Errorf("wget gup failed %w", err)
// 	}
// 	if err := IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/gup"); err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return fmt.Errorf("chmod gup failed %w", err)
// 	}
// 	return nil
// }
