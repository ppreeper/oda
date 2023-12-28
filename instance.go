package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/crypto/pbkdf2"
)

func instanceStart() error {
	fmt.Println("Start the instance")
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	cwd, project := GetProject()
	dirs := GetDirs()
	version := GetVersion()
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
	proxyGenerate()
	proxyStop()
	proxyStart()
	return nil
}

func instanceStop() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()
	if err := exec.Command("podman", "stop", project+".local").Run(); err != nil {
		return err
	}
	fmt.Println(project+".local", "stopped")
	return nil
}

func instanceRestart() error {
	if err := instanceStop(); err != nil {
		return err
	}
	if err := instanceStart(); err != nil {
		return err
	}
	return nil
}

func instanceAppInstallUpgrade(install bool, modules ...string) error {
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

func instanceScaffold(module string) error {
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

func instancePS() error {
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

func instanceExec() error {
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

func instanceLogs() error {
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

func instancePSQL() error {
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

func adminUsername() error {
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
		"-c", "update res_users set login='"+user1+"' where id=2;",
	)
	if err := podCmd.Run(); err != nil {
		return err
	}
	fmt.Println("Admin username changed to", user1)
	return nil
}

func adminPassword() error {
	fmt.Println("admin password")
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
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

	fmt.Println("password1", password1, "password2", password2)

	salt := []byte(randStr(22))
	iteration := 350_000
	key := pbkdf2.Key([]byte(password1), salt, iteration, sha512.Size, sha512.New)
	pass := base64.RawStdEncoding.EncodeToString(key)

	password := fmt.Sprintf("\\$pbkdf2-sha512\\$%d\\$%s\\$%s", iteration, string(salt), pass)
	fmt.Println("password", password)

	cwd, project := GetProject()

	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbname := GetOdooConf(cwd, "db_name")
	fmt.Println("podman",
		"exec", "-it",
		"-e", "PGPASSWORD="+dbpassword,
		project+".local",
		"psql", "-h", dbhost,
		"-U", dbuser,
		dbname,
		"-c", "\"update res_users set password='"+strings.TrimSpace(password)+"' where id=2;\"")

	podCmd := exec.Command("podman",
		"exec", "-it",
		"-e", "PGPASSWORD="+dbpassword,
		project+".local",
		"psql", "-h", dbhost,
		"-U", dbuser,
		dbname,
		"-c", "\"update res_users set password='"+strings.TrimSpace(password)+"' where id=2;\"")
	if err := podCmd.Run(); err != nil {
		fmt.Println("error updating password", err)
		return err
	}
	return nil
}

// func change_password(new_password string) string {
//     // """Generate Password Hash"""
//     new_password := new_password.strip()
//     if new_password == ""{
//         return ""
// 	}
//     ctx = CryptContext(schemes=["pbkdf2_sha512"])
//     pw_hash = ctx.hash(new_password)
//     return pw_hash
// }

// define the given charset, char only

// n is the length of random string we want to generate
func randStr(n int) string {
	charset := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
