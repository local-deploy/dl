package project

import (
	"github.com/pterm/pterm"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//CopyFiles Copying files from the server
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
		pterm.FgRed.Printfln("Error: %w \n", err)
		os.Exit(1)
	}

	err = c.downloadArchive()
	if err == nil {
		extractArchive()
	}
}

//packFiles Add files to archive
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

//formatIgnoredPath Exclude path from tar
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

func extractArchive() {
	var err error

	pterm.FgBlue.Println("Extract files")

	localPath := filepath.Join(Env.GetString("PWD"))
	archive := filepath.Join(localPath, "production.tar.gz")

	err = exec.Command("tar", "-xzf", archive, "-C", localPath).Run()
	err = exec.Command("rm", "-f", archive).Run()

	if err != nil {
		pterm.FgRed.Println(err)
	}
}
