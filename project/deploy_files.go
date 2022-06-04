package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/varrcan/dl/helper"
)

type callMethod struct{}

// CopyFiles Copying files from the server
func (c SshClient) CopyFiles(ctx context.Context) {
	var err error
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{
		ID:     "Files",
		Status: progress.Working,
	})

	switch c.Server.FwType {
	case "bitrix":
		err = c.packFiles(ctx, "local")
	case "wordpress":
		err = c.packFiles(ctx, "wp-admin wp-includes")
	default:
		return
	}

	if err != nil {
		fmt.Printf("Error: %s \n", err)
		os.Exit(1)
	}

	err = c.downloadArchive(ctx)
	if err == nil {
		extractArchive(ctx, c.Server.FwType)

		var a callMethod
		reflect.ValueOf(&a).MethodByName(strings.Title(c.Server.FwType + "Access")).Call([]reflect.Value{})
	}

	w.Event(progress.Event{
		ID:     "Files",
		Status: progress.Done,
	})
}

// packFiles Add files to archive
func (c SshClient) packFiles(ctx context.Context, path string) error {
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{
		ID:       "Archive files",
		ParentID: "Files",
		Status:   progress.Working,
	})

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

	if err != nil {
		return err
	}

	w.Event(progress.Event{
		ID:       "Archive files",
		ParentID: "Files",
		Status:   progress.Done,
	})

	return nil
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

func (c SshClient) downloadArchive(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	serverPath := filepath.Join(c.Server.Catalog, "production.tar.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.tar.gz")

	w.Event(progress.Event{
		ID:       "Download archive",
		ParentID: "Files",
		Status:   progress.Working,
	})

	err := c.download(ctx, serverPath, localPath)

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Download error", fmt.Sprint(err)))
	}

	err = c.cleanRemote(serverPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("File deletion error", fmt.Sprint(err)))
	}

	w.Event(progress.Event{
		ID:       "Download archive",
		ParentID: "Files",
		Status:   progress.Done,
	})

	return err
}

func extractArchive(ctx context.Context, path string) {
	var err error
	w := progress.ContextWriter(ctx)

	// w.Event(progress.Waiting("Extract files"))

	w.Event(progress.Event{
		ID:       "Extract archive",
		ParentID: "Files",
		Status:   progress.Working,
	})

	localPath := Env.GetString("PWD")
	archive := filepath.Join(localPath, "production.tar.gz")

	// TODO: rewrite to Go
	outTar, err := exec.Command("tar", "-xzf", archive, "-C", localPath).CombinedOutput()
	outRm, err := exec.Command("rm", "-f", archive).CombinedOutput()

	err = helper.ChmodR(path, 0775)

	if err != nil {
		fmt.Println(string(outTar))
		fmt.Println(string(outRm))
		fmt.Println(err)
	}

	w.Event(progress.Event{
		ID:       "Extract archive",
		ParentID: "Files",
		Status:   progress.Done,
	})
}

// BitrixAccess Change bitrix database accesses
func (a *callMethod) BitrixAccess() {
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
		fmt.Println(err)
	}
}

// WordpressAccess Change WordPress database accesses
func (a *callMethod) WordpressAccess() {
	var err error
	localPath := Env.GetString("PWD")
	settingsFile := filepath.Join(localPath, "wp-config.php")

	err = exec.Command("sed", "-i",
		"-e", `/'DB_HOST' => /c define('DB_HOST', 'db');`,
		"-e", `/'DB_NAME' => /c define('DB_NAME', 'db');`,
		"-e", `/'DB_USER' => /c define('DB_USER', 'db');`,
		"-e", `/'DB_PASSWORD' => /c define('DB_PASSWORD', 'db');`,
		settingsFile).Run()

	if err != nil {
		fmt.Println(err)
	}
}
