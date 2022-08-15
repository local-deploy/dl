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
)

type dbSettings struct {
	Host, DataBase, Login, Password, Port string
	ExcludedTables                        []string
}

// DumpDb Database import from server
func (c SshClient) DumpDb(ctx context.Context) {
	var db *dbSettings
	var err error

	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Database", Status: progress.Working})

	mysqlDataBase := Env.GetString("MYSQL_DATABASE_SRV")
	mysqlLogin := Env.GetString("MYSQL_LOGIN_SRV")
	mysqlPassword := Env.GetString("MYSQL_PASSWORD_SRV")
	if len(mysqlDataBase) > 0 && len(mysqlLogin) > 0 && len(mysqlPassword) > 0 {
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
		switch c.Server.FwType {
		case "bitrix":
			db, err = c.accessBitrixDb()
		case "laravel":
			db, err = c.accessLaravelDb()
		case "wordpress": // TODO
		}

		if err != nil {
			w.Event(progress.ErrorMessageEvent("Database access error", fmt.Sprint(err)))
			return
		}
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

// accessBitrixDb Attempt to determine database accesses
func (c SshClient) accessBitrixDb() (*dbSettings, error) {
	catCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&",
		`$(which php) -r '$settings = include "bitrix/.settings.php"; echo $settings["connections"]["value"]["default"]["host"]."\n";
echo $settings["connections"]["value"]["default"]["database"]."\n";
echo $settings["connections"]["value"]["default"]["login"]."\n";
echo $settings["connections"]["value"]["default"]["password"]."\n";'`,
	}, " ")
	cat, err := c.Run(catCmd)
	if err != nil {
		return nil, err
	}

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")
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

func (c SshClient) accessLaravelDb() (*dbSettings, error) {
	catCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&", "export $(grep -v '^#' .env | xargs)", "&&",
		`echo $DB_HOST`, "&&",
		`echo $DB_DATABASE`, "&&",
		`echo $DB_USERNAME`, "&&",
		`echo $DB_PASSWORD`,
	}, " ")
	cat, err := c.Run(catCmd)

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")

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
func (c SshClient) mysqlDump(ctx context.Context, db *dbSettings) error {
	w := progress.ContextWriter(ctx)
	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Working})

	port := db.Port
	if len(port) == 0 {
		port = "3306"
	}

	ignoredTablesString := db.formatIgnoredTables()
	dumpCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&",
		"mysqldump",
		"--host=" + db.Host,
		"--port=" + port,
		"--user=" + db.Login,
		"--password=" + strconv.Quote(db.Password),
		"--single-transaction=1",
		"--lock-tables=false",
		"--no-data",
		"--no-tablespaces",
		db.DataBase,
		"|",
		"gzip > " + c.Server.Catalog + "/production.sql.gz",
		"&&",
		"mysqldump",
		"--host=" + db.Host,
		"--port=" + port,
		"--user=" + db.Login,
		"--password=" + strconv.Quote(db.Password),
		"--single-transaction=1",
		"--force",
		"--lock-tables=false",
		"--no-tablespaces",
		"--no-create-info",
		ignoredTablesString,
		db.DataBase,
		"|",
		"gzip >> " + c.Server.Catalog + "/production.sql.gz",
	}, " ")
	_, err := c.Run(dumpCmd)

	if err != nil {
		return err
	}

	w.Event(progress.Event{ID: "Create database dump", ParentID: "Database", Status: progress.Done})

	return nil
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
func (c SshClient) downloadDump(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Working})

	serverPath := filepath.Join(c.Server.Catalog, "production.sql.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")

	err := c.download(ctx, serverPath, localPath)

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Download error", fmt.Sprint(err)))
		return err
	}

	err = c.cleanRemote(serverPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("File deletion error", fmt.Sprint(err)))
		return err
	}

	w.Event(progress.Event{ID: "Download database dump", ParentID: "Database", Status: progress.Done})

	return err
}

// importDb Importing a database into a local container
func (c SshClient) importDb(ctx context.Context) {
	var err error

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

	outImport, err := exec.Command("bash", "-c", gunzip+" < "+localPath+" | "+docker+" exec -i "+siteDB+" /usr/bin/mysql --user=root --password="+mysqlRootPassword+" "+mysqlDB+"").CombinedOutput() //nolint:gosec
	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: string(outImport)})
		return
	}

	if c.Server.FwType == "bitrix" {
		local := Env.GetString("LOCAL_DOMAIN")
		nip := Env.GetString("NIP_DOMAIN")

		strSQL := `"UPDATE b_option SET VALUE = 'Y' WHERE MODULE_ID = 'main' AND NAME = 'update_devsrv';
UPDATE b_lang SET SERVER_NAME='` + site + `.localhost' WHERE LID='s1';
UPDATE b_lang SET b_lang.DOC_ROOT='' WHERE 1=(SELECT DOC_ROOT FROM (SELECT COUNT(LID) FROM b_lang) as cnt);
INSERT INTO b_lang_domain VALUES ('s1', '` + local + `');
INSERT INTO b_lang_domain VALUES ('s1', '` + nip + `');"`

		outUpdate, err := exec.Command("bash", "-c", "echo "+strSQL+" | "+docker+" exec -i "+siteDB+" /usr/bin/mysql --user="+mysqlUser+" --password="+mysqlPassword+" --host=db "+mysqlDB+"").CombinedOutput() //nolint:gosec
		if err != nil {
			w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: string(outUpdate)})
			return
		}
	}

	err = exec.Command("rm", localPath).Run()

	if err != nil {
		w.Event(progress.Event{ID: "Import database", ParentID: "Database", Status: progress.Error, Text: fmt.Sprint(err)})
	}

	w.Event(progress.Event{
		ID:       "Import database",
		ParentID: "Database",
		Status:   progress.Done,
	})
}
