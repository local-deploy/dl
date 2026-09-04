package project

import (
	"reflect"
	"testing"
)

func TestCertHosts(t *testing.T) {
	tests := []struct {
		name    string
		domains string
		want    []string
	}{
		{
			name:    "Project without DOMAINS keeps the previous SAN set",
			domains: "",
			want:    []string{"mysite.localhost", "mysite.10.0.0.5.nip.io"},
		},
		{
			name:    "SAN covers every project domain",
			domains: "english, russian",
			want: []string{
				"english.localhost", "english.10.0.0.5.nip.io",
				"russian.localhost", "russian.10.0.0.5.nip.io",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domains, err := ParseDomains(tt.domains, "mysite", "/var/www/html", testIP)
			if err != nil {
				t.Fatalf("ParseDomains() unexpected error = %v", err)
			}

			if got := certHosts(domains); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("certHosts() = %v, want %v", got, tt.want)
			}
		})
	}
}
