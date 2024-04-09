package utils

import (
	"os"
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
