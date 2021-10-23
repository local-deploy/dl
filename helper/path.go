package helper

import (
	"os"
	"path/filepath"
)

//HomeDir user home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

//ConfigDir config directory (~/.dl)
func ConfigDir() (string, error) {
	home, err := HomeDir()

	return filepath.Join(home, ".dl"), err
}

//IsConfigDirExists checking for the existence of a configuration directory
func IsConfigDirExists() bool {
	confDir, _ := ConfigDir()
	_, err := os.Stat(confDir)

	if err != nil {
		return false
	}

	return true
}
