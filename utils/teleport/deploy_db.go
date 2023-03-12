package teleport

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/project"
	"github.com/local-deploy/dl/utils/client"
	"github.com/sirupsen/logrus"
)

var remotePhpPath string

func dumpDb(ctx context.Context, t *teleport) {
	var db *project.DbSettings
	var err error

	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", Status: progress.Working})

	mysqlDataBase := project.Env.GetString("MYSQL_DATABASE_SRV")
	mysqlLogin := project.Env.GetString("MYSQL_LOGIN_SRV")
	mysqlPassword := project.Env.GetString("MYSQL_PASSWORD_SRV")
	if len(mysqlDataBase) > 0 && len(mysqlLogin) > 0 && len(mysqlPassword) > 0 {
		logrus.Info("Manual database access settings are used")
		excludedTables := strings.Split(strings.TrimSpace(project.Env.GetString("EXCLUDED_TABLES")), ",")

		db = &project.DbSettings{
			Host:           project.Env.GetString("MYSQL_HOST_SRV"),
			Port:           project.Env.GetString("MYSQL_PORT_SRV"),
			DataBase:       mysqlDataBase,
			Login:          mysqlLogin,
			Password:       mysqlPassword,
			ExcludedTables: excludedTables,
		}
	} else {
		t.checkPhpAvailable()

		logrus.Info("Attempt to access database")
		db, err = t.accessBitrixDb()

		if err != nil {
			w.Event(progress.ErrorMessageEvent("Database access error", fmt.Sprint(err)))
			return
		}
	}

	if len(db.Port) == 0 {
		logrus.Info("Port not set, standard port 3306 is used")
		db.Port = "3306"
	}

	err = t.mysqlDump(ctx, db)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed to create database dump", fmt.Sprint(err)))
		w.Event(progress.Event{ID: "Files", Status: progress.Error})
		return
	}

	err = t.downloadDump(ctx)
	if err != nil {
		w.Event(progress.Event{ID: "Files", Status: progress.Error})
		return
	}

	sshClient := &client.Client{Config: &client.Config{FwType: "bitrix"}}
	c := &project.SshClient{Client: sshClient}
	c.ImportDb(ctx)

	w.Event(progress.Event{
		ID:     "Database",
		Status: progress.Done,
	})
}

func (t *teleport) checkPhpAvailable() {
	logrus.Info("Check if PHP available")
	phpCmd := strings.Join([]string{"cd", t.Catalog, "&&", "which php"}, " ")
	logrus.Infof("Run command: %s", phpCmd)
	binary, err := t.run(phpCmd)
	if err == nil {
		remotePhpPath = binary
		logrus.Infof("PHP available: %s", remotePhpPath)
		return
	}
	logrus.Info("PHP not available")
}

func (t *teleport) accessBitrixDb() (*project.DbSettings, error) {
	serverPath := filepath.Join(t.Catalog, "bitrix/.settings.php")
	localPath := filepath.Join(project.Env.GetString("PWD"), ".tmp.php")
	err := t.download(serverPath, localPath)
	if err != nil {
		return nil, err
	}

	// TODO: Fix!
	catCmd := strings.Join([]string{"cd", project.Env.GetString("PWD"), "&&",
		`$(which php) -r '$settings = include ".tmp.php"; echo $settings["connections"]["value"]["default"]["host"]."\n";
echo $settings["connections"]["value"]["default"]["database"]."\n";
echo $settings["connections"]["value"]["default"]["login"]."\n";
echo $settings["connections"]["value"]["default"]["password"]."\n";'`,
	}, " ")
	cat, err := exec.Command("bash", "-c", catCmd).CombinedOutput()
	if err != nil {
		return nil, err
	}

	dbArray := helper.CleanSlice(strings.Split(strings.TrimSpace(string(cat)), "\n"))
	logrus.Infof("Received variables: %s", dbArray)
	if len(dbArray) != 4 {
		return nil, errors.New("failed to define DB variables, please specify accesses manually")
	}

	rmCmd := strings.Join([]string{"cd", project.Env.GetString("PWD"), "&&", "rm", ".tmp.php"}, " ")
	_, err = exec.Command("bash", "-c", rmCmd).CombinedOutput()
	if err != nil {
		return nil, err
	}

	excludedTables := strings.Split(strings.TrimSpace(project.Env.GetString("EXCLUDED_TABLES")), ",")

	return &project.DbSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

func (t *teleport) mysqlDump(ctx context.Context, db *project.DbSettings) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Working})

	dump := t.checkMySqlDumpAvailable()
	if dump != nil {
		return errors.New("mysqldump not installed, database dump not possible")
	}

	ignoredTablesString := db.FormatIgnoredTables()
	dumpTablesParams := db.DumpTablesParams()
	dumpDataParams := db.DumpDataParams()
	dumpCmd := strings.Join([]string{"cd", t.Catalog, "&&",
		"mysqldump",
		dumpTablesParams,
		db.DataBase,
		"|",
		"gzip > " + t.Catalog + "/production.sql.gz",
		"&&",
		"mysqldump",
		dumpDataParams,
		ignoredTablesString,
		db.DataBase,
		"|",
		"gzip >> " + t.Catalog + "/production.sql.gz",
	}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := t.run(dumpCmd)

	if err != nil {
		return err
	}

	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Done})

	return nil
}

func (t *teleport) checkMySqlDumpAvailable() error {
	logrus.Info("Check if mysqldump available")
	dumpCmd := strings.Join([]string{"cd", t.Catalog, "&&", "which mysqldump"}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := t.run(dumpCmd)
	if err != nil {
		logrus.Info("mysqldump not available")
		return err
	}
	logrus.Info("mysqldump available")
	return nil
}

func (t *teleport) downloadDump(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Working})

	serverPath := filepath.Join(t.Catalog, "production.sql.gz")
	localPath := filepath.Join(project.Env.GetString("PWD"), "production.sql.gz")

	logrus.Infof("Download dump: %s", serverPath)
	err := t.download(serverPath, localPath)

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Download error", fmt.Sprint(err)))
		return err
	}

	err = t.delete(serverPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("File deletion error", fmt.Sprint(err)))
		return err
	}

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Done})

	return err
}
