package oda

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ProxyStart() error {
	fmt.Println("proxy start")
	dirs := GetDirs()
	if err := exec.Command("caddy", "start",
		"--config", filepath.Join(dirs.Project, "Caddyfile")).Run(); err != nil {
		return err
	}
	return nil
}

func ProxyStop() error {
	fmt.Println("proxy stop")
	dirs := GetDirs()
	if err := exec.Command("caddy",
		"stop", "--config", filepath.Join(dirs.Project, "Caddyfile")).Run(); err != nil {
		return err
	}
	return nil
}

func ProxyRestart() error {
	if err := ProxyStop(); err != nil {
		return err
	}
	if err := ProxyStart(); err != nil {
		return err
	}
	return nil
}

func ProxyGenerate() error {
	fmt.Println("proxy generate")
	conf := GetConf()

	projects := GetCurrentOdooProjects()

	containers, err := GetContainers()
	if err != nil {
		return err
	}

	dirs := GetDirs()
	caddyFile := filepath.Join(dirs.Project, "Caddyfile")
	caddyOut, err := os.Create(caddyFile)
	if err != nil {
		return err
	}
	defer caddyOut.Close()
	for _, container := range containers {
		for _, project := range projects {
			if container.Name == project {
				caddyOut.WriteString(container.Name + "." + conf.Domain + " {" + "\n")
				caddyOut.WriteString("tls internal" + "\n")
				caddyOut.WriteString("reverse_proxy http://" + container.IP4 + ":8069" + "\n")
				caddyOut.WriteString("reverse_proxy /websocket http://" + container.IP4 + ":8072" + "\n")
				caddyOut.WriteString("reverse_proxy /longpolling/* http://" + container.IP4 + ":8072" + "\n")
				caddyOut.WriteString("encode gzip zstd" + "\n")
				caddyOut.WriteString("file_server" + "\n")
				caddyOut.WriteString("log" + "\n")
				caddyOut.WriteString("}" + "\n")
			}
		}
	}
	if err := exec.Command("caddy", "fmt", "-w", caddyFile).Run(); err != nil {
		return err
	}
	return nil
}
