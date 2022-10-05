package project

import "github.com/varrcan/dl/utils/client"

type callMethod struct{}

type sshClient struct {
	*client.Client
}

type dbSettings struct {
	Host, DataBase, Login, Password, Port string
	ExcludedTables                        []string
}
