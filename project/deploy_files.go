package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/varrcan/dl/helper"
)

// CopyFiles Copying files from the server
func (c SshClient) CopyFiles() {
	var err error

	switch c.Server.FwType {
	case "bitrix":
		err = c.packFiles("bitrix")
	case "wordpress":
		err = c.packFiles("wp-admin wp-includes")
	default:
		return
	}

	if err != nil {
		pterm.FgRed.Printfln("Error: %s \n", err)
		os.Exit(1)
	}

	err = c.downloadArchive()
	if err == nil {
		extractArchive(c.Server.FwType)
		bitrixAccess()
		// helper.CallMethod(&err, c.Server.FwType+"Access")
	}
}

// packFiles Add files to archive
func (c SshClient) packFiles(path string) error {
	pterm.FgBlue.Println("Create files archive")

	excludeTarString := formatIgnoredPath()
	tarCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&",
		"tar",
		"--dereference",
		"-zcf",
		"production.tar.gz",
		excludeTarString,
		path,
	}, " ")
	_, err := c.Run(tarCmd)

	return err
}

// formatIgnoredPath Exclude path from tar
func formatIgnoredPath() string {
	var ignoredPath []string

	excluded := Env.GetString("EXCLUDED_FILES")
	if len(excluded) == 0 {
		return ""
	}

	excludedPath := strings.Split(strings.TrimSpace(excluded), ",")
	for _, value := range excludedPath {
		ignoredPath = append(ignoredPath, "--exclude="+value)
	}

	return strings.Join(ignoredPath, " ")
}

func (c SshClient) downloadArchive() error {
	pterm.FgBlue.Println("Download archive")
	serverPath := filepath.Join(c.Server.Catalog, "production.tar.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.tar.gz")

	err := c.download(serverPath, localPath)

	if err != nil {
		pterm.FgRed.Println("Download error: ", err)
	}

	err = c.cleanRemote(serverPath)
	if err != nil {
		pterm.FgRed.Println("File deletion error: ", err)
	}
	return err
}

func extractArchive(path string) {
	var err error

	pterm.FgBlue.Println("Extract files")

	localPath := Env.GetString("PWD")
	archive := filepath.Join(localPath, "production.tar.gz")

	// TODO: rewrite to Go
	outTar, err := exec.Command("tar", "-xzf", archive, "-C", localPath).CombinedOutput()
	outRm, err := exec.Command("rm", "-f", archive).CombinedOutput()

	err = helper.ChmodR(path, 0775)

	if err != nil {
		pterm.FgRed.Println(string(outTar))
		pterm.FgRed.Println(string(outRm))
		pterm.FgRed.Println(err)
	}
}

func bitrixAccess() {
	var err error
	localPath := Env.GetString("PWD")
	settingsFile := filepath.Join(localPath, "bitrix", ".settings.php")
	dbconnFile := filepath.Join(localPath, "bitrix", "php_interface", "dbconn.php")

	err = exec.Command("sed", "-i", "-e", `/'debug' => /c 'debug' => true,`,
		"-e", `/'host' => /c 'host' => 'db',`,
		"-e", `/'database' => /c 'database' => 'db',`,
		"-e", `/'login' => /c 'login' => 'db',`,
		"-e", `/'password' => /c 'password' => 'db',`,
		settingsFile).Run()

	err = exec.Command("sed", "-i", "-e", `/$DBHost /c $DBHost = \"db\";`,
		"-e", `/$DBLogin /c $DBLogin = \"db\";`,
		"-e", `/$DBPassword /c $DBPassword = \"db\";`,
		"-e", `/$DBName /c $DBName = \"db\";`,
		dbconnFile).Run()

	if err != nil {
		pterm.FgRed.Println(err)
	}
}
