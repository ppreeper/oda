package lib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ppreeper/oda/ui"
)

// ########

func UserHomeDir(username string) (homedir string) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("/etc/passwd read failed %v"), err)
		return
	}
	defer passwd.Close()

	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		passwdUser := strings.Split(scanner.Text(), ":")
		if passwdUser[0] == username {
			homedir = passwdUser[5]
		}
	}
	return
}

func UserConfigDir(username string) (configdir string) {
	homedir := UserHomeDir(username)
	return filepath.Join(homedir, ".config")
}
