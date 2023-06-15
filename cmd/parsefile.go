package cmd

import (
	"bufio"
	"os"
	"strings"
)

func parseFile(filename, key string) (value string) {
	file, err := os.Open(filename)
	if err != nil {
		return "cannot find file, make sure you are in the odoo project folder"
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	vv := []string{}
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), key) {
			vv = strings.Split(scanner.Text(), "=")
			for i := range vv {
				vv[i] = strings.TrimSpace(vv[i])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "scanner error"
	}
	if len(vv) == 2 {
		value = vv[1]
	}
	return value
}
