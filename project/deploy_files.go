package project

import (
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"strings"
)

//CopyFiles Copying files from the server
func (c SshClient) CopyFiles() {
	var err error

	switch c.Server.FwType {
	case "bitrix":
		err = c.packFiles("bitrix")
	case "laravel": //TODO
	case "wordpress": //TODO
	}

	if err != nil {
		pterm.FgRed.Printfln("Error: %w \n", err)
		os.Exit(1)
	}

	c.downloadArchive()
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
		path + "/",
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

func (c SshClient) downloadArchive() {
	pterm.FgBlue.Println("Download archive")
	serverPath := filepath.Join(c.Server.Catalog, "production.tar.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.tar.gz")

	err := c.Download(serverPath, localPath)

	if err != nil {
		pterm.FgRed.Println("Download error: ", err)
		os.Exit(1)
	}

	pterm.FgBlue.Println("Cleaning up temporary archive")
	err = c.cleanRemote(serverPath)
	if err != nil {
		pterm.FgRed.Println("File deletion error: ", err)
	}
}
