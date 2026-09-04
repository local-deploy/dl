package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
)

// domainNameRe valid domain name: letters, digits, hyphens and dots, starts with a letter or a digit
var domainNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)

// setDomains fill the project domain model and the variables derived from it.
// LOCAL_DOMAIN and NIP_DOMAIN keep pointing at the main domain — the first entry of the list.
func setDomains(projectName, localIP string) {
	domains, err := ParseDomains(Env.GetString("DOMAINS"), projectName, Env.GetString("DOCUMENT_ROOT"), localIP)
	if err != nil {
		pterm.FgRed.Printfln("Error in the DOMAINS variable: %s", err)
		os.Exit(1)
	}

	Domains = domains

	Env.SetDefault("LOCAL_DOMAIN", domains[0].LocalDomain)
	Env.SetDefault("NIP_DOMAIN", domains[0].NipDomain)
	Env.SetDefault("TRAEFIK_RULE", TraefikRule(domains))
}

// ParseDomains parse the DOMAINS variable into the project domain model.
// Every entry is either "<name>" or "<name>:<document-root>"; entries without an explicit
// document root inherit defaultRoot. An empty domains value yields a single entry built
// from hostName and defaultRoot, which keeps projects without DOMAINS working as before.
func ParseDomains(domains, hostName, defaultRoot, localIP string) ([]DomainMapping, error) {
	if len(strings.TrimSpace(domains)) == 0 {
		name := strings.ToLower(strings.TrimSpace(hostName))
		if len(name) == 0 {
			return nil, fmt.Errorf("the HOST_NAME variable is not defined")
		}

		return []DomainMapping{newDomainMapping(name, defaultRoot, localIP)}, nil
	}

	entries := strings.Split(domains, ",")
	mappings := make([]DomainMapping, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		name, root, hasRoot := strings.Cut(entry, ":")

		name = strings.ToLower(strings.TrimSpace(name))
		if len(name) == 0 {
			return nil, fmt.Errorf("empty domain name in the DOMAINS variable")
		}
		if !domainNameRe.MatchString(name) {
			return nil, fmt.Errorf("invalid domain name %q: only letters, digits, hyphens and dots are allowed, "+
				"the name must start with a letter or a digit", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, fmt.Errorf("duplicate domain name %q in the DOMAINS variable", name)
		}
		seen[name] = struct{}{}

		if hasRoot {
			root = strings.TrimSpace(root)
			if len(root) == 0 {
				return nil, fmt.Errorf("empty document root for domain %q", name)
			}
			if !filepath.IsAbs(root) {
				return nil, fmt.Errorf("document root %q for domain %q must be an absolute path", root, name)
			}
		} else {
			root = defaultRoot
		}

		mappings = append(mappings, newDomainMapping(name, root, localIP))
	}

	return mappings, nil
}

// newDomainMapping build a domain entry with the derived local and nip.io domains
func newDomainMapping(name, root, localIP string) DomainMapping {
	return DomainMapping{
		Name:         name,
		DocumentRoot: root,
		LocalDomain:  fmt.Sprintf("%s.localhost", name),
		NipDomain:    fmt.Sprintf("%s.%s.nip.io", name, localIP),
	}
}

// TraefikRule build the traefik routing rule covering every domain of the project:
// the local domain and its subdomains, the nip.io domain and its subdomains.
//
// The rule is passed to compose as the value of TRAEFIK_RULE, so backslashes are written
// once here: unlike the labels in the compose templates, the value is not re-parsed as YAML.
func TraefikRule(domains []DomainMapping) string {
	rules := make([]string, 0, len(domains)*4)

	for _, domain := range domains {
		name := regexp.QuoteMeta(domain.Name)
		rules = append(rules,
			fmt.Sprintf("Host(`%s`)", domain.LocalDomain),
			fmt.Sprintf("HostRegexp(`^.+\\.%s\\.localhost$`)", name),
			fmt.Sprintf("HostRegexp(`^%s\\..+\\.nip\\.io$`)", name),
			fmt.Sprintf("HostRegexp(`^.+\\.%s\\..+\\.nip\\.io$`)", name),
		)
	}

	return strings.Join(rules, " || ")
}
