package server

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/ppreeper/str"
)

func CaddyfileUpdate(domain string) error {
	hostname, _ := os.Hostname()

	caddyfile := "{{.Hostname}}.{{.Domain}} {\n" +
		"tls internal\n" +
		"reverse_proxy http://{{.Hostname}}:8069\n" +
		"reverse_proxy /websocket http://{{.Hostname}}:8072\n" +
		"reverse_proxy /longpolling/* http://{{.Hostname}}:8072\n" +
		"encode gzip zstd\n" +
		"file_server\n" +
		"log\n" +
		"}\n"
	tmpl, err := template.New("caddyfile").Parse(caddyfile)
	if err != nil {
		return err
	}

	data := struct {
		Hostname string
		Domain   string
	}{
		Hostname: hostname,
		Domain:   domain,
	}

	f, err := os.Create("/etc/caddy/Caddyfile")
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	cmd := exec.Command("sudo",
		"caddy", "fmt", "--overwrite", "/etc/caddy/Caddyfile",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("caddyfile format failed: %w", err)
	}

	return nil
}

func HostsUpdate(domain string) error {
	hostname, _ := os.Hostname()

	hosts := str.LJustLen("127.0.1.1", 15) + "{{.Hostname}} {{.Hostname}}.{{.Domain}}\n" +
		str.LJustLen("127.0.0.1", 15) + "localhost\n" +
		str.LJustLen("::1", 15) + "localhost ip6-localhost ip6-loopback\n" +
		str.LJustLen("ff02::1", 15) + "ip6-allnodes\n" +
		str.LJustLen("ff02::2", 15) + "ip6-allrouters"
	tmpl, err := template.New("caddyfile").Parse(hosts)
	if err != nil {
		return err
	}

	data := struct {
		Hostname string
		Domain   string
	}{
		Hostname: hostname,
		Domain:   domain,
	}

	f, err := os.Create("/etc/hosts")
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
