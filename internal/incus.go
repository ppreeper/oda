package oda

import (
	"fmt"
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
	// fmt.Println("incus", "config", "device", "add", name, mount, "disk", "source="+source, "path="+target)
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
	// printf "uid $(id -u) 1000\ngid $(id -g) 1000" | lxc config set test raw.idmap -

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// restore postgresql database
	if err := exec.Command("incus", "config", "set", name, "raw.idmap", "both "+currentUser.Uid+" 1001").Run(); err != nil {
		return err
	}

	return nil
}
