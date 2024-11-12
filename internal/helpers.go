package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/ui"
)

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("error opening directory %v", err)
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("error reading directory names %v", err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("error removing directory %v", err)
		}
	}
	return nil
}

func GetGitHubUsernameToken() (username, token string) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fi, err := os.Open(filepath.Join(homedir, ".gitcreds"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(fi)
	re := regexp.MustCompile(`https://(.+):(.+)@github.com$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			username = match[1]
			token = match[2]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return
}

// GetCurrentOdooRepos Get Currently Copied Odoo Repos
func GetCurrentOdooRepos() []string {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("load oda config failed %v", err.Error()))
		return []string{}
	}

	dirnames := []string{}
	entries, err := os.ReadDir(odaConf.Dirs.Repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirnames = append(dirnames, entry.Name())
		}
	}
	slices.Sort(dirnames)
	slices.Reverse(dirnames)
	dirnames = removeDuplicate(dirnames)
	dirnames = removeValue(dirnames, "odoo")
	dirnames = removeValue(dirnames, "enterprise")
	dirnames = removeValue(dirnames, "design-themes")
	dirnames = removeValue(dirnames, "industry")
	return dirnames
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetCurrentOdooProjects() []string {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("load oda config failed %v", err.Error()))
		return []string{}
	}

	dirnames := []string{}
	entries, err := os.ReadDir(odaConf.Dirs.Project)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirnames = append(dirnames, entry.Name())
		}
	}
	slices.Sort(dirnames)
	dirnames = removeDuplicate(dirnames)
	dirnames = removeValue(dirnames, "backups")
	dirnames = removeValue(dirnames, "odoo")
	dirnames = removeValue(dirnames, "enterprise")
	return dirnames
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetCurrentOdooProjectsUser(username string) []string {
	odaConf, err := config.LoadOdaConfigUser(username)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("load oda config failed %v", err.Error()))
		return []string{}
	}

	dirnames := []string{}
	entries, err := os.ReadDir(odaConf.Dirs.Project)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirnames = append(dirnames, entry.Name())
		}
	}
	slices.Sort(dirnames)
	dirnames = removeDuplicate(dirnames)
	dirnames = removeValue(dirnames, "backups")
	dirnames = removeValue(dirnames, "odoo")
	dirnames = removeValue(dirnames, "enterprise")
	return dirnames
}

func IsProject() bool {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	odaConf, _ := config.LoadOdaConfig()

	base := filepath.Base(cwd)
	pdir := filepath.Join(odaConf.Dirs.Project, base)
	odooconf := filepath.Join(cwd, "conf", "odoo.conf")

	if cwd != pdir {
		fmt.Fprintln(os.Stderr, "not in a project directory")
		return false
	}
	if _, err := os.Stat(odooconf); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "not in a project directory")
		return false
	}
	return true
}

// existsIn searches list for value
func existsIn[T comparable](sliceList []T, value T) bool {
	for _, item := range sliceList {
		if value == item {
			return true
		}
	}
	return false
}

// removeDuplicate Remove Duplicate Values from Slice
func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// removeValue Remove Value from Slice
func removeValue[T comparable](slice []T, value T) []T {
	for i, item := range slice {
		if item == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func selectOnly(sliceList []string, value string) []string {
	list := []string{}
	for _, item := range sliceList {
		if strings.Contains(item, value) {
			list = append(list, item)
		}
	}
	return list
}
