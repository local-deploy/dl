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

// DomainMapping represents a domain-to-document-root mapping
type DomainMapping struct {
	Name         string // domain name (e.g., "site1")
	DocumentRoot string // absolute path (e.g., "/var/www/html/site1")
	LocalDomain  string // auto-generated: "site1.localhost"
	NipDomain    string // auto-generated: "site1.192.168.1.100.nip.io"
}
