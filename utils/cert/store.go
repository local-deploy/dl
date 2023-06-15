package cert

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/local-deploy/dl/helper"
	"github.com/pterm/pterm"
)

var (
	nssDBs = []string{
		filepath.Join(os.Getenv("HOME"), ".pki/nssdb"),
		filepath.Join(os.Getenv("HOME"), "snap/chromium/current/.pki/nssdb"), // Snapcraft
		"/etc/pki/nssdb", // CentOS 7
	}
	firefoxProfiles = []string{
		os.Getenv("HOME") + "/.mozilla/firefox/*",
		os.Getenv("HOME") + "/snap/firefox/common/.mozilla/firefox/*",
	}
	firefoxPaths = []string{
		"/usr/bin/firefox",
		"/usr/bin/firefox-nightly",
		"/usr/bin/firefox-developer-edition",
		"/snap/firefox",
		"/Applications/Firefox.app",
		"/Applications/FirefoxDeveloperEdition.app",
		"/Applications/Firefox Developer Edition.app",
		"/Applications/Firefox Nightly.app",
	}
)

func hasBrowser() bool {
	allPaths := append(append([]string{}, nssDBs...), firefoxPaths...)
	for _, path := range allPaths {
		if pathExists(path) {
			return true
		}
	}
	return false
}

// Check if the certificate is installed
func (c *Cert) Check() bool {
	success := true
	if c.forEachProfile(func(profile string) {
		err := exec.Command(c.CertutilPath, "-V", "-d", profile, "-u", "L", "-n", c.caUniqueName()).Run() //nolint:gosec
		if err != nil {
			success = false
		}
	}) == 0 {
		success = false
	}
	return success
}

// Install certificate installation
func (c *Cert) Install() bool {
	if c.forEachProfile(func(profile string) {
		cmd := exec.Command(c.CertutilPath, "-A", "-d", profile, "-t", "C,,", "-n", c.caUniqueName(), "-i", filepath.Join(c.CaPath, c.CaFileName)) //nolint:gosec
		out, err := execCertutil(cmd)
		if err != nil {
			pterm.FgRed.Printfln("Error: failed to execute \"%s\": %s\n\n%s\n", "certutil -A -d "+profile, err, out)
			os.Exit(1)
		}
	}) == 0 {
		pterm.FgRed.Println("Error: no browsers security databases found")
		return false
	}
	if !c.Check() {
		pterm.FgRed.Println("Installing in browsers failed. Please report the issue with details about your environment at https://github.com/local-deploy/dl/issues/new")
		pterm.FgYellow.Println("Note that if you never started browsers, you need to do that at least once.")
		return false
	}
	return true
}

// Uninstall deleting certificate
func (c *Cert) Uninstall() {
	c.forEachProfile(func(profile string) {
		err := exec.Command(c.CertutilPath, "-V", "-d", profile, "-u", "L", "-n", c.caUniqueName()).Run() //nolint:gosec
		if err != nil {
			return
		}
		cmd := exec.Command(c.CertutilPath, "-D", "-d", profile, "-n", c.caUniqueName()) //nolint:gosec
		out, err := execCertutil(cmd)
		if err != nil {
			pterm.FgRed.Printfln("Error: failed to execute \"%s\": %s\n\n%s\n", "certutil -D -d "+profile, err, out)
		}
	})
}

func execCertutil(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil && bytes.Contains(out, []byte("SEC_ERROR_READ_ONLY")) {
		origArgs := cmd.Args[1:]
		cmd = commandWithSudo(cmd.Path)
		cmd.Args = append(cmd.Args, origArgs...)
		out, err = cmd.CombinedOutput()
	}
	return out, err
}

func (c *Cert) forEachProfile(f func(profile string)) (found int) {
	var profiles []string
	profiles = append(profiles, nssDBs...)
	for _, ff := range firefoxProfiles {
		pp, _ := filepath.Glob(ff)
		profiles = append(profiles, pp...)
	}
	for _, profile := range profiles {
		if stat, err := os.Stat(profile); err != nil || !stat.IsDir() {
			continue
		}
		if pathExists(filepath.Join(profile, "cert9.db")) {
			f("sql:" + profile)
			found++
		} else if pathExists(filepath.Join(profile, "cert8.db")) {
			f("dbm:" + profile)
			found++
		}
	}
	return
}

var sudoWarningOnce sync.Once

func commandWithSudo(cmd ...string) *exec.Cmd {
	u, err := user.Current()
	if err == nil && u.Uid == "0" {
		return exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	}
	if !helper.BinaryExists("sudo") {
		sudoWarningOnce.Do(func() {
			pterm.FgRed.Println(`Warning: "sudo" is not available, and dl is not running as root. The (un)install operation might fail.️`)
		})
		return exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	}

	userName := "user"
	if u != nil && len(u.Username) > 0 {
		userName = u.Username
	}

	return exec.Command("sudo", append([]string{"--prompt=[sudo] password for " + userName + ":", "--"}, cmd...)...) //nolint:gosec
}
