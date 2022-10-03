package disk

import (
	"reflect"
	"testing"
)

func TestFreeSpaceHome(t *testing.T) {
	tests := []struct {
		name  string
		wantD *Disk
	}{
		{name: "Free space in home directory", wantD: &Disk{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotD := FreeSpaceHome(); reflect.TypeOf(gotD) != reflect.TypeOf(tt.wantD) {
				t.Errorf("FreeSpaceHome() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func TestHumanSize(t *testing.T) {
	type args struct {
		size float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Convert bytes to KB", args: args{size: 46164}, want: "45.1 KB"},
		{name: "Convert bytes to MB", args: args{size: 5987456}, want: "5.7 MB"},
		{name: "Convert bytes to GB", args: args{size: 3459834587}, want: "3.2 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanSize(tt.args.size); got != tt.want {
				t.Errorf("HumanSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
