package helper

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

// HomeDir user home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// ConfigDir config directory (~/.config/dl)
func ConfigDir() string {
	conf, err := os.UserConfigDir()
	if err != nil {
		pterm.FgRed.Println(err)
		os.Exit(1)
	}

	return filepath.Join(conf, "dl")
}

// TemplateDir template directory (~/.config/dl or /etc/dl)
func TemplateDir() string {
	if IsAptInstall() {
		return filepath.Join("/", "etc", "dl")
	}

	return ConfigDir()
}

// binDir path to bin directory
func binDir() string {
	bin, err := os.Executable()
	if err != nil {
		pterm.FgRed.Println(err)
		os.Exit(1)
	}

	return path.Dir(bin)
}

// BinPath path to bin
func BinPath() string {
	return filepath.Join(binDir(), "dl")
}

// IsAptInstall checking for install from apt
func IsAptInstall() bool {
	dir := binDir()

	return strings.EqualFold(dir, "/usr/bin")
}

// IsConfigFileExists checking for the existence of a configuration file
func IsConfigFileExists() bool {
	confDir := ConfigDir()
	config := filepath.Join(confDir, "config.yaml")

	_, err := os.Stat(config)

	return err == nil
}

// IsBinFileExists checks the existence of a binary
func IsBinFileExists() bool {
	_, err := os.Stat(BinPath())

	return err == nil
}

// ChmodR change file permissions recursively
func ChmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, mode)
		}

		return err
	})
}

// CreateDirectory recursively create directories if they don't exist
func CreateDirectory(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCompose get link to executable file and arguments
func GetCompose() (string, string) {
	if isComposePlugin() {
		docker, _ := exec.LookPath("docker")
		return docker, "compose"
	}

	dockerCompose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		os.Exit(1)
	}
	return dockerCompose, ""
}

// isComposePlugin check if docker compose installed as a plugin
func isComposePlugin() bool {
	_, err := exec.Command("docker", "compose").CombinedOutput()

	return err == nil
}

// CleanSlice delete an empty value in a slice
func CleanSlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
