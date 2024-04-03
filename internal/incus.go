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

func GetContainer(name string) (Container, error) {
	c, err := GetContainers()
	if err != nil {
		return Container{}, err
	}
	for _, v := range c {
		if v.Name == name {
			return v, nil
		}
	}
	return Container{}, fmt.Errorf("container %s not found", name)
}

func IncusLaunch(name, image string) error {
	container, _ := GetContainer(name)
	if container == (Container{}) {
		if err := exec.Command("incus", "launch", "images:"+image, name).Run(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("container %s already exists", name)
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
			return err
		}
	}
	return nil
}

func IncusStart(name string) error {
	if err := exec.Command("incus", "start", name).Run(); err != nil {
		return err
	}
	return nil
}

func IncusStop(name string) error {
	if err := exec.Command("incus", "stop", name).Run(); err != nil {
		return err
	}
	return nil
}

func IncusExec(name string, args ...string) error {
	arg := []string{"exec", name, "--"}
	arg = append(arg, args...)
	if err := exec.Command("incus", arg...).Run(); err != nil {
		return err
	}
	return nil
}

func IncusRestart(name string) error {
	if err := IncusStop(name); err != nil {
		return err
	}
	if err := IncusStart(name); err != nil {
		return err
	}
	return nil
}

func IncusDelete(name string) error {
	container, _ := GetContainer(name)
	if container != (Container{}) {
		if err := exec.Command("incus", "delete", name, "-f").Run(); err != nil {
			return err
		}
	}
	return nil
}

func IncusMount(name, mount, source, target string) error {
	if err := exec.Command("incus", "config", "device", "add", name, mount, "disk", "source="+source, "path="+target).Run(); err != nil {
		return err
	}
	return nil
}

func IncusGetUid(name, username string) (string, error) {
	out, err := exec.Command("incus", "exec", name, "-t", "--", "grep", "^"+username, "/etc/passwd").Output()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), ":")[2], nil
}

func IncusIdmap(name string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	// set idmap both to map the current user to 1001
	if err := exec.Command("incus", "config", "set", name, "raw.idmap", "both "+currentUser.Uid+" 1001").Run(); err != nil {
		return err
	}
	return nil
}

func IncusHosts(name, domain string) error {
	localHosts := "/tmp/" + name + "-hosts"
	remoteHosts := name + "/etc/hosts"

	// pull the hosts file
	if err := exec.Command("incus", "file", "pull", remoteHosts, localHosts).Run(); err != nil {
		return err
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
		if strings.Contains(hostline, name) {
			newHostlines = append(newHostlines, strings.Fields(hostline)[0]+" "+name+" "+name+"."+domain)
			continue
		}
		newHostlines = append(newHostlines, hostline)
	}

	fo, err := os.Create(localHosts)
	if err != nil {
		return err
	}
	defer fo.Close()
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	// push the updated hosts file
	if err := exec.Command("incus", "file", "push", localHosts, remoteHosts).Run(); err != nil {
		return err
	}

	// clean up
	os.Remove(localHosts)

	fmt.Println("update hosts file", name, domain)
	return nil
}

func IncusCaddyfile(name, domain string) error {
	localCaddyFile := "/tmp/" + name + "-Caddyfile"

	if err := IncusExec(name, "mkdir", "-p", "/etc/caddy"); err != nil {
		return err
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
		return err
	}
	defer f.Close()
	tmpl.Execute(f, struct {
		Name   string
		Domain string
	}{
		Name:   name,
		Domain: domain,
	})

	if err := exec.Command("incus", "file", "push", localCaddyFile, name+"/etc/caddy/Caddyfile").Run(); err != nil {
		return err
	}

	if err := IncusExec(name, "caddy", "fmt", "-w", "/etc/caddy/Caddyfile"); err != nil {
		return err
	}

	os.Remove(localCaddyFile)

	return nil
}
