package project

import (
	"reflect"
	"strings"
	"testing"
)

const testIP = "10.0.0.5"

func TestParseDomains(t *testing.T) {
	tests := []struct {
		name        string
		domains     string
		hostName    string
		defaultRoot string
		want        []DomainMapping
	}{
		{
			name:        "No DOMAINS falls back to HOST_NAME",
			domains:     "",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "mysite", DocumentRoot: "/var/www/html", LocalDomain: "mysite.localhost", NipDomain: "mysite.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Blank DOMAINS equals absence",
			domains:     "   ",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "mysite", DocumentRoot: "/var/www/html", LocalDomain: "mysite.localhost", NipDomain: "mysite.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "HOST_NAME is lowercased in the fallback",
			domains:     "",
			hostName:    "MySite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "mysite", DocumentRoot: "/var/www/html", LocalDomain: "mysite.localhost", NipDomain: "mysite.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Shared document root",
			domains:     "omsk, msk",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "omsk", DocumentRoot: "/var/www/html", LocalDomain: "omsk.localhost", NipDomain: "omsk.10.0.0.5.nip.io"},
				{Name: "msk", DocumentRoot: "/var/www/html", LocalDomain: "msk.localhost", NipDomain: "msk.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Separate document roots",
			domains:     "english:/var/www/html/public-en, russian:/var/www/html/public-ru",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "english", DocumentRoot: "/var/www/html/public-en", LocalDomain: "english.localhost", NipDomain: "english.10.0.0.5.nip.io"},
				{Name: "russian", DocumentRoot: "/var/www/html/public-ru", LocalDomain: "russian.localhost", NipDomain: "russian.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Mixed list",
			domains:     "omsk, msk, en:/var/www/html/en",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "omsk", DocumentRoot: "/var/www/html", LocalDomain: "omsk.localhost", NipDomain: "omsk.10.0.0.5.nip.io"},
				{Name: "msk", DocumentRoot: "/var/www/html", LocalDomain: "msk.localhost", NipDomain: "msk.10.0.0.5.nip.io"},
				{Name: "en", DocumentRoot: "/var/www/html/en", LocalDomain: "en.localhost", NipDomain: "en.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Single domain with its own document root",
			domains:     "site:/var/www/html/public",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "site", DocumentRoot: "/var/www/html/public", LocalDomain: "site.localhost", NipDomain: "site.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Whitespace and case are normalized",
			domains:     "  Omsk ,  MSK : /var/www/html  ",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "omsk", DocumentRoot: "/var/www/html", LocalDomain: "omsk.localhost", NipDomain: "omsk.10.0.0.5.nip.io"},
				{Name: "msk", DocumentRoot: "/var/www/html", LocalDomain: "msk.localhost", NipDomain: "msk.10.0.0.5.nip.io"},
			},
		},
		{
			name:        "Dotted and hyphenated names are allowed",
			domains:     "my-site.dev",
			hostName:    "mysite",
			defaultRoot: "/var/www/html",
			want: []DomainMapping{
				{Name: "my-site.dev", DocumentRoot: "/var/www/html", LocalDomain: "my-site.dev.localhost", NipDomain: "my-site.dev.10.0.0.5.nip.io"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDomains(tt.domains, tt.hostName, tt.defaultRoot, testIP)
			if err != nil {
				t.Fatalf("ParseDomains() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDomains() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseDomainsValidation(t *testing.T) {
	tests := []struct {
		name     string
		domains  string
		hostName string
		wantErr  string
	}{
		{
			name:     "Empty domain name",
			domains:  "english,,russian",
			hostName: "mysite",
			wantErr:  "empty domain name",
		},
		{
			name:     "Relative document root",
			domains:  "english:public-en",
			hostName: "mysite",
			wantErr:  "must be an absolute path",
		},
		{
			name:     "Empty document root after the colon",
			domains:  "english:",
			hostName: "mysite",
			wantErr:  "empty document root for domain \"english\"",
		},
		{
			name:     "Invalid characters in the domain name",
			domains:  "eng lish",
			hostName: "mysite",
			wantErr:  "invalid domain name",
		},
		{
			name:     "Domain name starting with a hyphen",
			domains:  "-english",
			hostName: "mysite",
			wantErr:  "invalid domain name",
		},
		{
			name:     "Duplicate domains",
			domains:  "english, english",
			hostName: "mysite",
			wantErr:  "duplicate domain name \"english\"",
		},
		{
			name:     "Duplicate domains differing in case",
			domains:  "English, english",
			hostName: "mysite",
			wantErr:  "duplicate domain name \"english\"",
		},
		{
			name:     "Undefined HOST_NAME without DOMAINS",
			domains:  "",
			hostName: "",
			wantErr:  "HOST_NAME variable is not defined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDomains(tt.domains, tt.hostName, "/var/www/html", testIP)
			if err == nil {
				t.Fatalf("ParseDomains() expected an error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseDomains() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestTraefikRule(t *testing.T) {
	t.Run("Matches the rule used before the change", func(t *testing.T) {
		domains, err := ParseDomains("", "mysite", "/var/www/html", testIP)
		if err != nil {
			t.Fatalf("ParseDomains() unexpected error = %v", err)
		}

		// the rule inlined in the compose templates before the change, with the YAML
		// escaping of the backslashes undone
		want := "Host(`mysite.localhost`) || " +
			"HostRegexp(`^.+\\.mysite\\.localhost$`) || " +
			"HostRegexp(`^mysite\\..+\\.nip\\.io$`) || " +
			"HostRegexp(`^.+\\.mysite\\..+\\.nip\\.io$`)"

		if got := TraefikRule(domains); got != want {
			t.Errorf("TraefikRule() = %s, want %s", got, want)
		}
	})

	t.Run("Covers every domain of the project", func(t *testing.T) {
		domains, err := ParseDomains("english, russian", "mysite", "/var/www/html", testIP)
		if err != nil {
			t.Fatalf("ParseDomains() unexpected error = %v", err)
		}

		got := TraefikRule(domains)
		for _, want := range []string{
			"Host(`english.localhost`)",
			"Host(`russian.localhost`)",
			"HostRegexp(`^.+\\.english\\.localhost$`)",
			"HostRegexp(`^english\\..+\\.nip\\.io$`)",
			"HostRegexp(`^.+\\.russian\\..+\\.nip\\.io$`)",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("TraefikRule() = %s, want it to contain %s", got, want)
			}
		}
	})

	t.Run("Dots in the domain name are escaped", func(t *testing.T) {
		domains, err := ParseDomains("my.site", "mysite", "/var/www/html", testIP)
		if err != nil {
			t.Fatalf("ParseDomains() unexpected error = %v", err)
		}

		got := TraefikRule(domains)
		if !strings.Contains(got, "HostRegexp(`^.+\\.my\\.site\\.localhost$`)") {
			t.Errorf("TraefikRule() = %s, want the dot in the domain name escaped", got)
		}
	})
}
