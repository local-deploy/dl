package github

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v41/github"
)

// Release latest release data
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type Release struct {
	Version, AssetsName, AssetsURL, PageURL string
}

// GetLatestRelease Getting the latest release
func GetLatestRelease(owner, repo string) (r *Release, err error) {
	client := github.NewClient(&http.Client{
		Timeout: 5 * time.Second,
	})

	release, _, err := client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}

	r = &Release{
		Version:    *release.TagName,
		AssetsName: *release.Assets[0].Name,
		AssetsURL:  *release.Assets[0].BrowserDownloadURL,
		PageURL:    *release.HTMLURL,
	}
	return
}
