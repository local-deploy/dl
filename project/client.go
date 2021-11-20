package project

import (
	"bufio"
	"github.com/melbahja/goph"
	"github.com/pterm/pterm"
	"github.com/varrcan/dl/helper"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// SshClient represents ssh client
type SshClient struct {
	*goph.Client
	Server *Server
}

// Server config
type Server struct {
	Server, Key, User, Catalog, FwType string
	Port                               uint
}

//NewClient returns new client and error if any
func NewClient(server *Server) (c *SshClient, err error) {
	home, _ := helper.HomeDir()

	auth, err := goph.Key(filepath.Join(home, ".ssh", server.Key), "")
	if err != nil {
		pterm.FgRed.Println(err)
		return
	}

	c = &SshClient{
		Server: server,
	}

	c.Client, err = goph.NewConn(&goph.Config{
		User:     c.Server.User,
		Addr:     c.Server.Server,
		Port:     c.Server.Port,
		Auth:     auth,
		Callback: verifyHost,
	})
	return
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

	pterm.FgYellow.Printf("The authenticity of host %s can't be established \nFingerprint key: %s \n", host, ssh.FingerprintSHA256(key))
	pterm.FgYellow.Print("Are you sure you want to continue connecting (Y/n)?")

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
