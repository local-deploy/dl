package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/local-deploy/dl/utils"
	"github.com/pterm/pterm"
)

const (
	nginxTemplateName  = "nginx.conf.gotmpl"
	apacheTemplateName = "apache-vhosts.conf.gotmpl"

	nginxConfigName  = "nginx.conf"
	apacheConfigName = "apache-vhosts.conf"
)

// webserverDomain a domain as the web server templates see it
type webserverDomain struct {
	DomainMapping
	// Bitrix whether the document root of this domain holds a bitrix directory
	Bitrix bool
}

// webserverConfig the data passed to the web server templates
type webserverConfig struct {
	HostName string
	Domains  []webserverDomain
}

// renderWebserverConfig render a web server configuration template over the domain model
func renderWebserverConfig(tmpl string, config webserverConfig) ([]byte, error) {
	t, err := template.New("webserver").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, config); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// newWebserverConfig build the template data from the project domain model
func newWebserverConfig(hostName string, domains []DomainMapping) webserverConfig {
	config := webserverConfig{
		HostName: hostName,
		Domains:  make([]webserverDomain, 0, len(domains)),
	}

	for _, domain := range domains {
		config.Domains = append(config.Domains, webserverDomain{
			DomainMapping: domain,
			Bitrix:        utils.BitrixCheck(domain.DocumentRoot),
		})
	}

	return config
}

// generateWebserverConfig render the template and write the result into the project
// subdirectory of the dl configuration directory, next to the project certificates
func generateWebserverConfig(templateName, configName, hostName, networkName string, domains []DomainMapping) (string, error) {
	tmpl, err := os.ReadFile(filepath.Join(utils.TemplateDir(), templateName))
	if err != nil {
		return "", fmt.Errorf("unable to read the %s template: %w", templateName, err)
	}

	data, err := renderWebserverConfig(string(tmpl), newWebserverConfig(hostName, domains))
	if err != nil {
		return "", fmt.Errorf("unable to render the %s template: %w", templateName, err)
	}

	confDir := filepath.Join(utils.ConfigDir(), "conf", networkName)
	if err := utils.CreateDirectory(confDir); err != nil {
		return "", fmt.Errorf("unable to create the %s directory: %w", confDir, err)
	}

	confFile := filepath.Join(confDir, configName)
	if err := os.WriteFile(confFile, data, 0644); err != nil { //nolint:gosec
		return "", fmt.Errorf("unable to write %s: %w", confFile, err)
	}

	return confFile, nil
}

// setWebserverConfig generate the web server configuration for the current project and
// point NGINX_CONF or APACHE_CONF at it. A user-supplied NGINX_CONF keeps its priority:
// when it is set in .env, nothing is generated and the file the user named is mounted.
func setWebserverConfig() {
	phpVersion := Env.GetString("PHP_VERSION")
	hostName := Env.GetString("HOST_NAME")
	networkName := Env.GetString("NETWORK_NAME")

	switch {
	case strings.Contains(phpVersion, "apache"):
		conf, err := generateWebserverConfig(apacheTemplateName, apacheConfigName, hostName, networkName, Domains)
		if err != nil {
			pterm.FgRed.Printfln("Error: %s", err)
			os.Exit(1)
		}
		Env.Set("APACHE_CONF", conf)

	case strings.Contains(phpVersion, "fpm"):
		// A user-supplied NGINX_CONF, resolved in setDefaultEnv, suppresses the generation.
		// Such a file is written as a .conf.template expecting the nginx image to expand
		// ${HOST_NAME}, ${VIRTUAL_HOST} and ${DOCUMENT_ROOT} through envsubst, so it keeps
		// the mount point it had before the change — the generated configuration is the
		// only one that goes into conf.d ready to use.
		if len(Env.GetString("NGINX_CONF")) > 0 {
			Env.Set("NGINX_CONF_TARGET", "/etc/nginx/templates/default.conf.template")
			return
		}

		conf, err := generateWebserverConfig(nginxTemplateName, nginxConfigName, hostName, networkName, Domains)
		if err != nil {
			pterm.FgRed.Printfln("Error: %s", err)
			os.Exit(1)
		}
		Env.Set("NGINX_CONF", conf)
	}
}
