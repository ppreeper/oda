package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func GetOdooConf(cwd, key string) string {
	odooconf := filepath.Join("/", "opt", "odoo", "conf", "odoo.conf")
	c, err := os.Open(odooconf)
	if err != nil {
		fmt.Println("Error loading odoo.conf file", err)
		return ""
	}
	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile(`^` + key + ` = (.+)$`)
		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			return match[1]
		}
	}
	return ""
}

func GetOdooBackups(project string) (backups, addons []string) {
	root_path := "/opt/odoo"
	entries, err := os.ReadDir(filepath.Join(root_path, "backups"))
	if err != nil {
		fmt.Println(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			fname := strings.Split(entry.Name(), "__")
			if len(fname) == 2 {
				backups = append(backups, entry.Name())
			} else if len(fname) == 3 {
				addons = append(addons, entry.Name())
			}
		}
	}
	slices.Sort(backups)
	slices.Sort(addons)
	backups = removeDuplicate(backups)
	addons = removeDuplicate(addons)
	if project != "" {
		backups = selectOnly(backups, project)
		addons = selectOnly(addons, project)
	}
	return
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("error opening directory %w", err)
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("error reading directory names %w", err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("error removing directory %w", err)
		}
	}
	return nil
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

func selectOnly(sliceList []string, value string) []string {
	list := []string{}
	for _, item := range sliceList {
		if strings.Contains(item, value) {
			list = append(list, item)
		}
	}
	return list
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
