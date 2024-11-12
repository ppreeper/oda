package incus

import (
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
	"strings"
	"time"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

type Instance struct {
	Name  string
	State string
	IP4   string
}

type Incus struct {
	OdaConf *config.OdaConf
}

func NewIncus(odaconf *config.OdaConf) *Incus {
	return &Incus{
		OdaConf: odaconf,
	}
}

func (i *Incus) Incusapi(verb string, data string, urlparam ...string) []byte {
	var response *http.Response
	var httpc http.Client

	switch i.OdaConf.Incus.Type {
	case "http":
		httpc = http.Client{}
	case "unix":
		httpc = http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", i.OdaConf.Incus.Socket)
				},
			},
		}
	}

	urlpath, _ := url.JoinPath(i.OdaConf.Incus.URL, urlparam...)

	switch verb {
	case "GET":
		req, err := http.NewRequest("GET", urlpath, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating request", err)
			return []byte{}
		}
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error sending request", err)
			return []byte{}
		}
	case "POST":
		req, err := http.NewRequest("POST", urlpath, strings.NewReader(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating request", err)
			return []byte{}
		}
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error sending request", err)
			return []byte{}
		}
	case "PUT":
		req, err := http.NewRequest("PUT", urlpath, strings.NewReader(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating request", err)
			return []byte{}
		}
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error sending request", err)
			return []byte{}
		}
	case "DELETE":
		req, err := http.NewRequest("DELETE", urlpath, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating request", err)
			return []byte{}
		}
		req.Header.Set("Content-Type", "application/json")
		response, err = httpc.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error sending request", err)
			return []byte{}
		}
	}
	defer response.Body.Close()

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading response", err)
		return []byte{}
	}
	return respBytes
}

func (i *Incus) GetInstance(instanceName string) (Instance, error) {
	respBytes := i.Incusapi("GET", "", "instances", instanceName)
	var instance IncusInstance
	if err := json.Unmarshal(respBytes, &instance); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return Instance{}, err
	}

	i.Incusapi("GET", "", "instances", instanceName, "state")
	instanceStatus := i.GetInstanceState(instanceName)
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

func (i *Incus) GetInstances() ([]Instance, error) {
	var instances []Instance

	respBytes := i.Incusapi("GET", "", "instances")

	var incusInstances IncusInstances
	if err := json.Unmarshal(respBytes, &incusInstances); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return []Instance{}, err
	}
	for _, v := range incusInstances.Metadata {
		instance, err := i.GetInstance(strings.TrimPrefix(v, "/1.0/instances/"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error getting instance %w", err)
			return []Instance{}, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (i *Incus) GetInstanceState(instanceName string) IncusInstanceStatus {
	respBytes := i.Incusapi("GET", "", "instances", instanceName, "state")
	var instanceStatus IncusInstanceStatus
	if err := json.Unmarshal(respBytes, &instanceStatus); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshalling json", err)
		return IncusInstanceStatus{}
	}
	return instanceStatus
}

// SetInstanceState changes the state of the instance
// possible state values: start, stop, restart, freeze, unfreeze
// running values are: RUNNING, STOPPED, FROZEN
func (i *Incus) SetInstanceState(instanceName string, state string) {
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
	respBytes := i.Incusapi("PUT", string(dataBytes), "instances", instanceName, "state")
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
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(instanceName, "started"))
	case "stop":
		containeState = "STOPPED"
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(instanceName, "stopped"))
	case "restart":
		containeState = "RUNNING"
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(instanceName, "restarted"))
	case "freeze":
		containeState = "FROZEN"
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(instanceName, "frozen"))
	case "unfreeze":
		containeState = "RUNNING"
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(instanceName, "unfrozen"))
	}

	i.WaitForInstance(instanceName, containeState)
}

func (i *Incus) WaitForInstance(instanceName string, containeState string) {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Waiting for instance", instanceName, "to be", containeState))
	for {
		currentState := i.GetInstanceState(instanceName)
		if strings.EqualFold(currentState.Metadata.Status, containeState) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (i *Incus) CreateInstance(instanceName string, source string, cpu int, mem string) {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Creating instance", instanceName, "from", source))
	data := map[string]any{
		"name":  instanceName,
		"type":  "container",
		"start": true,
		"config": map[string]any{
			"limits.cpu":    fmt.Sprintf("%d", cpu),
			"limits.memory": mem,
		},
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
	i.Incusapi("POST", string(dataBytes), "instances")
	// respBytes := incusapi("POST", string(dataBytes), "instances")
	// fmt.Println(string(respBytes))
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Instance", instanceName, "created"))
	i.SetInstanceState(instanceName, "start")
}

func (i *Incus) CopyInstance(sourceName string, instanceName string) {
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Copying instance", instanceName, "from", sourceName))
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
	i.Incusapi("POST", string(dataBytes), "instances")
}

func (i *Incus) DeleteInstance(instanceName string) {
	fmt.Fprintln(os.Stderr, ui.WarningStyle.Render("destroying:", instanceName))
	i.Incusapi("DELETE", "", "instances", instanceName)
}

func (i *Incus) IncusGetUid(instanceName, username string) (string, error) {
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

func (i *Incus) IncusExec(instanceName string, args ...string) error {
	cmdArgs := []string{"exec", instanceName, "--"}
	cmdArgs = append(cmdArgs, args...)
	c := exec.Command("incus", cmdArgs...)
	err := c.Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			return fmt.Errorf("failed to exec command %w", err)
		}
	}
	return nil
}

func (i *Incus) IncusExecVerbose(instanceName string, args ...string) error {
	cmdArgs := []string{"exec", instanceName, "--"}
	cmdArgs = append(cmdArgs, args...)
	c := exec.Command("incus", cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			return fmt.Errorf("failed to exec command %w", err)
		}
	}
	return nil
}

func (i *Incus) InstanceMounts(project string) error {
	repoDir := i.OdaConf.Dirs.Repo
	projDir := i.OdaConf.Dirs.Project

	cwd, _ := lib.GetProject()
	projectCfg, err := config.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("could not load project config %w", err)
	}
	version := projectCfg.Version

	i.IncusMount(project, "backups", projDir+"/backups", "/opt/odoo/backups")
	i.IncusMount(project, "addons", cwd+"/addons", "/opt/odoo/addons")
	i.IncusMount(project, "conf", cwd+"/conf", "/opt/odoo/conf")
	i.IncusMount(project, "data", cwd+"/data", "/opt/odoo/data")

	branch := config.GetVersion(version)
	for _, repo := range branch.Repos {
		i.IncusMount(project, repo, repoDir+"/"+version+"/"+repo, "/opt/odoo/"+repo)
	}

	return nil
}

func (i *Incus) IncusMount(instanceName, mount, source, target string) error {
	cmd := "incus"
	cmdArgs := []string{"config", "device", "add", instanceName, mount, "disk", "source=" + source, "path=" + target}
	return exec.Command(cmd, cmdArgs...).Run()
}

func (i *Incus) IncusIdmap(instanceName string) error {
	fmt.Println("Setting idmap for", instanceName)
	currentUser, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get current user %v"), err)
		return err
	}
	// set idmap both to map the current user to 1001
	cmdArgs := []string{"config", "set", instanceName, "raw.idmap", "both " + currentUser.Uid + " 1001"}
	if err := exec.Command("incus", cmdArgs...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not set idmap %v"), err)
		return err
	}
	return nil
}
