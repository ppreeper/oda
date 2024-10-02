package oda

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

type Instance struct {
	Name  string
	State string
	IP4   string
}

func GetInstances() (instances []Instance, err error) {
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"list", "-f", "csv", "-c", "ns4"}
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	out, err := exec.Command(cmd, cmdArgs...).Output()
	if err != nil {
		return instances, err
	}
	for _, v := range strings.Split(string(out), "\n") {
		instance := strings.Split(v, ",")
		switch len(instance) {
		case 3:
			ip := strings.Split(instance[2], " ")[0]
			instances = append(instances, Instance{Name: instance[0], State: instance[1], IP4: ip})
		case 2:
			instances = append(instances, Instance{Name: instance[0], State: instance[1], IP4: ""})
		}
	}
	return instances, nil
}

func GetInstance(instanceName string) (Instance, error) {
	c, err := GetInstances()
	if err != nil {
		return Instance{}, err
	}
	for _, v := range c {
		if v.Name == instanceName {
			return v, nil
		}
	}
	return Instance{}, fmt.Errorf("instance %s not found", instanceName)
}

func WaitForInstance(instanceName string) error {
	fmt.Println("Waiting for instance", instanceName)
	for {
		instance, _ := GetInstance(instanceName)
		if instance == (Instance{}) {
			return fmt.Errorf("instance %s not found", instanceName)
		}
		if instance.State == "RUNNING" && instance.IP4 != "" {
			// debounce
			time.Sleep(5 * time.Second)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func IncusLaunch(instanceName, image string) error {
	fmt.Println("IncusLaunch", "instanceName", instanceName, "image", image)
	conf := GetConf()
	instance, _ := GetInstance(instanceName)
	if instance == (Instance{}) {
		cmd := "incus"
		cmdArgs := []string{"launch", "images:" + image, instanceName}
		if conf.InstanceType == "VM" {
			cmdArgs = append(cmdArgs, "--vm")
		}
		if conf.DEBUG == "true" {
			fmt.Println(cmd, cmdArgs)
		}
		if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
			return fmt.Errorf("failed to launch instance %s %w", instanceName, err)
		}
	} else {
		return fmt.Errorf("instance %s already exists", instanceName)
	}
	return nil
}

func IncusCopy(source, target string) error {
	fmt.Println("Copying image from", source, "to", target)
	conf := GetConf()
	sourceInstance, _ := GetInstance(source)
	targetInstance, _ := GetInstance(target)
	if sourceInstance == (Instance{}) {
		return fmt.Errorf("instance %s does not exist", source)
	}
	if targetInstance != (Instance{}) {
		return fmt.Errorf("instance %s already exists", target)
	}
	if sourceInstance != (Instance{}) && targetInstance == (Instance{}) {
		cmd := "incus"
		cmdArgs := []string{"copy", source, target}
		if conf.DEBUG == "true" {
			fmt.Println(cmd, cmdArgs)
		}
		if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
			return fmt.Errorf("failed to copy instance %s %w", source, err)
		}
	}
	return nil
}

func IncusStart(instanceName string) error {
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"start", instanceName}
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
		return fmt.Errorf("failed to start instance %s %w", instanceName, err)
	}

	if err := WaitForInstance(instanceName); err != nil {
		if conf.DEBUG == "true" {
			fmt.Println(err)
		}
		return fmt.Errorf("error waiting for instance %w", err)
	}

	return nil
}

func IncusStop(instanceName string) error {
	fmt.Println("IncusStop:", instanceName)
	return exec.Command("incus", "stop", instanceName).Run()
}

func IncusExec(instanceName string, args ...string) error {
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"exec", instanceName, "--"}
	cmdArgs = append(cmdArgs, args...)
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	err := exec.Command(cmd, cmdArgs...).Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			return fmt.Errorf("failed to exec command %w", err)
		}
	}
	return nil
}

func IncusRestart(instanceName string) error {
	if err := IncusStop(instanceName); err != nil {
		return fmt.Errorf("failed to stop instance %s", instanceName)
	}
	if err := IncusStart(instanceName); err != nil {
		return fmt.Errorf("failed to start instance %s", instanceName)
	}
	return nil
}

func IncusDelete(instanceName string) error {
	conf := GetConf()
	instance, _ := GetInstance(instanceName)
	if instance != (Instance{}) {
		cmd := "incus"
		cmdArgs := []string{"delete", instanceName, "-f"}
		if conf.DEBUG == "true" {
			fmt.Println(cmd, cmdArgs)
		}
		if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
			return fmt.Errorf("failed to delete instance %s", instanceName)
		}
	}
	return nil
}

func IncusMount(instanceName, mount, source, target string) error {
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"config", "device", "add", instanceName, mount, "disk", "source=" + source, "path=" + target}
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	return exec.Command(cmd, cmdArgs...).Run()
}

func IncusGetUid(instanceName, username string) (string, error) {
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"exec", instanceName, "-t", "--", "grep", "^" + username, "/etc/passwd"}
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	out, err := exec.Command(cmd, cmdArgs...).Output()
	if err != nil {
		return "", fmt.Errorf("could not get uid for %s %w", username, err)
	}
	return strings.Split(string(out), ":")[2], nil
}

func IncusIdmap(instanceName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user %w", err)
	}
	// set idmap both to map the current user to 1001
	conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"config", "set", instanceName, "raw.idmap", "both " + currentUser.Uid + " 1001"}
	if conf.DEBUG == "true" {
		fmt.Println(cmd, cmdArgs)
	}
	if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
		return fmt.Errorf("could not set idmap %w", err)
	}
	return nil
}

func IncusHosts(instanceName, domain string) error {
	// localHosts := "/tmp/" + instanceName + "-hosts"
	// remoteHosts := instanceName + "/etc/hosts"

	// os.Remove(localHosts)
	// pull the hosts file
	// cmd := "incus"
	// cmdArgs := []string{"file", "pull", remoteHosts, localHosts}
	// if conf.DEBUG == "true" {
	// 	fmt.Println(cmd, cmdArgs)
	// }
	// err := exec.Command(cmd, cmdArgs...).Run()
	// if errors.Is(err, exec.ErrWaitDelay) {
	// 	return fmt.Errorf("hosts file %v pull to %v failed %w", remoteHosts, localHosts, err)
	// }

	err := IncusExec(instanceName, "sudo", "/usr/local/bin/oda", "hosts", domain)
	if errors.Is(err, exec.ErrWaitDelay) {
		return fmt.Errorf("failed to update hosts file %w", err)
	}

	// update the hosts file
	// hosts, err := os.Open(localHosts)
	// if err != nil {
	// 	return fmt.Errorf("hosts file read failed %w", err)
	// }
	// defer hosts.Close()

	// hostlines := []string{}
	// scanner := bufio.NewScanner(hosts)
	// for scanner.Scan() {
	// 	hostlines = append(hostlines, scanner.Text())
	// }

	// newHostlines := []string{}

	// for _, hostline := range hostlines {
	// 	if strings.Contains(hostline, instanceName) {
	// 		newHostlines = append(newHostlines, strings.Fields(hostline)[0]+" "+instanceName+" "+instanceName+"."+domain)
	// 		continue
	// 	}
	// 	newHostlines = append(newHostlines, hostline)
	// }

	// fo, err := os.Create(localHosts)
	// if err != nil {
	// 	return fmt.Errorf("hosts file write failed %w", err)
	// }
	// defer fo.Close()
	// for _, hostline := range newHostlines {
	// 	fo.WriteString(hostline + "\n")
	// }

	// push the updated hosts file
	// cmd = "incus"
	// cmdArgs = []string{"file", "push", localHosts, remoteHosts}
	// if conf.DEBUG == "true" {
	// 	fmt.Println(cmd, cmdArgs)
	// }
	// err = exec.Command(cmd, cmdArgs...).Run()
	// if errors.Is(err, exec.ErrWaitDelay) {
	// 	return fmt.Errorf("hosts file push failed %w", err)
	// }

	// // clean up
	// os.Remove(localHosts)

	fmt.Println("update hosts file", instanceName, domain)
	return nil
}

func IncusCaddyfile(instanceName, domain string) error {
	err := IncusExec(instanceName, "sudo", "/usr/local/bin/oda", "caddy", domain)
	if errors.Is(err, exec.ErrWaitDelay) {
		return fmt.Errorf("failed to update hosts file %w", err)
	}

	// localCaddyFile := "/tmp/" + instanceName + "-Caddyfile"

	// if err := IncusExec(instanceName, "mkdir", "-p", "/etc/caddy"); err != nil {
	// 	return fmt.Errorf("failed to create /etc/caddy directory %w", err)
	// }

	// caddyTemplate := `{{.Name}}.{{.Domain}} {
	// 	tls internal
	// 	reverse_proxy http://{{.Name}}:8069
	// 	reverse_proxy /websocket http://{{.Name}}:8072
	// 	reverse_proxy /longpolling/* http://{{.Name}}:8072
	// 	encode gzip zstd
	// 	file_server
	// 	log
	// }`

	// tmpl := template.Must(template.New("").Parse(caddyTemplate))

	// f, err := os.Create(localCaddyFile)
	// if err != nil {
	// 	return fmt.Errorf("failed to create Caddyfile %w", err)
	// }
	// defer f.Close()
	// tmpl.Execute(f, struct {
	// 	Name   string
	// 	Domain string
	// }{
	// 	Name:   instanceName,
	// 	Domain: domain,
	// })

	// cmd := "incus"
	// cmdArgs := []string{"file", "push", localCaddyFile, instanceName + "/etc/caddy/Caddyfile"}
	// if conf.DEBUG == "true" {
	// 	fmt.Println(cmd, cmdArgs)
	// }
	// if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
	// 	return fmt.Errorf("failed to push Caddyfile %w", err)
	// }

	// if err := IncusExec(instanceName, "caddy", "fmt", "-w", "/etc/caddy/Caddyfile"); err != nil {
	// 	return fmt.Errorf("failed to format Caddyfile %w", err)
	// }

	// os.Remove(localCaddyFile)

	return nil
}
