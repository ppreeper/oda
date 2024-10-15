package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Instance struct {
	Name  string
	State string
	IP4   string
}

// incusapi is a function that sends a query to the incus server
// and returns the response from the server
func incusapi(verb string, data string, urlparam ...string) []byte {
	urlpath, err := url.JoinPath(viper.GetString("incus.url"), urlparam...)
	cobra.CheckErr(err)

	var response *http.Response
	var httpc http.Client

	switch viper.GetString("incus.type") {
	case "http":
		httpc = http.Client{}
	case "unix":
		httpc = http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", viper.GetString("incus.socket"))
				},
			},
		}
	}

	switch verb {
	case "GET":
		req, err := http.NewRequest("GET", urlpath, nil)
		cobra.CheckErr(err)
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		cobra.CheckErr(err)
	case "POST":
		req, err := http.NewRequest("POST", urlpath, strings.NewReader(data))
		cobra.CheckErr(err)
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		cobra.CheckErr(err)
	case "PUT":
		req, err := http.NewRequest("PUT", urlpath, strings.NewReader(data))
		cobra.CheckErr(err)
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		cobra.CheckErr(err)
	case "DELETE":
		req, err := http.NewRequest("DELETE", urlpath, nil)
		cobra.CheckErr(err)
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		cobra.CheckErr(err)
	}
	defer response.Body.Close()

	respBytes, err := io.ReadAll(response.Body)
	cobra.CheckErr(err)

	return respBytes
}

func GetInstances() ([]Instance, error) {
	var instances []Instance

	respBytes := incusapi("GET", "", "instances")

	var incusInstances IncusInstances
	if err := json.Unmarshal(respBytes, &incusInstances); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return []Instance{}, err
	}
	for _, v := range incusInstances.Metadata {
		instance, err := GetInstance(strings.TrimPrefix(v, "/1.0/instances/"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error getting instance %w", err)
			return []Instance{}, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func GetInstance(instanceName string) (Instance, error) {
	respBytes := incusapi("GET", "", "instances", instanceName)
	var instance IncusInstance
	if err := json.Unmarshal(respBytes, &instance); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return Instance{}, err
	}

	incusapi("GET", "", "instances", instanceName, "state")
	instanceStatus := GetInstanceState(instanceName)
	var ip4 string
	if len(instanceStatus.Metadata.Network.Eth0.Addresses) != 0 {
		ip4 = instanceStatus.Metadata.Network.Eth0.Addresses[0].Address
	}

	return Instance{
		Name:  instance.Metadata.Name,
		State: instance.Metadata.Status,
		IP4:   ip4,
	}, nil
}

// SetInstanceState changes the state of the instance
// possible state values: start, stop, restart, freeze, unfreeze
// running values are: RUNNING, STOPPED, FROZEN
func SetInstanceState(instanceName string, state string) {
	// fmt.Println("Setting instance", instanceName, "to", state)
	data := map[string]any{
		"action":   state,
		"force":    false,
		"stateful": false,
		"timeout":  30,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshalling json", err)
		return
	}
	// fmt.Println(string(dataBytes))
	respBytes := incusapi("PUT", string(dataBytes), "instances", instanceName, "state")
	// fmt.Println(string(respBytes))
	var instanceStatus IncusInstanceStatus
	if err := json.Unmarshal(respBytes, &instanceStatus); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return
	}
	if instanceStatus.Type == "error" {
		return
	}

	var containeState string
	switch state {
	case "start":
		containeState = "RUNNING"
	case "stop":
		containeState = "STOPPED"
	case "restart":
		containeState = "RUNNING"
	case "freeze":
		containeState = "FROZEN"
	case "unfreeze":
		containeState = "RUNNING"
	}

	WaitForInstance(instanceName, containeState)
}

func GetInstanceState(instanceName string) IncusInstanceStatus {
	respBytes := incusapi("GET", "", "instances", instanceName, "state")
	var instanceStatus IncusInstanceStatus
	if err := json.Unmarshal(respBytes, &instanceStatus); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return IncusInstanceStatus{}
	}
	return instanceStatus
}

func WaitForInstance(instanceName string, containeState string) {
	for {
		currentState := GetInstanceState(instanceName)
		if strings.EqualFold(currentState.Metadata.Status, containeState) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func CreateInstance(instanceName string, source string) {
	fmt.Println("Creating instance", instanceName)
	data := map[string]any{
		"name":  instanceName,
		"type":  "container",
		"start": true,
		"source": map[string]any{
			"type":     "image",
			"alias":    source,
			"server":   "https://images.linuxcontainers.org",
			"protocol": "simplestreams",
			"mode":     "pull",
		},
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshalling json", err)
		return
	}
	// fmt.Println(string(dataBytes))
	incusapi("POST", string(dataBytes), "instances")
	// respBytes := incusapi("POST", string(dataBytes), "instances")
	// fmt.Println(string(respBytes))
	fmt.Println("Instance", instanceName, "created")
	// SetInstanceState(instanceName, "start")
}

func CopyInstance(sourceName string, instanceName string) {
	data := map[string]any{
		"name": instanceName,
		"source": map[string]any{
			"type":   "copy",
			"source": sourceName,
		},
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshalling json", err)
		return
	}
	incusapi("POST", string(dataBytes), "instances")
}

func DeleteInstance(instanceName string) {
	incusapi("DELETE", "", "instances", instanceName)
}

func IncusGetUid(instanceName, username string) (string, error) {
	// conf := GetConf()
	cmd := "incus"
	cmdArgs := []string{"exec", instanceName, "-t", "--", "grep", "^" + username, "/etc/passwd"}
	// if conf.DEBUG == "true" {
	// 	fmt.Fprintln(cmd, cmdArgs)
	// }
	out, err := exec.Command(cmd, cmdArgs...).Output()
	if err != nil {
		return "", fmt.Errorf("could not get uid for %s %w", username, err)
	}
	return strings.Split(string(out), ":")[2], nil
}

func IncusExec(instanceName string, args ...string) error {
	cmd := "incus"
	cmdArgs := []string{"exec", instanceName, "--"}
	cmdArgs = append(cmdArgs, args...)
	err := exec.Command(cmd, cmdArgs...).Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			return fmt.Errorf("failed to exec command %w", err)
		}
	}
	return nil
}

func InstanceMounts(project string) error {
	// conf := GetConf()
	repoDir := viper.GetString("dirs.repo")
	projDir := viper.GetString("dirs.project")
	cwd, _ := GetProject()
	versionFloat := viper.GetFloat64("version")
	version := fmt.Sprintf("%0.1f", versionFloat)

	IncusMount(project, "odoo", repoDir+"/"+version+"/odoo", "/opt/odoo/odoo")
	IncusMount(project, "enterprise", repoDir+"/"+version+"/enterprise", "/opt/odoo/enterprise")
	IncusMount(project, "designthemes", repoDir+"/"+version+"/design-themes", "/opt/odoo/design-themes")
	IncusMount(project, "industry", repoDir+"/"+version+"/industry", "/opt/odoo/industry")
	IncusMount(project, "backups", projDir+"/backups", "/opt/odoo/backups")
	IncusMount(project, "addons", cwd+"/addons", "/opt/odoo/addons")
	IncusMount(project, "conf", cwd+"/conf", "/opt/odoo/conf")
	IncusMount(project, "data", cwd+"/data", "/opt/odoo/data")

	return nil
}

func IncusMount(instanceName, mount, source, target string) error {
	cmd := "incus"
	cmdArgs := []string{"config", "device", "add", instanceName, mount, "disk", "source=" + source, "path=" + target}
	return exec.Command(cmd, cmdArgs...).Run()
}

func IncusIdmap(instanceName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user %w", err)
	}
	// set idmap both to map the current user to 1001
	cmd := "incus"
	cmdArgs := []string{"config", "set", instanceName, "raw.idmap", "both " + currentUser.Uid + " 1001"}
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

	err := IncusExec(instanceName, "sudo", "/usr/local/bin/odas", "hosts", domain)
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
	err := IncusExec(instanceName, "sudo", "/usr/local/bin/odas", "caddy", domain)
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

func SSHConfigGenerate(project string) error {
	HOME, _ := os.UserHomeDir()
	// conf := GetConf()
	domain := viper.GetString("system.domain")
	sshkey := viper.GetString("system.sshkey")
	instance, err := GetInstance(project)
	if err != nil {
		return fmt.Errorf("error getting instance %w", err)
	}
	// priority;host;hostname;user;identityfile;port
	sshconfig := fmt.Sprintf("%d;%s.%s;%s;%s;%s;%d", 10, project, domain, instance.IP4, "odoo", sshkey, 22)
	sshconfigCSV := filepath.Join(HOME, ".ssh", "sshconfig.csv")
	// READ config
	hosts, err := os.Open(sshconfigCSV)
	if err != nil {
		return fmt.Errorf("hosts file read failed %w", err)
	}
	defer hosts.Close()
	hostlines := []string{}
	scanner := bufio.NewScanner(hosts)
	for scanner.Scan() {
		hostlines = append(hostlines, scanner.Text())
	}
	headerLine := hostlines[0]
	// Remove old lines
	newHostlines := []string{}
	for _, hostline := range hostlines[1:] {
		hostlineSplit := strings.Split(hostline, ";")
		if hostlineSplit[1] == project+"."+domain {
			continue
		}
		newHostlines = append(newHostlines, hostline)
	}
	// Add new lines
	newHostlines = append(newHostlines, sshconfig)
	// WRITE config
	fo, err := os.Create(sshconfigCSV)
	if err != nil {
		return fmt.Errorf("hosts file write failed %w", err)
	}
	defer fo.Close()
	fo.WriteString(headerLine + "\n")
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	return nil
}
