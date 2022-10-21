package project

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/sirupsen/logrus"
	"github.com/varrcan/dl/helper"
	"github.com/varrcan/dl/utils/client"
)

var remotePhpPath string

// DumpDb Database import from server
func DumpDb(ctx context.Context, client *client.Client) {
	var db *dbSettings
	var err error

	c := &sshClient{client}

	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", Status: progress.Working})

	mysqlDataBase := Env.GetString("MYSQL_DATABASE_SRV")
	mysqlLogin := Env.GetString("MYSQL_LOGIN_SRV")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD_SRV")
	if len(mysqlDataBase) > 0 && len(mysqlLogin) > 0 && len(mysqlPassword) > 0 {
		logrus.Info("Manual database access settings are used")
		excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

		db = &dbSettings{
			Host:           Env.GetString("MYSQL_HOST_SRV"),
			Port:           Env.GetString("MYSQL_PORT_SRV"),
			DataBase:       mysqlDataBase,
			Login:          mysqlLogin,
			Password:       mysqlPassword,
			ExcludedTables: excludedTables,
		}
	} else {
		c.checkPhpAvailable()

		logrus.Info("Attempt to access database")
		switch c.Config.FwType {
		case "bitrix":
			db, err = c.accessBitrixDb()
		case "laravel":
			db, err = c.accessLaravelDb()
		case "wordpress":
			db, err = c.accessWpDb()
		}

		if err != nil {
			w.Event(progress.ErrorMessageEvent("Database access error", fmt.Sprint(err)))
			return
		}
	}

	if len(db.Port) == 0 {
		logrus.Info("Port not set, standard port 3306 is used")
		db.Port = "3306"
	}

	err = c.mysqlDump(ctx, db)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed to create database dump", fmt.Sprint(err)))
		w.Event(progress.Event{ID: "Files", Status: progress.Error})
		return
	}

	err = c.downloadDump(ctx)
	if err != nil {
		w.Event(progress.Event{ID: "Files", Status: progress.Error})
		return
	}

	c.importDb(ctx)

	w.Event(progress.Event{
		ID:     "Database",
		Status: progress.Done,
	})
}

// checkPhpAvailable It possible that PHP not installed on the server in the host system. For example, through docker.
func (c sshClient) checkPhpAvailable() {
	logrus.Info("Check if PHP available")
	phpCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&", "which php"}, " ")
	logrus.Infof("Run command: %s", phpCmd)
	binary, err := c.Run(phpCmd)
	if err == nil {
		remotePhpPath = string(binary)
		logrus.Infof("PHP available: %s", remotePhpPath)
		return
	}
	logrus.Info("PHP not available")
}

// accessBitrixDb Attempt to determine database accesses
func (c sshClient) accessBitrixDb() (*dbSettings, error) {
	var catCmd string
	if len(remotePhpPath) > 0 {
		// A more precise way to define variables
		catCmd = strings.Join([]string{"cd", c.Config.Catalog, "&&",
			`$(which php) -r '$settings = include "bitrix/.settings.php"; echo $settings["connections"]["value"]["default"]["host"]."\n";
echo $settings["connections"]["value"]["default"]["database"]."\n";
echo $settings["connections"]["value"]["default"]["login"]."\n";
echo $settings["connections"]["value"]["default"]["password"]."\n";'`,
		}, " ")
	} else {
		// Defining variables with grep
		catCmd = strings.Join([]string{"cd", c.Config.Catalog, "&&",
			`cat bitrix/.settings.php | grep "'host' *\=>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
			`cat bitrix/.settings.php | grep "'database' *\=>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
			`cat bitrix/.settings.php | grep "'login' *\=>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
			`cat bitrix/.settings.php | grep "'password' *\=>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`,
		}, " ")
	}

	logrus.Infof("Run command: %s", catCmd)
	cat, err := c.Run(catCmd)
	if err != nil {
		return nil, err
	}

	dbArray := helper.CleanSlice(strings.Split(strings.TrimSpace(string(cat)), "\n"))
	logrus.Infof("Received variables: %s", dbArray)
	if len(dbArray) != 4 {
		return nil, errors.New("failed to define DB variables, please specify accesses manually")
	}

	excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

	return &dbSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

// accessWpDb Attempt to determine database accesses
func (c sshClient) accessWpDb() (*dbSettings, error) {
	catCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&",
		`$(which php) -r 'error_reporting(0); define("SHORTINIT",true); $settings = include "wp-config.php"; echo DB_HOST."\n"; echo DB_NAME."\n"; echo DB_USER."\n"; echo DB_PASSWORD."\n";'`,
	}, " ")
	logrus.Infof("Run command: %s", catCmd)
	cat, err := c.Run(catCmd)
	if err != nil {
		return nil, err
	}

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")
	logrus.Infof("Received variables: %s", dbArray)
	if len(dbArray) != 4 {
		return nil, errors.New("failed to define DB variables, please specify accesses manually")
	}

	excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

	return &dbSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

func (c sshClient) accessLaravelDb() (*dbSettings, error) {
	catCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&", "export $(grep -v '^#' .env | xargs)", "&&",
		`echo $DB_HOST`, "&&",
		`echo $DB_DATABASE`, "&&",
		`echo $DB_USERNAME`, "&&",
		`echo $DB_PASSWORD`,
	}, " ")
	logrus.Infof("Run command: %s", catCmd)
	cat, err := c.Run(catCmd)

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")
	logrus.Infof("Received variables: %s", dbArray)
	if len(dbArray) != 4 {
		return nil, errors.New("failed to define DB variables, please specify accesses manually")
	}

	excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

	return &dbSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

// mysqlDump Create database dump
func (c sshClient) mysqlDump(ctx context.Context, db *dbSettings) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Working})

	dump := c.checkMySqlDumpAvailable()
	if dump != nil {
		return errors.New("mysqldump not installed, database dump not possible")
	}

	ignoredTablesString := db.formatIgnoredTables()
	dumpTablesParams := db.dumpTablesParams()
	dumpDataParams := db.dumpDataParams()
	dumpCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&",
		"mysqldump",
		dumpTablesParams,
		db.DataBase,
		"|",
		"gzip > " + c.Config.Catalog + "/production.sql.gz",
		"&&",
		"mysqldump",
		dumpDataParams,
		ignoredTablesString,
		db.DataBase,
		"|",
		"gzip >> " + c.Config.Catalog + "/production.sql.gz",
	}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := c.Run(dumpCmd)

	if err != nil {
		return err
	}

	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Done})

	return nil
}

func (c sshClient) checkMySqlDumpAvailable() error {
	logrus.Info("Check if mysqldump available")
	dumpCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&", "which mysqldump"}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := c.Run(dumpCmd)
	if err != nil {
		logrus.Info("mysqldump not available")
		return err
	}
	logrus.Info("mysqldump available")
	return nil
}

func (d dbSettings) dumpTablesParams() string {
	params := []string{
		"--host=" + d.Host,
		"--port=" + d.Port,
		"--user=" + d.Login,
		"--password=" + strconv.Quote(d.Password),
		"--single-transaction=1",
		"--lock-tables=false",
		"--no-data",
		"--no-tablespaces",
	}

	mysqlVersion := Env.GetString("MYSQL_VERSION")
	if mysqlVersion == "8.0" {
		params = append(params, "--column-statistics=0")
	}

	return strings.Join(params, " ")
}

func (d dbSettings) dumpDataParams() string {
	params := []string{
		"--host=" + d.Host,
		"--port=" + d.Port,
		"--user=" + d.Login,
		"--password=" + strconv.Quote(d.Password),
		"--single-transaction=1",
		"--force",
		"--lock-tables=false",
		"--no-tablespaces",
		"--no-create-info",
	}

	mysqlVersion := Env.GetString("MYSQL_VERSION")
	if mysqlVersion == "8.0" {
		params = append(params, "--column-statistics=0")
	}

	return strings.Join(params, " ")
}

// formatIgnoredTables Exclude tables from dump
func (d dbSettings) formatIgnoredTables() string {
	var ignoredTables []string

	if len(d.ExcludedTables) == 0 {
		return ""
	}

	for _, value := range d.ExcludedTables {
		ignoredTables = append(ignoredTables, "--ignore-table="+d.DataBase+"."+value)
	}

	return strings.Join(ignoredTables, " ")
}

// downloadDump Downloading a dump and deleting an archive from the server
func (c sshClient) downloadDump(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Working})

	serverPath := filepath.Join(c.Config.Catalog, "production.sql.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")

	logrus.Infof("Download dump: %s", serverPath)
	err := c.Download(ctx, serverPath, localPath)

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Download error", fmt.Sprint(err)))
		return err
	}

	err = c.CleanRemote(serverPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("File deletion error", fmt.Sprint(err)))
		return err
	}

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Done})

	return err
}

// importDb Importing a database into a local container
func (c sshClient) importDb(ctx context.Context) {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Working})

	docker, err := exec.LookPath("docker")
	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: fmt.Sprint(err)})
		return
	}
	gunzip, err := exec.LookPath("gunzip")
	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: fmt.Sprint(err)})
		return
	}

	// TODO: переписать на sdk
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")
	site := Env.GetString("HOST_NAME")
	siteDB := site + "_db"

	mysqlDB := Env.GetString("MYSQL_DATABASE")
	mysqlUser := Env.GetString("MYSQL_USER")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD")
	mysqlRootPassword := Env.GetString("MYSQL_ROOT_PASSWORD")

	commandImport := gunzip + " < " + "production.sql.gz" + " | " + docker + " exec -i " + siteDB + " /usr/bin/mysql --user=root --password=" + mysqlRootPassword + " " + mysqlDB + ""
	logrus.Infof("Run command: %s", commandImport)
	outImport, err := exec.Command("bash", "-c", commandImport).CombinedOutput() //nolint:gosec
	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: string(outImport)})
	}

	if c.Config.FwType == "bitrix" {
		local := Env.GetString("LOCAL_DOMAIN")
		nip := Env.GetString("NIP_DOMAIN")

		strSQL := `"UPDATE b_option SET VALUE = 'Y' WHERE MODULE_ID = 'main' AND NAME = 'update_devsrv';
UPDATE b_lang SET SERVER_NAME='` + site + `.localhost' WHERE LID='s1';
UPDATE b_lang SET b_lang.DOC_ROOT='' WHERE 1=(SELECT DOC_ROOT FROM (SELECT COUNT(LID) FROM b_lang) as cnt);
INSERT INTO b_lang_domain VALUES ('s1', '` + local + `');
INSERT INTO b_lang_domain VALUES ('s1', '` + nip + `');"`

		commandUpdate := "echo " + strSQL + " | " + docker + " exec -i " + siteDB + " /usr/bin/mysql --user=" + mysqlUser + " --password=" + mysqlPassword + " --host=db " + mysqlDB + ""
		logrus.Infof("Run command: %s", commandUpdate)
		outUpdate, err := exec.Command("bash", "-c", commandUpdate).CombinedOutput() //nolint:gosec
		if err != nil {
			w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: string(outUpdate)})
			return
		}
	}

	logrus.Infof("Delete dump: %s", localPath)
	err = exec.Command("rm", localPath).Run()

	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: fmt.Sprint(err)})
	}

	w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Done})
}
