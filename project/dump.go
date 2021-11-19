package project

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"github.com/pterm/pterm"
	"github.com/varrcan/dl/helper"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type serverSettings struct {
	Server, Key, User, Catalog string
	Port                       uint
}

type dbSettings struct {
	Host, DataBase, Login, Password string
	ExcludedTables                  []string
}

var sshClient *goph.Client
var server serverSettings
var db dbSettings

//DumpDb Database import from server
func DumpDb() {
	server = getRemote()
	sshClient = newClient()

	// Defer closing the network connection.
	defer func(client *goph.Client) {
		err := client.Close()
		if err != nil {
			pterm.FgRed.Println(err)
			return
		}
	}(sshClient)

	ls := strings.Join([]string{"cd", server.Catalog, "&&", "ls"}, " ")
	out, err := sshClient.Run(ls)
	if err != nil {
		pterm.FgRed.Println(err)
	}

	if strings.Contains(string(out), "bitrix") {
		pterm.FgGreen.Println("Bitrix CMS detected")
		err := accessBitrixDb()
		if err != nil {
			pterm.FgRed.Printfln("Database access error: %w \n", err)
			os.Exit(1)
		}
	}

	err = mysqlDump()
	if err != nil {
		pterm.FgRed.Printfln("Failed to create database dump: %w \n", err)
		os.Exit(1)
	}

	downloadDump()
	importDb()
}

func newClient() (c *goph.Client) {
	home, _ := helper.HomeDir()

	auth, err := goph.Key(filepath.Join(home, ".ssh", server.Key), "")
	if err != nil {
		pterm.FgRed.Println(err)
		return
	}

	c, err = goph.NewConn(&goph.Config{
		User:     server.User,
		Addr:     server.Server,
		Port:     server.Port,
		Auth:     auth,
		Callback: verifyHost,
	})
	return
}

func getRemote() serverSettings {
	return serverSettings{
		Server:  Env.GetString("SERVER"),
		Port:    Env.GetUint("PORT_SRV"),
		User:    Env.GetString("USER_SRV"),
		Key:     Env.GetString("SSH_KEY"),
		Catalog: Env.GetString("CATALOG_SRV"),
	}
}

func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	if hostFound && err == nil {
		return nil
	}

	if askIsHostTrusted(host, key) == false {
		pterm.FgRed.Println("Connection aborted")
		return nil
	}

	return goph.AddKnownHost(host, remote, key, "")
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("The authenticity of host %s can't be established \nFingerprint key: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Are you sure you want to continue connecting (Y/n)?")

	a, err := reader.ReadString('\n')

	if err != nil {
		pterm.FgRed.Println(err)
		return false
	}

	switch strings.ToLower(strings.TrimSpace(a)) {
	case "n":
		return false
	case "y":
	default:
		return true
	}

	return true
}

func accessBitrixDb() error {
	catCmd := strings.Join([]string{"cd", server.Catalog, "&&",
		`cat bitrix/.settings.php | grep "'host' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'database' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'login' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`, "&&",
		`cat bitrix/.settings.php | grep "'password' =>" | awk '{print $3}' | sed -e 's/^.\{1\}//' | sed 's/^\(.*\).$/\1/' | sed 's/^\(.*\).$/\1/'`,
	}, " ")
	cat, err := sshClient.Run(catCmd)

	dbArray := strings.Split(strings.TrimSpace(string(cat)), "\n")

	if len(dbArray) != 4 {
		return errors.New("failed to define variables")
	}

	excludedTables := strings.Split(strings.TrimSpace(Env.GetString("EXCLUDED_TABLES")), ",")

	db = dbSettings{
		Host:           dbArray[0],
		DataBase:       dbArray[1],
		Login:          dbArray[2],
		Password:       dbArray[3],
		ExcludedTables: excludedTables,
	}

	return err
}

func mysqlDump() error {
	pterm.FgGreen.Println("Create database dump")

	ignoredTablesString := formatIgnoredTables()
	dumpCmd := strings.Join([]string{"cd", server.Catalog, "&&",
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
		"gzip > " + server.Catalog + "/production.sql.gz",
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
		"gzip >> " + server.Catalog + "/production.sql.gz",
	}, " ")
	_, err := sshClient.Run(dumpCmd)

	return err
}

func formatIgnoredTables() string {
	var ignoredTables []string

	if len(db.ExcludedTables) == 0 {
		return ""
	}

	for _, value := range db.ExcludedTables {
		ignoredTables = append(ignoredTables, "--ignore-table="+db.DataBase+"."+value)
	}

	return strings.Join(ignoredTables, " ")
}

func downloadDump() {
	serverPath := filepath.Join(server.Catalog, "production.sql.gz")
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")

	err := sshClient.Download(serverPath, localPath)

	if err != nil {
		pterm.FgRed.Println("Download error: ", err)
		os.Exit(1)
	}

	err = cleanRemote(serverPath)
	if err != nil {
		pterm.FgRed.Println("File deletion error: ", err)
	}
}

func importDb() {
	bash, lookErr := exec.LookPath("bash")
	docker, lookErr := exec.LookPath("docker")
	gunzip, lookErr := exec.LookPath("gunzip")
	if lookErr != nil {
		pterm.FgRed.Println(lookErr)
		return
	}

	//TODO: проверить, что контейнер запущен
	localPath := filepath.Join(Env.GetString("PWD"), "production.sql.gz")
	site := Env.GetString("APP_NAME")
	siteDb := site + "_db"

	cmdDump := &exec.Cmd{
		Path:   bash,
		Args:   []string{bash, "-c", gunzip + " < " + localPath + " | " + docker + " exec -i " + siteDb + " /usr/bin/mysql --user=root --password=root db"},
		Env:    CmdEnv(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	strSQL := "\"UPDATE b_option SET VALUE = 'Y' WHERE MODULE_ID = 'main' AND NAME = 'update_devsrv'; UPDATE b_lang SET SERVER_NAME='" + site + "' WHERE LID='s1';\""
	cmdUpdateSite := &exec.Cmd{
		Path: bash,
		Args: []string{bash, "-c", "echo " + strSQL + " | " + docker + " exec -i " + siteDb + " /usr/bin/mysql --user=db --password=db --host=db db"},
		Env:  CmdEnv(),
	}

	pterm.FgGreen.Println("Import database")
	err := cmdDump.Run()

	pterm.FgGreen.Println("Update additional options")
	err = cmdUpdateSite.Run()
	if err != nil {
		pterm.FgRed.Println(err)
	}
}

func cleanRemote(remotePath string) (err error) {
	pterm.FgGreen.Println("Cleaning up temporary files")

	ftp, err := sshClient.NewSftp()
	if err != nil {
		return err
	}

	defer func(ftp *sftp.Client) {
		err := ftp.Close()
		if err != nil {
			pterm.FgRed.Println(err)
		}
	}(ftp)

	err = ftp.Remove(remotePath)

	return err
}
