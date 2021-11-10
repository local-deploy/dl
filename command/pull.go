package command

import (
	"bufio"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"github.com/varrcan/dl/project"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type access struct {
	Server, Key, User, Catalog string
	Port                       uint
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Downloading db and files from the production server",
	Long:  `Downloading database and kernel files from the production server.`,
	Run: func(cmd *cobra.Command, args []string) {
		pull()
	},
}

func pull() {
	project.LoadEnv()

	//client, err := goph.NewConn(&goph.Config{
	//	User:     remote.User,
	//	Addr:     remote.Server,
	//	Port:     remote.Port,
	//	Auth:     auth,
	//	Callback: verifyHost,
	//})

	client, err := newClient()
	if err != nil {
		pterm.FgRed.Println(err)
		return
	}

	// Defer closing the network connection.
	defer func(client *goph.Client) {
		err := client.Close()
		if err != nil {
			pterm.FgRed.Println(err)
			return
		}
	}(client)

	out, err := client.Run("ls")

	if err != nil {
		pterm.FgRed.Println(err)
		return
	}

	fmt.Println(string(out))
}

func newClient() (c *goph.Client, err error) {
	remote := getRemote()
	home, _ := helper.HomeDir()

	auth, err := goph.Key(filepath.Join(home, ".ssh", remote.Key), "")
	if err != nil {
		pterm.FgRed.Println(err)
		return
	}

	//callback, err := goph.DefaultKnownHosts()

	//if err != nil {
	//	pterm.FgRed.Println(err)
	//	return
	//}

	c, err = goph.NewConn(&goph.Config{
		User:     remote.User,
		Addr:     remote.Server,
		Port:     remote.Port,
		Auth:     auth,
		Timeout:  goph.DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	return
}

func getRemote() access {
	return access{
		Server:  project.Env.GetString("SERVER"),
		Port:    project.Env.GetUint("PORT_SRV"),
		User:    project.Env.GetString("USER_SRV"),
		Key:     project.Env.GetString("SSH_KEY"),
		Catalog: project.Env.GetString("CATALOG_SRV"),
	}
}

func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	//
	// If you want to connect to new hosts.
	// here your should check new connections public keys
	// if the key not trusted you shuld return an error
	//

	// hostFound: is host in known hosts file.
	// err: error if key not in known hosts file OR host in known hosts file but key changed!
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	// Host in known hosts but key mismatch!
	// Maybe because of MAN IN THE MIDDLE ATTACK!
	if hostFound && err != nil {

		return err
	}

	// handshake because public key already exists.
	if hostFound && err == nil {

		return nil
	}

	// Ask user to check if he trust the host public key.
	if askIsHostTrusted(host, key) == false {

		pterm.FgRed.Println("aborted!")
		return nil
	}

	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		pterm.FgRed.Println(err)
		return false
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
}
