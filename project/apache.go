package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateApacheConfig builds Apache vhost configuration from DomainMappings
func GenerateApacheConfig() string {
	var b strings.Builder

	for i, dm := range DomainMappings {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "<VirtualHost *:80>\n")
		fmt.Fprintf(&b, "    ServerName %s\n", dm.LocalDomain)
		fmt.Fprintf(&b, "    ServerAlias %s\n", dm.NipDomain)
		fmt.Fprintf(&b, "    DocumentRoot %s\n", dm.DocumentRoot)
		fmt.Fprintf(&b, "\n")
		fmt.Fprintf(&b, "    <Directory %s>\n", dm.DocumentRoot)
		fmt.Fprintf(&b, "        AllowOverride All\n")
		fmt.Fprintf(&b, "        Require all granted\n")
		fmt.Fprintf(&b, "    </Directory>\n")
		fmt.Fprintf(&b, "</VirtualHost>\n")
	}

	return b.String()
}

// WriteApacheConfig writes the generated Apache vhost config to .docker/apache/vhosts.conf
func WriteApacheConfig() error {
	pwd := Env.GetString("PWD")
	dir := filepath.Join(pwd, ".docker", "apache")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	config := GenerateApacheConfig()
	configPath := filepath.Join(dir, "vhosts.conf")

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	return nil
}
