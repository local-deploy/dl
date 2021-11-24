package project

import (
	"errors"
	"github.com/pterm/pterm"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type dbSettings struct {
	Host, DataBase, Login, Password string
	ExcludedTables                  []string
}

//DumpDb Database import from server
func (c SshClient) DumpDb() {
	var db *dbSettings
	var err error

	switch c.Server.FwType {
	case "bitrix":
		db, err = c.accessBitrixDb()
	case "laravel":
		db, err = c.accessLaravelDb()
	case "wordpress": //TODO
	}

	if err != nil {
		pterm.FgRed.Printfln("Database access error: %w \n", err)
		os.Exit(1)
	}

	err = c.mysqlDump(db)
	if err != nil {
		pterm.FgRed.Printfln("Failed to create database dump: %w \n", err)
		os.Exit(1)
	}

	c.downloadDump()
	c.importDb()
}

//accessBitrixDb Attempt to determine database accesses
func (c SshClient) accessBitrixDb() (*dbSettings, error) {
	catCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&",
		`cat bitrix/.settings.php | grep "'host' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'database' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'login' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'password' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`,
	}, " ")
	cat, err := c.Run(catCmd)

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")

	if len(dbArray) != 4 {
		return nil, errors.New("failed to define variables")
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
		return nil, errors.New("failed to define variables")
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

//mysqlDump Create database dump
func (c SshClient) mysqlDump(db *dbSettings) error {
	pterm.FgGreen.Println("Create database dump")

	ignoredTablesString := db.formatIgnoredTables()
	dumpCmd := strings.Join([]string{"cd", c.Server.Catalog, "&&",
		"mysqldump",
		"--host=" + db.Host,
		"--user=" + db.Login,
		"--password=" + db.Password,
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
		"--user=" + db.Login,
		"--password=" + db.Password,
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

	return err
}

//formatIgnoredTables Exclude tables from dump
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

//downloadDump Downloading a dump and deleting an archive from the server
func (c SshClient) downloadDump() {
	pterm.FgGreen.Println("Download database dump")
	serverPath := filepath.Join(c.Server.Catalog, "production.sql.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")

	err := c.download(serverPath, localPath)

	if err != nil {
		pterm.FgRed.Println("Download error: ", err)
		os.Exit(1)
	}

	err = c.cleanRemote(serverPath)
	if err != nil {
		pterm.FgRed.Println("File deletion error: ", err)
	}
}

//importDb Importing a database into a local container
func (c SshClient) importDb() {
	var err error

	pterm.FgGreen.Println("Import database")

	docker, lookErr := exec.LookPath("docker")
	gunzip, lookErr := exec.LookPath("gunzip")
	if lookErr != nil {
		pterm.FgRed.Println(lookErr)
		return
	}

	//TODO: переписать на sdk

	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")
	site := Env.GetString("APP_NAME")
	siteDb := site + "_db"

	err = exec.Command("bash", "-c", gunzip+" < "+localPath+" | "+docker+" exec -i "+siteDb+" /usr/bin/mysql --user=root --password=root db").Run()
	if err != nil {
		pterm.FgRed.Println(err)
	}

	if c.Server.FwType == "bitrix" {
		local := Env.GetString("LOCAL_DOMAIN")
		nip := Env.GetString("NIP_DOMAIN")

		strSQL := `"UPDATE b_option SET VALUE = 'Y' WHERE MODULE_ID = 'main' AND NAME = 'update_devsrv'; 
UPDATE b_lang SET SERVER_NAME='` + site + `.localhost' WHERE LID='s1'; 
UPDATE b_lang SET b_lang.DOC_ROOT='' WHERE 1=(SELECT DOC_ROOT FROM (SELECT COUNT(LID) FROM b_lang) as cnt); 
INSERT INTO b_lang_domain VALUES ('s1', '` + local + `'); 
INSERT INTO b_lang_domain VALUES ('s1', '` + nip + `');"`

		err = exec.Command("bash", "-c", "echo "+strSQL+" | "+docker+" exec -i "+siteDb+" /usr/bin/mysql --user=db --password=db --host=db db").Run()
		if err != nil {
			pterm.FgRed.Println(err)
		}
	}

	err = exec.Command("rm", localPath).Run()

	if err != nil {
		pterm.FgRed.Println(err)
	}
}
