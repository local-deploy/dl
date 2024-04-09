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
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/client"
	"github.com/sirupsen/logrus"
)

var remotePhpPath string

// DumpDB Database import from server
func DumpDB(ctx context.Context, client *client.Client, tables []string) {
	var err error

	c := &SSHClient{client}
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", Status: progress.Working})

	db, err := c.getMysqlSettings()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprint(err)))
		return
	}

	if len(db.Port) == 0 {
		logrus.Info("Port not set, standard port 3306 is used")
		db.Port = "3306"
	}

	if len(tables) > 0 {
		dumpTables := strings.Join(tables, " ")
		err = c.mysqlDumpTables(ctx, db, dumpTables)
	} else {
		err = c.mysqlDump(ctx, db)
	}

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Failed to create database dump: %s", err)))
		return
	}

	err = c.downloadDump(ctx)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Failed to download dump: %s", err)))
		return
	}

	err = c.ImportDB(ctx)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Database", fmt.Sprintf("Access error: %s", err)))
		return
	}

	w.Event(progress.Event{ID: "Database", Status: progress.Done})
}

func (c SSHClient) getMysqlSettings() (*DBSettings, error) {
	var err error
	var db *DBSettings

	mysqlDataBase := Env.GetString("MYSQL_DATABASE_SRV")
	mysqlLogin := Env.GetString("MYSQL_LOGIN_SRV")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD_SRV")
	if len(mysqlDataBase) > 0 && len(mysqlLogin) > 0 && len(mysqlPassword) > 0 {
		logrus.Info("Manual database access settings are used")
		excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

		db = &DBSettings{
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
			db, err = c.accessBitrixDB()
		case "laravel":
			db, err = c.accessLaravelDB()
		case "wordpress":
			db, err = c.accessWpDB()
		}

		if err != nil {
			return nil, fmt.Errorf("access error: %w", err)
		}
	}

	return db, nil
}

// checkPhpAvailable It possible that PHP not installed on the server in the host system. For example, through docker.
func (c SSHClient) checkPhpAvailable() {
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

// accessBitrixDB Attempt to determine database accesses
func (c SSHClient) accessBitrixDB() (*DBSettings, error) {
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

	dbArray := utils.CleanSlice(strings.Split(strings.TrimSpace(string(cat)), "\n"))
	logrus.Infof("Received variables: %s", dbArray)
	if len(dbArray) != 4 {
		return nil, errors.New("failed to define DB variables, please specify accesses manually")
	}

	excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

	return &DBSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

// accessWpDB Attempt to determine database accesses
func (c SSHClient) accessWpDB() (*DBSettings, error) {
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

	return &DBSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

func (c SSHClient) accessLaravelDB() (*DBSettings, error) {
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

	return &DBSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}, err
}

// mysqlDump Create database dump
func (c SSHClient) mysqlDump(ctx context.Context, db *DBSettings) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", StatusText: "Create database dump"})

	dump := c.checkMySQLDumpAvailable()
	if dump != nil {
		return errors.New("mysqldump not installed, database dump not possible")
	}

	ignoredTablesString := db.FormatIgnoredTables()
	dumpTablesParams := db.DumpTablesParams()
	dumpDataParams := db.DumpDataParams()
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

	return nil
}

// mysqlDumpTables Create only tables dump
func (c SSHClient) mysqlDumpTables(ctx context.Context, db *DBSettings, dumpTables string) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", StatusText: "Creating database dump"})

	dump := c.checkMySQLDumpAvailable()
	if dump != nil {
		return errors.New("mysqldump not installed, database dump not possible")
	}

	dumpDataParams := db.DumpDataTablesParams()
	dumpCmd := strings.Join([]string{"cd", c.Config.Catalog, "&&",
		"mysqldump",
		dumpDataParams,
		db.DataBase,
		dumpTables,
		"|",
		"gzip > " + c.Config.Catalog + "/production.sql.gz",
	}, " ")
	logrus.Infof("Run command: %s", dumpCmd)
	_, err := c.Run(dumpCmd)

	if err != nil {
		return err
	}

	return nil
}

// DumpDataTablesParams options for only tables dump
func (d DBSettings) DumpDataTablesParams() string {
	params := []string{
		"--host=" + d.Host,
		"--port=" + d.Port,
		"--user=" + d.Login,
		"--password=" + strconv.Quote(d.Password),
		"--single-transaction=1",
		"--force",
		"--lock-tables=false",
		"--no-tablespaces",
	}

	mysqlVersion := Env.GetString("MYSQL_VERSION")
	if mysqlVersion == "8.0" {
		params = append(params, "--column-statistics=0")
	}

	return strings.Join(params, " ")
}

func (c SSHClient) checkMySQLDumpAvailable() error {
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

// DumpTablesParams table dump options
func (d DBSettings) DumpTablesParams() string {
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

// DumpDataParams options for data dump
func (d DBSettings) DumpDataParams() string {
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

// FormatIgnoredTables Exclude tables from dump
func (d DBSettings) FormatIgnoredTables() string {
	if len(d.ExcludedTables) == 0 {
		return ""
	}

	ignoredTables := make([]string, len(d.ExcludedTables))
	for i, value := range d.ExcludedTables {
		ignoredTables[i] = "--ignore-table=" + d.DataBase + "." + strings.TrimSpace(value)
	}

	return strings.Join(ignoredTables, " ")
}

// downloadDump Downloading a dump and deleting an archive from the server
func (c SSHClient) downloadDump(ctx context.Context) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", StatusText: "Download database dump"})

	serverPath := filepath.Join(c.Config.Catalog, "production.sql.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")

	logrus.Infof("Download dump: %s", serverPath)
	err := c.Download(ctx, serverPath, localPath)

	if err != nil {
		return err
	}

	err = c.CleanRemote(serverPath)
	if err != nil {
		return err
	}

	return err
}

// ImportDB Importing a database into a local container
func (c SSHClient) ImportDB(ctx context.Context) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", StatusText: "Import database"})

	docker, err := exec.LookPath("docker")
	if err != nil {
		return err
	}
	gunzip, err := exec.LookPath("gunzip")
	if err != nil {
		return err
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
		return errors.New(string(outImport))
	}

	if c.Config.FwType == "bitrix" {
		local := Env.GetString("LOCAL_DOMAIN")
		nip := Env.GetString("NIP_DOMAIN")

		strSQL := `"UPDATE b_option SET VALUE = 'Y' WHERE MODULE_ID = 'main' AND NAME = 'update_devsrv';
UPDATE b_lang SET SERVER_NAME='` + site + `.localhost' WHERE LID='s1';
UPDATE b_lang SET b_lang.DOC_ROOT='' WHERE 1=(SELECT DOC_ROOT FROM (SELECT COUNT(LID) FROM b_lang) as cnt);
INSERT IGNORE INTO b_lang_domain VALUES ('s1', '` + local + `');
INSERT IGNORE INTO b_lang_domain VALUES ('s1', '` + nip + `');"`

		commandUpdate := "echo " + strSQL + " | " + docker + " exec -i " + siteDB + " /usr/bin/mysql --user=" + mysqlUser + " --password=" + mysqlPassword + " --host=db " + mysqlDB + ""
		logrus.Infof("Run command: %s", commandUpdate)
		outUpdate, err := exec.Command("bash", "-c", commandUpdate).CombinedOutput() //nolint:gosec
		if err != nil {
			return errors.New(string(outUpdate))
		}
	}

	logrus.Infof("Delete dump: %s", localPath)
	err = exec.Command("rm", localPath).Run()
	if err != nil {
		return err
	}

	return nil
}
