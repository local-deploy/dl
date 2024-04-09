package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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
		return filepath.Join("/", "etc", "dl", "config-files")
	}

	return filepath.Join(ConfigDir(), "config-files")
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

// CertDir certificate directory
func CertDir() string {
	return filepath.Join(ConfigDir(), "certs")
}

// CertutilPath determine the path to the certutil
func CertutilPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		switch {
		case BinaryExists("certutil"):
			certutilPath, _ := exec.LookPath("certutil")
			return certutilPath, nil
		case BinaryExists("/usr/local/opt/nss/bin/certutil"):
			certutilPath := "/usr/local/opt/nss/bin/certutil"
			return certutilPath, nil
		default:
			out, err := exec.Command("brew", "--prefix", "nss").Output()
			if err == nil {
				certutilPath := filepath.Join(strings.TrimSpace(string(out)), "bin", "certutil")
				if pathExists(certutilPath) {
					return certutilPath, nil
				}
			}
		}

	case "linux":
		if BinaryExists("certutil") {
			certutilPath, _ := exec.LookPath("certutil")
			return certutilPath, nil
		}
	}

	certutilInstallHelp := ""
	switch {
	case BinaryExists("apt"):
		certutilInstallHelp = "apt install libnss3-tools"
	case BinaryExists("yum"):
		certutilInstallHelp = "yum install nss-tools"
	case BinaryExists("zypper"):
		certutilInstallHelp = "zypper install mozilla-nss-tools"
	}

	return "", fmt.Errorf("certutil not found. Please install it: %s", certutilInstallHelp)
}

// BinaryExists check for the existence of a binary file
func BinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// BinPath path to bin
func BinPath() string {
	return filepath.Join(binDir(), "dl")
}

// IsAptInstall checking for install from apt
func IsAptInstall() bool {
	return strings.EqualFold(binDir(), "/usr/bin")
}

// IsConfigFileExists checking for the existence of a configuration file
func IsConfigFileExists() bool {
	config := filepath.Join(ConfigDir(), "config.yaml")

	return pathExists(config)
}

// IsBinFileExists checks the existence of a binary
func IsBinFileExists() bool {
	return pathExists(BinPath())
}

// IsCertPathExists check if the certificate directory exists
func IsCertPathExists() bool {
	return pathExists(CertDir())
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

// CreateDirectory recursively create directories
func CreateDirectory(path string) error {
	if !pathExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveDirectory recursively remove directories
func RemoveDirectory(path string) error {
	if pathExists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveFilesInPath deleting files in a directory
func RemoveFilesInPath(path string) {
	if pathExists(path) {
		dir, _ := os.ReadDir(path)
		if len(dir) > 0 {
			for _, dirEntry := range dir {
				if dirEntry.IsDir() {
					continue
				}
				childPath := filepath.Join(path, dirEntry.Name())

				err := os.RemoveAll(childPath)
				if err != nil {
					continue
				}
			}
		}
	}
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
