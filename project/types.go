package project

import "github.com/local-deploy/dl/utils/client"

// CallMethod to dynamically determine the function name
type CallMethod struct{}

// SSHClient client ssh
type SSHClient struct {
	*client.Client
}

// DBSettings database settings
type DBSettings struct {
	Host, DataBase, Login, Password, Port string
	ExcludedTables                        []string
}
