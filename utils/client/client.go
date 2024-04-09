package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/local-deploy/dl/utils"
	"github.com/pterm/pterm"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Client represents ssh client
type Client struct {
	*ssh.Client
	Config *Config
}

// Config for Client
type Config struct {
	Auth             Auth
	User             string
	Addr             string
	Key              string
	Port             uint
	Catalog          string
	FwType           string
	UsePassword      bool
	UseKeyPassphrase bool
	Timeout          time.Duration
	Callback         ssh.HostKeyCallback
}

// DefaultTimeout is the timeout of ssh client connection.
var DefaultTimeout = 20 * time.Second

// NewClient returns new client and error if any
func NewClient(config *Config) (c *Client, err error) {
	c, err = NewConn(&Config{
		User:     config.User,
		Addr:     config.Addr,
		Port:     config.Port,
		Catalog:  config.Catalog,
		FwType:   config.FwType,
		Timeout:  DefaultTimeout,
		Auth:     getAuth(config),
		Callback: verifyHost,
	})

	return
}

// NewConn returns new client and error if any.
func NewConn(config *Config) (c *Client, err error) {
	c = &Client{
		Config: config,
	}

	c.Client, err = Dial("tcp", config)
	return
}

// Dial starts a client connection to SSH server based on config.
func Dial(proto string, c *Config) (*ssh.Client, error) {
	return ssh.Dial(proto, net.JoinHostPort(c.Addr, fmt.Sprint(c.Port)), &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         c.Timeout,
		HostKeyCallback: c.Callback,
	})
}

// Run starts a new SSH session and runs the cmd, it returns CombinedOutput and err if any.
func (c Client) Run(cmd string) ([]byte, error) {
	var (
		err  error
		sess *ssh.Session
	)

	if sess, err = c.NewSession(); err != nil {
		return nil, err
	}
	defer func(sess *ssh.Session) {
		err := sess.Close()
		if err != nil {
			return
		}
	}(sess)

	return sess.CombinedOutput(cmd)
}

func getAuth(config *Config) Auth {
	if config.UsePassword {
		auth := Password(askPass("Enter SSH Password: "))

		return auth
	} else {
		home, _ := utils.HomeDir()
		auth, err := Key(filepath.Join(home, ".ssh", config.Key), getPassphrase(config.UseKeyPassphrase))
		if err != nil {
			pterm.FgRed.Println(err)
			return nil
		}
		return auth
	}
}

func getPassphrase(ask bool) string {
	if ask {
		return askPass("Enter Private Key Passphrase: ")
	}
	return ""
}

func askPass(msg string) string {
	fmt.Print(msg)
	pass, err := terminal.ReadPassword(0)
	if err != nil {
		panic(err)
	}
	fmt.Println("")

	return strings.TrimSpace(string(pass))
}

func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	hostFound, err := CheckKnownHost(host, remote, key, "")

	// Host in known hosts but key mismatch
	if hostFound && err != nil {
		return err
	}

	// handshake because public key already exists
	if hostFound && err == nil {
		return nil
	}

	if askIsHostTrusted(host, key) == false {
		pterm.FgRed.Println("Connection aborted")
		return nil
	}
	return AddKnownHost(host, remote, key, "")
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {
	reader := bufio.NewReader(os.Stdin)

	pterm.FgYellow.Printf("The authenticity of host %s can't be established \nFingerprint key: %s \n", host, ssh.FingerprintSHA256(key))
	pterm.FgYellow.Print("Are you sure you want to continue connecting (Y/n)? ")

	a, err := reader.ReadString('\n')
	if err != nil {
		pterm.FgRed.Println(err)
		return false
	}

	a = strings.TrimSpace(a)
	return strings.ToLower(a) == "y" || a == ""
}
