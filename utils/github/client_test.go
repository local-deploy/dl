package github

import (
	"reflect"
	"testing"
)

func TestGetLatestRelease(t *testing.T) {
	type args struct {
		owner string
		repo  string
	}
	tests := []struct {
		name    string
		args    args
		wantR   *Release
		wantErr bool
	}{
		{
			name: "Get the latest version",
			args: args{
				owner: "local-deploy",
				repo:  "dl",
			},
			wantR:   &Release{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := GetLatestRelease(tt.args.owner, tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(gotR) != reflect.TypeOf(tt.wantR) {
				t.Errorf("GetLatestRelease() gotR = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}
