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

func dumpDB(ctx context.Context, t *teleport, tables []string) {
	var err error

	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", Status: progress.Working})

	db, err := t.getMysqlSettings()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprint(err)))
		return
	}

	if len(tables) > 0 {
		dumpTables := strings.Join(tables, " ")
		err = t.mysqlDumpTables(ctx, db, dumpTables)
	} else {
		err = t.mysqlDump(ctx, db)
	}

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Failed to create database dump: %s", err)))
		return
	}

	err = t.downloadDump(ctx)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Failed to download dump: %s", err)))
		return
	}

	sshClient := &client.Client{Config: &client.Config{FwType: "bitrix"}}
	c := &project.SshClient{Client: sshClient}
	err = c.ImportDB(ctx)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Access error: %s", err)))
		return
	}

	w.Event(progress.Event{ID: "Database", Status: progress.Done})
}

// Struct teleport has methods on both value and pointer receivers. Such usage is not recommended by the Go Documentation.
func (t *teleport) getMysqlSettings() (*project.DbSettings, error) {
	var db *project.DbSettings
	var err error

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
		db, err = t.accessBitrixDB()
		if err != nil {
			return nil, fmt.Errorf("access error: %w", err)
		}
	}

	if len(db.Port) == 0 {
		logrus.Info("Port not set, standard port 3306 is used")
		db.Port = "3306"
	}

	return db, nil
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

func (t *teleport) accessBitrixDB() (*project.DbSettings, error) {
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
	w.Event(progress.Event{ID: "Database", StatusText: "Create database dump"})

	dump := t.checkMySQLDumpAvailable()
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

	return nil
}

// mysqlDumpTables Create only tables dump
func (t *teleport) mysqlDumpTables(ctx context.Context, db *project.DbSettings, dumpTables string) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", StatusText: "Creating database dump"})

	dump := t.checkMySQLDumpAvailable()
	if dump != nil {
		return errors.New("mysqldump not installed, database dump not possible")
	}

	dumpDataParams := db.DumpDataTablesParams()
	dumpCmd := strings.Join([]string{"cd", t.Catalog, "&&",
		"mysqldump",
		dumpDataParams,
		db.DataBase,
		dumpTables,
		"|",
		"gzip > " + t.Catalog + "/production.sql.gz",
	}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := t.run(dumpCmd)

	if err != nil {
		return err
	}

	return nil
}

func (t *teleport) checkMySQLDumpAvailable() error {
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

	w.Event(progress.Event{ID: "Database", StatusText: "Download database dump"})

	serverPath := filepath.Join(t.Catalog, "production.sql.gz")
	localPath := filepath.Join(project.Env.GetString("PWD"), "production.sql.gz")

	logrus.Infof("Download dump: %s", serverPath)
	err := t.download(serverPath, localPath)

	if err != nil {
		return err
	}

	err = t.delete(serverPath)
	if err != nil {
		return err
	}

	return err
}
