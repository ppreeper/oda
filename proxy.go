package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func proxyStart() error {
	fmt.Println("proxy start")
	dirs := GetDirs()
	if err := exec.Command("caddy", "start",
		"--config", filepath.Join(dirs.Project, "Caddyfile")).Run(); err != nil {
		return err
	}
	return nil
}

func proxyStop() error {
	fmt.Println("proxy stop")
	dirs := GetDirs()
	if err := exec.Command("caddy",
		"stop", "--config", filepath.Join(dirs.Project, "Caddyfile")).Run(); err != nil {
		return err
	}
	return nil
}

func proxyRestart() error {
	if err := proxyStop(); err != nil {
		return err
	}
	if err := proxyStart(); err != nil {
		return err
	}
	return nil
}

func proxyGenerate() error {
	fmt.Println("proxy generate")
	conf := GetConf()
	out, err := exec.Command("podman",
		"ps", "--format", "'{{.Image}};{{.Names}};{{.Ports}}'",
	).Output()
	if err != nil {
		return err
	}
	type Pod struct {
		Name  string
		Ports map[string]string
	}
	pods := strings.Split(string(out), "\n")
	podList := []Pod{}
	for _, pod := range pods {
		podSplit := strings.Split(pod, ";")
		if len(podSplit) == 3 {
			if strings.Contains(podSplit[0], conf.Odoobase) {
				aPod := Pod{
					Name:  "",
					Ports: make(map[string]string),
				}
				aPod.Name = podSplit[1]
				ports := strings.Split(podSplit[2], ",")
				for _, port := range ports {
					portSplit := strings.Split(port, "->")
					source := strings.Split(portSplit[0], ":")
					dest := strings.Split(portSplit[1], "/")
					aPod.Ports[dest[0]] = source[1]
				}
				podList = append(podList, aPod)
			}
		}
	}
	dirs := GetDirs()
	caddyFile := filepath.Join(dirs.Project, "Caddyfile")
	for _, pod := range podList {
		caddyOut, err := os.Create(caddyFile)
		if err != nil {
			return err
		}
		defer caddyOut.Close()
		caddyOut.WriteString(pod.Name + ":80 {" + "\n")
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
