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
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CopyFiles Copying files from the server
func CopyFiles(ctx context.Context, client *client.Client, override []string) {
	var (
		err  error
		path string
	)

	c := &SSHClient{client}

	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Files", Status: progress.Working})

	switch client.Config.FwType {
	case "bitrix":
		path = "bitrix"
	case "wordpress":
		path = "wp-admin wp-includes"
	default:
		return
	}

	if len(override) > 0 {
		path = strings.Join(override, " ")
	}

	logrus.Infof("Download path from server: %s", path)
	err = c.packFiles(ctx, path)

	if err != nil {
		fmt.Printf("Error: %s \n", err)
		os.Exit(1)
	}

	err = c.downloadArchive(ctx)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
		return
	}

	err = ExtractArchive(ctx, path)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
		return
	}

	var a CallMethod
	reflect.
		ValueOf(&a).
		MethodByName(cases.Title(language.Und, cases.NoLower).String(client.Config.FwType + "Access")).
		Call([]reflect.Value{})

	w.Event(progress.Event{ID: "Files", Status: progress.Done})
}

// packFiles Add files to archive
func (c SSHClient) packFiles(ctx context.Context, path string) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Files", StatusText: "Creating archive"})

	excludeTarString := FormatIgnoredPath()
	tarCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&",
		"tar",
		"--dereference",
		"-zcf",
		"production.tar.gz",
		excludeTarString,
		path,
	}, " ")
	logrus.Infof("Run archiving files: %s", tarCmd)
	_, err := c.Run(tarCmd)

	if err != nil {
		return err
	}

	return nil
}

// FormatIgnoredPath Exclude path from tar
func FormatIgnoredPath() string {
	excluded := Env.GetString("EXCLUDED_FILES")
	if len(excluded) == 0 {
		return ""
	}

	excludedPath := strings.Split(strings.TrimSpace(excluded), ",")
	logrus.Infof("Ignored path: %s", excluded)

	ignoredPath := make([]string, len(excludedPath))
	for i, value := range excludedPath {
		ignoredPath[i] = "--exclude=" + strings.TrimSpace(value)
	}

	return strings.Join(ignoredPath, " ")
}

func (c SSHClient) downloadArchive(ctx context.Context) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Files", StatusText: "Download archive"})

	serverPath := filepath.Join(c.Config.Catalog, "production.tar.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.tar.gz")

	logrus.Infof("Download archive: %s", serverPath)
	err := c.Download(ctx, serverPath, localPath)

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
		return err
	}

	logrus.Infof("Delete archive: %s", serverPath)
	err = c.CleanRemote(serverPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
		return err
	}

	return err
}

// ExtractArchive unzip the archive
func ExtractArchive(ctx context.Context, path string) error {
	var err error
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Files", StatusText: "Extract archive"})

	localPath := Env.GetString("PWD")
	backPath := Env.GetString("BACKEND_ROOT")
	destinationPath := localPath
	if len(backPath) > 0 {
		destinationPath = filepath.Join(localPath, backPath)
	}
	archive := filepath.Join(localPath, "production.tar.gz")
	logrus.Infof("Extract archive local path: %s", archive)

	_, err = os.Stat(destinationPath)
	if err != nil {
		if err := os.MkdirAll(destinationPath, 0o775); err != nil {
			w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
			return err
		}
	}

	// TODO: rewrite to Go
	outTar, err := exec.Command("tar", "-xzf", archive, "-C", destinationPath).CombinedOutput()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(string(outTar))))
		return err
	}

	logrus.Infof("Delete archive path: %s", archive)
	outRm, err := exec.Command("rm", "-f", archive).CombinedOutput()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(string(outRm))))
		return err
	}

	s := strings.Split(path, " ")
	for _, dir := range s {
		logrus.Infof("Run chmod 775: %s", dir)
		chmodDir := filepath.Join(destinationPath, dir)
		err = utils.ChmodR(chmodDir, 0775)
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Files", fmt.Sprint(err)))
			return err
		}
	}

	return nil
}

// BitrixAccess Change bitrix database accesses
func (a *CallMethod) BitrixAccess() {
	localPath := Env.GetString("PWD")
	backPath := Env.GetString("BACKEND_ROOT")
	destinationPath := localPath
	if len(backPath) > 0 {
		destinationPath = filepath.Join(localPath, backPath)
	}
	settingsFile := filepath.Join(destinationPath, "bitrix", ".settings.php")
	dbconnFile := filepath.Join(destinationPath, "bitrix", "php_interface", "dbconn.php")

	mysqlDB := Env.GetString("MYSQL_DATABASE")
	mysqlUser := Env.GetString("MYSQL_USER")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD")

	logrus.Infof("Replacing accesses in: %s", settingsFile)
	err := exec.Command("sed", "-i", "-e", `/'debug' *\=> /c 'debug' => true,`,
		"-e", `/'host' *\=> /c 'host' => 'db',`,
		"-e", `/'database' *\=> /c 'database' => '`+mysqlDB+`',`,
		"-e", `/'login' *\=> /c 'login' => '`+mysqlUser+`',`,
		"-e", `/'password' *\=> /c 'password' => '`+mysqlPassword+`',`,
		settingsFile).Run()
	if err != nil {
		fmt.Println(err)
	}

	logrus.Infof("Replacing accesses in: %s", dbconnFile)
	err = exec.Command("sed", "-i", "-e", `/$DBHost /c $DBHost = \"db\";`,
		"-e", `/$DBLogin /c $DBLogin = \"`+mysqlUser+`\";`,
		"-e", `/$DBPassword /c $DBPassword = \"`+mysqlPassword+`\";`,
		"-e", `/$DBName /c $DBName = \"`+mysqlDB+`\";`,
		dbconnFile).Run()

	if err != nil {
		fmt.Println(err)
	}
}

// WordpressAccess Change WordPress database accesses
func (a *CallMethod) WordpressAccess() {
	var err error
	localPath := Env.GetString("PWD")
	settingsFile := filepath.Join(localPath, "wp-config.php")

	mysqlDB := Env.GetString("MYSQL_DATABASE")
	mysqlUser := Env.GetString("MYSQL_USER")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD")

	logrus.Infof("Replacing accesses in: %s", settingsFile)
	err = exec.Command("sed", "-i",
		"-e", `/'DB_HOST'/c define('DB_HOST', 'db');`,
		"-e", `/'DB_NAME'/c define('DB_NAME', '`+mysqlDB+`');`,
		"-e", `/'DB_USER'/c define('DB_USER', '`+mysqlUser+`');`,
		"-e", `/'DB_PASSWORD'/c define('DB_PASSWORD', '`+mysqlPassword+`');`,
		settingsFile).Run()

	if err != nil {
		fmt.Println(err)
	}
}
