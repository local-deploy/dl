package helper

import (
	"os"
	"path/filepath"
)

//HomeDir user home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

//ConfigDir config directory (~/.config/dl)
func ConfigDir() (string, error) {
	home, err := HomeDir()

	return filepath.Join(home, ".config/dl"), err
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

//IsConfigFileExists checking for the existence of a configuration file
func IsConfigFileExists() bool {
	confDir, _ := ConfigDir()
	config := filepath.Join(confDir, "config.yaml")

	_, err := os.Stat(config)

	if err != nil {
		return false
	}

	return true
}
