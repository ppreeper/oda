package lib

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func GetOSVersionName() (osName string, osVersionID string, osVersionCodename string) {
	osRelease, err := os.Open("/etc/os-release")
	if err != nil {
		fmt.Println("Error loading os-release file", err)
		return
	}
	defer func() {
		if err := osRelease.Close(); err != nil {
			panic(err)
		}
	}()

	reName := regexp.MustCompile(`^NAME="(.+)"$`)
	reVersionID := regexp.MustCompile(`^VERSION_ID="(.+)"$`)
	reVersionCodename := regexp.MustCompile(`^VERSION_CODENAME=(.+)$`)

	scanner := bufio.NewScanner(osRelease)
	for scanner.Scan() {
		line := scanner.Text()
		if reName.MatchString(line) {
			match := reName.FindStringSubmatch(line)
			osName = match[1]
		}
		if reVersionID.MatchString(line) {
			match := reVersionID.FindStringSubmatch(line)
			osVersionID = match[1]
		}
		if reVersionCodename.MatchString(line) {
			match := reVersionCodename.FindStringSubmatch(line)
			osVersionCodename = match[1]
		}
	}

	return strings.TrimSpace(osName), strings.TrimSpace(osVersionID), strings.TrimSpace(osVersionCodename)
}
