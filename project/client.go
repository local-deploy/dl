package project

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"github.com/pterm/pterm"
	"github.com/varrcan/dl/helper"
	"golang.org/x/crypto/ssh"
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

// NewClient returns new client and error if any
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

	if !askIsHostTrusted(host, key) {
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

// cleanRemote Deleting file on the server
func (c SshClient) cleanRemote(remotePath string) (err error) {
	ftp, err := c.NewSftp()
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

// Download file from remote server
func (c SshClient) download(ctx context.Context, remotePath, localPath string) (err error) {
	// w := progress.ContextWriter(ctx)
	local, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer local.Close()

	ftp, err := c.NewSftp()
	if err != nil {
		return
	}
	defer ftp.Close()

	remote, err := ftp.Open(remotePath)
	if err != nil {
		return
	}
	defer remote.Close()

	if _, err = io.Copy(local, remote); err != nil {
		return
	}

	return local.Sync()
}
