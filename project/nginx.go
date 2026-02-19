package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/local-deploy/dl/utils"
	"github.com/sirupsen/logrus"
)

// configNeedsUpdate checks whether the generated web server config is stale
// by comparing a SHA-256 hash of .env contents + project folder name against a stored hash.
func configNeedsUpdate(hashFilePath string) bool {
	pwd := Env.GetString("PWD")
	envPath := filepath.Join(pwd, ".env")

	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return true
	}

	h := sha256.New()
	h.Write(envContent)
	h.Write([]byte(filepath.Base(pwd)))
	currentHash := hex.EncodeToString(h.Sum(nil))

	storedHash, err := os.ReadFile(hashFilePath)
	if err != nil {
		return true
	}

	return strings.TrimSpace(string(storedHash)) != currentHash
}

// saveConfigHash persists the current hash so subsequent runs can skip regeneration.
func saveConfigHash(hashFilePath string) {
	pwd := Env.GetString("PWD")
	envPath := filepath.Join(pwd, ".env")

	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return
	}

	h := sha256.New()
	h.Write(envContent)
	h.Write([]byte(filepath.Base(pwd)))
	currentHash := hex.EncodeToString(h.Sum(nil))

	_ = os.WriteFile(hashFilePath, []byte(currentHash), 0644)
}

// GenerateNginxConfig builds a complete nginx config from DomainMappings.
// One server block is generated per DomainMapping entry.
func GenerateNginxConfig() string {
	hostName := Env.GetString("HOST_NAME")
	var b strings.Builder

	for i, mapping := range DomainMappings {
		if i > 0 {
			b.WriteString("\n")
		}

		tryFiles := "try_files $uri $uri/ /index.php?$args;"
		if utils.BitrixCheck(mapping.DocumentRoot) {
			tryFiles = "try_files $uri $uri/ /index.php?$args /bitrix/urlrewrite.php?$args /bitrix/routing_index.php?$args;"
		}

		b.WriteString("server {\n")
		b.WriteString("    listen 80;\n")
		b.WriteString("    listen 443;\n")
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("    server_name %s %s;\n", mapping.LocalDomain, mapping.NipDomain))
		b.WriteString("    add_header Strict-Transport-Security \"max-age=31536000\" always;\n")
		b.WriteString("    client_max_body_size 200M;\n")
		b.WriteString("\n")
		b.WriteString("    charset utf-8;\n")
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("    set $root_path %s;\n", mapping.DocumentRoot))
		b.WriteString("    root $root_path;\n")
		b.WriteString("\n")
		b.WriteString("    location / {\n")
		b.WriteString("        root $root_path;\n")
		b.WriteString("        index index.php index.html;\n")
		b.WriteString(fmt.Sprintf("        %s\n", tryFiles))
		b.WriteString("    }\n")
		b.WriteString("\n")
		b.WriteString("    location ~ \\.php$ {\n")
		b.WriteString(fmt.Sprintf("        fastcgi_pass %s_php:9000;\n", hostName))
		b.WriteString("        fastcgi_index index.php;\n")
		b.WriteString("        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;\n")
		b.WriteString("        include /etc/nginx/fastcgi_params;\n")
		b.WriteString("     }\n")
		b.WriteString("\n")
		b.WriteString("    location ~* ^.+\\.(jpg|jpeg|gif|png|svg|js|css|mp3|ogg|mpeg|avi|zip|gz|bz2|rar|swf|ico|7z|doc|docx|map|ogg|otf|pdf|tff|tif|txt|wav|webp|woff|woff2|xls|xlsx|xml)$ {\n")
		b.WriteString("        expires 365d;\n")
		b.WriteString("        try_files $uri $uri/ 404 = @fallback;\n")
		b.WriteString("    }\n")
		b.WriteString("\n")
		b.WriteString("    location @fallback {\n")
		b.WriteString(fmt.Sprintf("        return 302 https://%s/$uri;\n", mapping.LocalDomain))
		b.WriteString("    }\n")
		b.WriteString("}\n")
	}

	return b.String()
}

// WriteNginxConfig writes the generated nginx config to /tmp/dl/<project>/nginx/.
// Skips regeneration if .env and project folder haven't changed.
// Returns the absolute path to the config file.
func WriteNginxConfig() (string, error) {
	networkName := Env.GetString("NETWORK_NAME")
	dir := filepath.Join(os.TempDir(), "dl", networkName, "nginx")
	confPath := filepath.Join(dir, "default.conf")
	hashPath := filepath.Join(dir, ".confhash")

	if !configNeedsUpdate(hashPath) {
		logrus.Info("Nginx config is up to date, skipping regeneration")
		return confPath, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create nginx config directory: %w", err)
	}

	config := GenerateNginxConfig()

	if err := os.WriteFile(confPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write nginx config: %w", err)
	}

	saveConfigHash(hashPath)
	logrus.Info("Nginx config regenerated")

	return confPath, nil
}
