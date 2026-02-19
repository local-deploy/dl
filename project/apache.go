package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// GenerateApacheConfig builds Apache vhost configuration from DomainMappings
func GenerateApacheConfig() string {
	var b []byte

	for i, dm := range DomainMappings {
		if i > 0 {
			b = append(b, '\n')
		}
		b = fmt.Appendf(b, "<VirtualHost *:80>\n")
		b = fmt.Appendf(b, "    ServerName %s\n", dm.LocalDomain)
		b = fmt.Appendf(b, "    ServerAlias %s\n", dm.NipDomain)
		b = fmt.Appendf(b, "    DocumentRoot %s\n", dm.DocumentRoot)
		b = fmt.Appendf(b, "\n")
		b = fmt.Appendf(b, "    <Directory %s>\n", dm.DocumentRoot)
		b = fmt.Appendf(b, "        AllowOverride All\n")
		b = fmt.Appendf(b, "        Require all granted\n")
		b = fmt.Appendf(b, "    </Directory>\n")
		b = fmt.Appendf(b, "</VirtualHost>\n")
	}

	return string(b)
}

// WriteApacheConfig writes the generated Apache vhost config to /tmp/dl/<project>/apache/.
// Skips regeneration if .env and project folder haven't changed.
// Returns the absolute path to the config file.
func WriteApacheConfig() (string, error) {
	networkName := Env.GetString("NETWORK_NAME")
	dir := filepath.Join(os.TempDir(), "dl", networkName, "apache")
	confPath := filepath.Join(dir, "vhosts.conf")
	hashPath := filepath.Join(dir, ".confhash")

	if !configNeedsUpdate(hashPath) {
		logrus.Info("Apache config is up to date, skipping regeneration")
		return confPath, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	config := GenerateApacheConfig()

	if err := os.WriteFile(confPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", confPath, err)
	}

	saveConfigHash(hashPath)
	logrus.Info("Apache config regenerated")

	return confPath, nil
}
