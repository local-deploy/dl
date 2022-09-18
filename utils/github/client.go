package github

import (
	"context"

	"github.com/google/go-github/v41/github"
)

// Release latest release data
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type Release struct {
	Version, AssetsName, AssetsUrl, PageUrl string
}

// GetLatestRelease Getting the latest release
func GetLatestRelease(owner, repo string) (r *Release, err error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}

	r = &Release{
		Version:    *release.TagName,
		AssetsName: *release.Assets[0].Name,
		AssetsUrl:  *release.Assets[0].BrowserDownloadURL,
		PageUrl:    *release.HTMLURL,
	}
	return
}
