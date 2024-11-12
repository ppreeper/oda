package internal

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

func aptInstall(instanceName string, pkgs ...string) error {
	fmt.Fprintln(os.Stderr,
		ui.SubStepStyle.Render("apt-get", "install", "-y", "--no-install-recommends"),
		ui.OddRowStyle.Render(pkgs...),
	)
	pkg := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	pkg = append(pkg, pkgs...)
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("apt-get install failed %w", err)
	}
	return nil
}

func npmInstall(instanceName string, pkgs ...string) error {
	fmt.Fprintln(os.Stderr,
		ui.SubStepStyle.Render("npm", "install", "-g"),
		ui.OddRowStyle.Render(pkgs...),
	)
	pkg := []string{"npm", "install", "-g"}
	pkg = append(pkg, pkgs...)

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("npm install failed %w", err)
	}
	return nil
}

func roleCaddy(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("install caddy to", instanceName))
	url := "https://caddyserver.com/api/download?os=linux&arch=amd64&p=github.com%2Fcaddy-dns%2Fcloudflare"

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "wget", "-qO", "/usr/local/bin/caddy", url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("wget caddy failed %w", err)
	}
	if err := inc.IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}
	if err := inc.IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("chmod caddy failed %w", err)
	}

	return nil
}

func roleCaddyService(instanceName string, embedFS embed.FS) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add caddy.service systemd file to", instanceName))
	fo, err := os.Create("/tmp/caddy.service")
	if err != nil {
		return fmt.Errorf("create caddy.service failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/caddy.service")
	lib.CheckErr(err)
	err = t.Execute(fo, data)
	lib.CheckErr(err)

	if err := exec.Command("incus", "file", "push", "/tmp/caddy.service", instanceName+"/etc/systemd/system/caddy.service").Run(); err != nil {
		return fmt.Errorf("push caddy.service failed %w", err)
	}

	os.Remove("/tmp/caddy.service")

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "systemctl", "enable", "caddy.service"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl enable caddy.service failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("mkdir /etc/caddy failed %w", err)
	}

	return nil
}

func rolePostgresqlRepo(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add postgresql repo to", instanceName))

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	if err := aptInstall(instanceName, "postgresql-common", "apt-transport-https", "ca-certificates"); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "/usr/share/postgresql-common/pgdg/apt.postgresql.org.sh", "-y"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt.postgresql.org.sh failed %w", err)
	}

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	return nil
}

func rolePostgresqlClient(instanceName string, dbVersion string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add postgresql-client to", instanceName))

	if err := aptInstall(instanceName, "postgresql-client-"+dbVersion); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt-get install postgresql-client-%s failed %w", dbVersion, err)
	}
	return nil
}

func rolePostgresqlServer(instanceName string, dbVersion string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add postgresql-"+dbVersion, "to", instanceName))

	if err := aptInstall(instanceName, "postgresql-"+dbVersion, "postgresql-contrib", "pgtop"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("apt-get install postgresql-%s failed %w", dbVersion, err)
	}
	return nil
}

func rolePostgresqlConf(instanceName string, embedFS embed.FS) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add postgresql.conf", instanceName))

	odaConf, _ := config.LoadOdaConfig()
	dbVersion := fmt.Sprintf("%d", odaConf.Database.Version)

	localFile := "/tmp/" + instanceName + "-postgresql.conf"

	fo, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create postgresql.conf: %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/postgresql.conf")
	lib.CheckErr(err)
	err = t.Execute(fo, data)
	lib.CheckErr(err)

	if err := exec.Command("incus", "file", "push", localFile, instanceName+"/etc/postgresql/"+dbVersion+"/main/conf.d/postgresql.conf").Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("failed to push postgresql.conf: %w", err)
	}

	os.Remove(localFile)

	return nil
}

func rolePghbaConf(instanceName string, embedFS embed.FS) error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}

	dbVersion := fmt.Sprintf("%d", odaConf.Database.Version)

	localFile := "/tmp/" + instanceName + "-pg_hba.conf"

	fo, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create pg_hba.conf: %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/pg_hba.conf")
	lib.CheckErr(err)
	err = t.Execute(fo, data)
	lib.CheckErr(err)

	if err := exec.Command("incus", "file", "push", localFile, instanceName+"/etc/postgresql/"+dbVersion+"/main/pg_hba.conf").Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("failed to push pg_hba.conf: %w", err)
	}

	os.Remove(localFile)

	return nil
}

func rolePreeperRepo(instanceName string, embedFS embed.FS) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Add preeper.org repo:", instanceName))

	fo, err := os.Create("/tmp/preeper.list")
	if err != nil {
		return fmt.Errorf("create preeper.list failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/preeper.list")
	if err != nil {
		return fmt.Errorf("parse preeper.list failed %w", err)
	}
	err = t.Execute(fo, data)
	if err != nil {
		return fmt.Errorf("execute preeper.list failed %w", err)
	}

	if err := exec.Command("incus", "file", "push", "/tmp/preeper.list", instanceName+"/etc/apt/sources.list.d/preeper.list").Run(); err != nil {
		return fmt.Errorf("push preeper.list failed %w", err)
	}

	os.Remove("/tmp//preeper.list")

	if err := roleUpdate(instanceName); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", instanceName, err)
	}

	aptInstall(instanceName, []string{"odaserver"}...)

	return nil
}

func roleWkhtmltopdf(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add wkthmltox to", instanceName))
	// wkhtmltopdf
	url := "https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.jammy_amd64.deb"

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "wget", "-qO", "wkhtmltox.deb", url); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("wget wkhtmltox.deb failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "apt-get", "install", "-y", "--no-install-recommends", "./wkhtmltox.deb"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("apt-get install wkhtmltox.deb failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "rm", "-rf", "wkhtmltox.deb"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("rm wkhtmltox.deb failed %w", err)
	}
	return nil
}

func roleOdooService(instanceName string, embedFS embed.FS) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add odoo.service systemd file to", instanceName))
	fo, err := os.Create("/tmp/odoo.service")
	if err != nil {
		return fmt.Errorf("create odoo.service failed %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/odoo.service")
	lib.CheckErr(err)
	err = t.Execute(fo, data)
	lib.CheckErr(err)

	if err := exec.Command("incus", "file", "push", "/tmp/odoo.service", instanceName+"/etc/systemd/system/odoo.service").Run(); err != nil {
		return fmt.Errorf("push odoo.service failed %w", err)
	}

	os.Remove("/tmp/odoo.service")

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "systemctl", "daemon-reload"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl daemon-reload failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "systemctl", "enable", "odoo.service"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("systemctl enable odoo.service failed %w", err)
	}

	return nil
}

func roleGeoIP2DB(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add geoip2 databases to", instanceName))
	// install geolite databases
	geolite := [][]string{
		{"GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"},
		{"GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"},
		{"GeoLite2-Country.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"},
	}

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	for _, geo := range geolite {
		fmt.Fprintln(os.Stderr, ui.OddRowStyle.Render("downloading", geo[0]))
		if err := inc.IncusExec(instanceName, "wget", "-qO", "/usr/share/GeoIP/"+geo[0], geo[1]); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
			return fmt.Errorf("wget %s failed %w", geo[0], err)
		}
	}

	return nil
}

func rolePaperSize(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("papersize:", instanceName))

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "/usr/sbin/paperconfig", "-p", "letter"); err != nil {
		return fmt.Errorf("papersize %s failed %w", instanceName, err)
	}

	return nil
}

func roleOdooUser(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add odoo user to", instanceName))
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "groupadd", "-f", "-g", "1001", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("groupadd odoo failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "useradd", "-ms", "/bin/bash", "-g", "1001", "-u", "1001", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("useradd odoo failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "usermod", "-aG", "sudo", "odoo"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
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

	if err := inc.IncusExec(instanceName, "chown", "root:root", "/etc/sudoers.d/odoo"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("chown odoo failed %w", err)
	}

	os.Remove("/tmp/odoo.sudo")

	// SSH key
	if err := inc.IncusExec(instanceName, "mkdir", "/home/odoo/.ssh"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("mkdir /home/odoo/.ssh failed %w", err)
	}

	sshKey := odaConf.System.SSHKey
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("SSHKey:", sshKey))
	homedir, _ := os.UserHomeDir()
	sshKeyFile := filepath.Join(homedir, ".ssh", sshKey+".pub")

	if err := exec.Command("incus", "file", "push", sshKeyFile, instanceName+"/home/odoo/.ssh/authorized_keys").Run(); err != nil {
		return fmt.Errorf("push authorized_keys failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "chmod", "0600", "/home/odoo/.ssh/authorized_keys"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("chmod authorized_keys failed %w", err)
	}

	if err := inc.IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/home/odoo/.ssh"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleOdooDirs(instanceName string, dirs []string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add odoo app directories to", instanceName))
	dirList := []string{"addons", "backups", "conf", "data"}
	dirList = append(dirList, dirs...)

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	for _, dir := range dirList {
		if err := inc.IncusExec(instanceName, "mkdir", "-p", "/opt/odoo/"+dir); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
			return fmt.Errorf("mkdir %s failed %w", dir, err)
		}
	}

	if err := inc.IncusExec(instanceName, "chown", "-R", "odoo:odoo", "/opt/odoo"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		return fmt.Errorf("chown odoo failed %w", err)
	}

	return nil
}

func roleUpdate(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("update packages", instanceName))

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExec(instanceName, "update"); err != nil {
		return fmt.Errorf("roleUpdate: update failed %w", err)
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("update complete"))
	return nil
}

func roleUpdateScript(instanceName string) error {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Add update script to", instanceName))

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

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

	if err := inc.IncusExec(instanceName, "chmod", "+x", "/usr/local/bin/update"); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
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
