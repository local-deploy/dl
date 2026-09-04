package project

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update the golden files")

// webserverCases the domain models the web server templates are rendered over
var webserverCases = []struct {
	name    string
	golden  string
	domains []webserverDomain
}{
	{
		name:   "Single domain",
		golden: "single",
		domains: []webserverDomain{
			{DomainMapping: DomainMapping{
				Name: "mysite", DocumentRoot: "/var/www/html",
				LocalDomain: "mysite.localhost", NipDomain: "mysite.10.0.0.5.nip.io",
			}},
		},
	},
	{
		name:   "Shared document root",
		golden: "shared-root",
		domains: []webserverDomain{
			{DomainMapping: DomainMapping{
				Name: "omsk", DocumentRoot: "/var/www/html",
				LocalDomain: "omsk.localhost", NipDomain: "omsk.10.0.0.5.nip.io",
			}},
			{DomainMapping: DomainMapping{
				Name: "msk", DocumentRoot: "/var/www/html",
				LocalDomain: "msk.localhost", NipDomain: "msk.10.0.0.5.nip.io",
			}},
		},
	},
	{
		name:   "Separate document roots",
		golden: "separate-roots",
		domains: []webserverDomain{
			{DomainMapping: DomainMapping{
				Name: "english", DocumentRoot: "/var/www/html/public-en",
				LocalDomain: "english.localhost", NipDomain: "english.10.0.0.5.nip.io",
			}},
			{DomainMapping: DomainMapping{
				Name: "russian", DocumentRoot: "/var/www/html/public-ru",
				LocalDomain: "russian.localhost", NipDomain: "russian.10.0.0.5.nip.io",
			}},
		},
	},
	{
		name:   "Bitrix document root",
		golden: "bitrix",
		domains: []webserverDomain{
			{DomainMapping: DomainMapping{
				Name: "mysite", DocumentRoot: "/var/www/html",
				LocalDomain: "mysite.localhost", NipDomain: "mysite.10.0.0.5.nip.io",
			}, Bitrix: true},
		},
	},
}

func TestRenderNginxConfig(t *testing.T) {
	runGoldenTest(t, nginxTemplateName, "nginx")
}

func TestRenderApacheConfig(t *testing.T) {
	runGoldenTest(t, apacheTemplateName, "apache")
}

func runGoldenTest(t *testing.T, templateName, prefix string) {
	t.Helper()

	tmpl, err := os.ReadFile(filepath.Join("..", "templates", templateName))
	if err != nil {
		t.Fatalf("unable to read the template: %v", err)
	}

	for _, tt := range webserverCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderWebserverConfig(string(tmpl), webserverConfig{
				HostName: "mysite",
				Domains:  tt.domains,
			})
			if err != nil {
				t.Fatalf("renderWebserverConfig() unexpected error = %v", err)
			}

			golden := filepath.Join("testdata", prefix+"-"+tt.golden+".conf")
			if *update {
				if err := os.WriteFile(golden, got, 0600); err != nil {
					t.Fatalf("unable to write the golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("unable to read the golden file (run go test -update to create it): %v", err)
			}

			if string(got) != string(want) {
				t.Errorf("rendered configuration does not match %s:\n--- got ---\n%s\n--- want ---\n%s", golden, got, want)
			}
		})
	}
}

// TestNginxRuntimeVariablesSurvive the change exists to keep the nginx runtime variables
// out of envsubst's reach, so the golden output must still carry them verbatim
func TestNginxRuntimeVariablesSurvive(t *testing.T) {
	got, err := os.ReadFile(filepath.Join("testdata", "nginx-single.conf"))
	if err != nil {
		t.Fatalf("unable to read the golden file: %v", err)
	}

	for _, want := range []string{
		"try_files $uri $uri/ /index.php?$args;",
		"root $root_path;",
		"set $root_path /var/www/html;",
		"fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;",
		"fastcgi_pass mysite_php:9000;",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("the rendered nginx configuration lost %q", want)
		}
	}
}

func TestBitrixInDocumentRoot(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unable to read the working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unable to enter the temporary directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// the project directory maps onto /var/www/html, so "en" here is /var/www/html/en
	if err := os.MkdirAll(filepath.Join("en", "bitrix"), 0755); err != nil {
		t.Fatalf("unable to create the fixture: %v", err)
	}

	tests := []struct {
		name         string
		documentRoot string
		want         bool
	}{
		{name: "Bitrix in a nested document root", documentRoot: "/var/www/html/en", want: true},
		{name: "No bitrix in the project root", documentRoot: "/var/www/html", want: false},
		{name: "Document root outside the mounted tree", documentRoot: "/var/www/site2", want: false},
		{name: "Unrelated absolute path", documentRoot: "/srv/site", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bitrixInDocumentRoot(tt.documentRoot); got != tt.want {
				t.Errorf("bitrixInDocumentRoot(%q) = %v, want %v", tt.documentRoot, got, tt.want)
			}
		})
	}
}

// TestApacheGrantsAccessToDocumentRoot apache2.conf denies everything outside /var/www/,
// so every generated virtual host must grant access to its own document root explicitly
func TestApacheGrantsAccessToDocumentRoot(t *testing.T) {
	got, err := os.ReadFile(filepath.Join("testdata", "apache-separate-roots.conf"))
	if err != nil {
		t.Fatalf("unable to read the golden file: %v", err)
	}

	for _, want := range []string{
		"<Directory /var/www/html/public-en>",
		"<Directory /var/www/html/public-ru>",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("the rendered apache configuration is missing %q", want)
		}
	}

	if strings.Count(string(got), "Require all granted") != 2 {
		t.Errorf("the rendered apache configuration must grant access once per virtual host:\n%s", got)
	}
}
