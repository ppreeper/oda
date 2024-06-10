package oda

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

type Container struct {
	Name  string
	State string
	IP4   string
}

func GetContainers() (containers []Container, err error) {
	out, err := exec.Command("incus", "list", "-f", "csv", "-c", "ns4").Output()
	if err != nil {
		return containers, err
	}
	for _, v := range strings.Split(string(out), "\n") {
		container := strings.Split(v, ",")
		switch len(container) {
		case 3:
			ip := strings.TrimSuffix(container[2], " (eth0)")
			containers = append(containers, Container{Name: container[0], State: container[1], IP4: ip})
		case 2:
			containers = append(containers, Container{Name: container[0], State: container[1], IP4: ""})
		}
	}
	return containers, nil
}

func GetContainer(containerName string) (Container, error) {
	c, err := GetContainers()
	if err != nil {
		return Container{}, err
	}
	for _, v := range c {
		if v.Name == containerName {
			return v, nil
		}
	}
	return Container{}, fmt.Errorf("container %s not found", containerName)
}

func IncusLaunch(containerName, image string) error {
	container, _ := GetContainer(containerName)
	if container == (Container{}) {
		if err := exec.Command("incus", "launch", "images:"+image, containerName).Run(); err != nil {
			return fmt.Errorf("failed to launch container %s %w", containerName, err)
		}
	} else {
		return fmt.Errorf("container %s already exists", containerName)
	}
	return nil
}

func IncusCopy(source, target string) error {
	fmt.Println("Copying image from", source, "to", target)
	sourceContainer, _ := GetContainer(source)
	targetContainer, _ := GetContainer(target)
	if sourceContainer == (Container{}) {
		return fmt.Errorf("container %s does not exist", source)
	}
	if targetContainer != (Container{}) {
		return fmt.Errorf("container %s already exists", target)
	}
	if sourceContainer != (Container{}) && targetContainer == (Container{}) {
		if err := exec.Command("incus", "copy", source, target).Run(); err != nil {
			return fmt.Errorf("failed to copy container %s %w", source, err)
		}
	}
	return nil
}

func IncusStart(containerName string) error {
	return exec.Command("incus", "start", containerName).Run()
}

func IncusStop(containerName string) error {
	return exec.Command("incus", "stop", containerName).Run()
}

func IncusExec(containerName string, args ...string) error {
	arg := []string{"exec", containerName, "--"}
	arg = append(arg, args...)
	if err := exec.Command("incus", arg...).Run(); err != nil {
		return fmt.Errorf("failed to exec command %w", err)
	}
	return nil
}

func IncusRestart(containerName string) error {
	if err := IncusStop(containerName); err != nil {
		return fmt.Errorf("failed to stop container %s", containerName)
	}
	if err := IncusStart(containerName); err != nil {
		return fmt.Errorf("failed to start container %s", containerName)
	}
	return nil
}

func IncusDelete(containerName string) error {
	container, _ := GetContainer(containerName)
	if container != (Container{}) {
		if err := exec.Command("incus", "delete", containerName, "-f").Run(); err != nil {
			return fmt.Errorf("failed to delete container %s", containerName)
		}
	}
	return nil
}

func IncusMount(containerName, mount, source, target string) error {
	return exec.Command("incus", "config", "device", "add", containerName, mount, "disk", "source="+source, "path="+target).Run()
}

func IncusGetUid(containerName, username string) (string, error) {
	out, err := exec.Command("incus", "exec", containerName, "-t", "--", "grep", "^"+username, "/etc/passwd").Output()
	if err != nil {
		return "", fmt.Errorf("could not get uid for %s %w", username, err)
	}
	return strings.Split(string(out), ":")[2], nil
}

func IncusIdmap(containerName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user %w", err)
	}
	// set idmap both to map the current user to 1001
	if err := exec.Command("incus", "config", "set", containerName, "raw.idmap", "both "+currentUser.Uid+" 1001").Run(); err != nil {
		return fmt.Errorf("could not set idmap %w", err)
	}
	return nil
}

func IncusHosts(containerName, domain string) error {
	localHosts := "/tmp/" + containerName + "-hosts"
	remoteHosts := containerName + "/etc/hosts"

	// pull the hosts file
	if err := exec.Command("incus", "file", "pull", remoteHosts, localHosts).Run(); err != nil {
		return fmt.Errorf("hosts file pull failed %w", err)
	}

	// update the hosts file
	hosts, err := os.Open(localHosts)
	if err != nil {
		return fmt.Errorf("hosts file read failed %w", err)
	}
	defer hosts.Close()

	hostlines := []string{}
	scanner := bufio.NewScanner(hosts)
	for scanner.Scan() {
		hostlines = append(hostlines, scanner.Text())
	}

	newHostlines := []string{}

	for _, hostline := range hostlines {
		if strings.Contains(hostline, containerName) {
			newHostlines = append(newHostlines, strings.Fields(hostline)[0]+" "+containerName+" "+containerName+"."+domain)
			continue
		}
		newHostlines = append(newHostlines, hostline)
	}

	fo, err := os.Create(localHosts)
	if err != nil {
		return fmt.Errorf("hosts file write failed %w", err)
	}
	defer fo.Close()
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	// push the updated hosts file
	if err := exec.Command("incus", "file", "push", localHosts, remoteHosts).Run(); err != nil {
		return fmt.Errorf("hosts file push failed %w", err)
	}

	// clean up
	os.Remove(localHosts)

	fmt.Println("update hosts file", containerName, domain)
	return nil
}

func IncusCaddyfile(containerName, domain string) error {
	localCaddyFile := "/tmp/" + containerName + "-Caddyfile"

	if err := IncusExec(containerName, "mkdir", "-p", "/etc/caddy"); err != nil {
		return fmt.Errorf("failed to create /etc/caddy directory %w", err)
	}

	caddyTemplate := `{{.Name}}.{{.Domain}} {
		tls internal
		reverse_proxy http://{{.Name}}:8069
		reverse_proxy /websocket http://{{.Name}}:8072
		reverse_proxy /longpolling/* http://{{.Name}}:8072
		encode gzip zstd
		file_server
		log
	}`

	tmpl := template.Must(template.New("").Parse(caddyTemplate))

	f, err := os.Create(localCaddyFile)
	if err != nil {
		return fmt.Errorf("failed to create Caddyfile %w", err)
	}
	defer f.Close()
	tmpl.Execute(f, struct {
		Name   string
		Domain string
	}{
		Name:   containerName,
		Domain: domain,
	})

	if err := exec.Command("incus", "file", "push", localCaddyFile, containerName+"/etc/caddy/Caddyfile").Run(); err != nil {
		return fmt.Errorf("failed to push Caddyfile %w", err)
	}

	if err := IncusExec(containerName, "caddy", "fmt", "-w", "/etc/caddy/Caddyfile"); err != nil {
		return fmt.Errorf("failed to format Caddyfile %w", err)
	}

	os.Remove(localCaddyFile)

	return nil
}
