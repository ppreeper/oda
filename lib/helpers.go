package lib

import (
	"os"
	"path/filepath"
)

func GetProject() (string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	return cwd, filepath.Base(cwd)
}
