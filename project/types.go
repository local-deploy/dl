package project

import "github.com/local-deploy/dl/utils/client"

// CallMethod to dynamically determine the function name
type CallMethod struct{}

// SshClient client ssh
type SshClient struct {
	*client.Client
}

// DbSettings database settings
type DbSettings struct {
	Host, DataBase, Login, Password, Port string
	ExcludedTables                        []string
}
