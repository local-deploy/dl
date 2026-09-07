package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestChmodR(t *testing.T) {
	type args struct {
		path string
		mode os.FileMode
	}

	dirName := "chmodTest"
	_ = os.Mkdir(dirName, 0600)

	tests := []struct {
		name string
		args args
	}{
		{name: "Chmod test", args: args{
			path: dirName,
			mode: 0775,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChmodR(tt.args.path, tt.args.mode); err != nil {
				t.Errorf("ChmodR() error = %v", err)
			}
			stat, _ := os.Stat(tt.args.path)
			if perm := stat.Mode(); perm.Perm() != 0775 {
				t.Errorf("ChmodR() error = %v", perm)
			}
		})
		_ = os.Remove(dirName)
	}
}

func TestCleanSlice(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "Clean empty",
			args: []string{"one", "", "two", "three"},
			want: []string{"one", "two", "three"},
		},
		{
			name: "Clean empty",
			args: []string{"", "one", "two", "three"},
			want: []string{"one", "two", "three"},
		},
		{
			name: "Clean empty",
			args: []string{"one", "two", "three", ""},
			want: []string{"one", "two", "three"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanSlice(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CleanSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDir(t *testing.T) {
	defaultDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("user config dir is not available: %v", err)
	}
	want := filepath.Join(defaultDir, "dl")

	tests := []struct {
		name   string
		envSet bool
		env    string
		want   string
	}{
		{
			name:   "Default value",
			envSet: false,
			want:   want,
		},
		{
			name:   "Override",
			envSet: true,
			env:    filepath.Join(os.TempDir(), "dl-isolated"),
			want:   filepath.Join(os.TempDir(), "dl-isolated"),
		},
		{
			name:   "Empty value equals absence",
			envSet: true,
			env:    "",
			want:   want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envSet {
				t.Setenv(ConfigDirEnv, tt.env)
			} else {
				// t.Setenv cannot unset, so restore the value manually
				old, exists := os.LookupEnv(ConfigDirEnv)
				_ = os.Unsetenv(ConfigDirEnv)
				t.Cleanup(func() {
					if exists {
						_ = os.Setenv(ConfigDirEnv, old)
					}
				})
			}

			if got := ConfigDir(); got != tt.want {
				t.Errorf("ConfigDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDirDerivedPaths(t *testing.T) {
	isolated := filepath.Join(os.TempDir(), "dl-isolated")
	t.Setenv(ConfigDirEnv, isolated)

	if got, want := TemplateDir(), filepath.Join(isolated, "templates"); got != want {
		t.Errorf("TemplateDir() = %v, want %v", got, want)
	}
	if got, want := CertDir(), filepath.Join(isolated, "certs"); got != want {
		t.Errorf("CertDir() = %v, want %v", got, want)
	}
}
