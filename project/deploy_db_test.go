package project

import (
	"strings"
	"testing"
)

func TestBitrixDomainsSQL(t *testing.T) {
	t.Run("Every project domain is registered for the main site", func(t *testing.T) {
		domains, err := ParseDomains("omsk, msk", "mysite", "/var/www/html", testIP)
		if err != nil {
			t.Fatalf("ParseDomains() unexpected error = %v", err)
		}

		got := bitrixDomainsSQL(domains)
		for _, want := range []string{
			"'omsk.localhost'",
			"'msk.localhost'",
			"'omsk.10.0.0.5.nip.io'",
			"'msk.10.0.0.5.nip.io'",
			"UPDATE b_lang SET SERVER_NAME='omsk.localhost' WHERE LID = ",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("bitrixDomainsSQL() = %s, want it to contain %s", got, want)
			}
		}

		if strings.Contains(got, "'s1'") {
			t.Errorf("bitrixDomainsSQL() = %s, want the main site selected by DEF instead of the s1 identifier", got)
		}
	})

	t.Run("Project without DOMAINS registers its single domain", func(t *testing.T) {
		domains, err := ParseDomains("", "mysite", "/var/www/html", testIP)
		if err != nil {
			t.Fatalf("ParseDomains() unexpected error = %v", err)
		}

		got := bitrixDomainsSQL(domains)
		if strings.Count(got, "INSERT IGNORE INTO b_lang_domain") != 2 {
			t.Errorf("bitrixDomainsSQL() = %s, want exactly two domain rows", got)
		}
	})

	t.Run("Empty domain list yields no domain statements", func(t *testing.T) {
		got := bitrixDomainsSQL(nil)
		if strings.Contains(got, "b_lang_domain") || strings.Contains(got, "SERVER_NAME") {
			t.Errorf("bitrixDomainsSQL() = %s, want no domain statements", got)
		}
		if !strings.Contains(got, "update_devsrv") {
			t.Errorf("bitrixDomainsSQL() = %s, want the devsrv option still set", got)
		}
	})
}

func TestBitrixMainSiteFallback(t *testing.T) {
	domains, err := ParseDomains("omsk", "mysite", "/var/www/html", testIP)
	if err != nil {
		t.Fatalf("ParseDomains() unexpected error = %v", err)
	}

	got := bitrixDomainsSQL(domains)

	// a dump where no site carries DEF='Y' must still get its domains registered instead
	// of inserting NULL, so the site is picked by sorting rather than by a WHERE filter
	if strings.Contains(got, "WHERE DEF = 'Y'") {
		t.Errorf("bitrixDomainsSQL() = %s, want the main site chosen by ordering, not filtered out", got)
	}
	if !strings.Contains(got, "ORDER BY (DEF = 'Y' AND ACTIVE = 'Y') DESC") {
		t.Errorf("bitrixDomainsSQL() = %s, want the default active site preferred first", got)
	}
	if !strings.Contains(got, "LIMIT 1") {
		t.Errorf("bitrixDomainsSQL() = %s, want exactly one site selected", got)
	}
}
