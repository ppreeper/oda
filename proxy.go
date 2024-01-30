package oda

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	pods, err := GetPods(false)
	if err != nil {
		return err
	}
	podList := []Pod{}
	for _, pod := range pods {
		if strings.Contains(pod.Image, conf.Odoobase) {
			podList = append(podList, pod)
		}
	}
	dirs := GetDirs()
	caddyFile := filepath.Join(dirs.Project, "Caddyfile")
	caddyOut, err := os.Create(caddyFile)
	if err != nil {
		return err
	}
	defer caddyOut.Close()
	for _, pod := range podList {
		caddyOut.WriteString(pod.Name + " {" + "\n")
		caddyOut.WriteString("tls internal" + "\n")
		caddyOut.WriteString("reverse_proxy http://127.0.0.1:" + pod.Ports["8069"] + "\n")
		caddyOut.WriteString("reverse_proxy /websocket http://127.0.0.1:" + pod.Ports["8072"] + "\n")
		caddyOut.WriteString("reverse_proxy /longpolling/* http://127.0.0.1:" + pod.Ports["8072"] + "\n")
		caddyOut.WriteString("}" + "\n")
	}

	if err := exec.Command("caddy", "fmt", "-w", caddyFile).Run(); err != nil {
		return err
	}
	return nil
}
