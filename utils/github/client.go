package github

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/google/go-github/v41/github"
)

// Release latest release data
type Release struct {
	Version, AssetsName, AssetsURL, PageURL string
}

// GetRelease Getting the specified release
func GetRelease(owner, repo string, tag string) (r *Release, err error) {
	client := github.NewClient(&http.Client{
		Timeout: 5 * time.Second,
	})

	var release *github.RepositoryRelease

	if len(tag) > 0 {
		release, _, err = client.Repositories.GetReleaseByTag(context.Background(), owner, repo, tag)
	} else {
		release, _, err = client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	}
	if err != nil {
		return nil, err
	}

	assetIndex, err := getAssetIndex(release)
	if err != nil {
		return nil, err
	}

	r = &Release{
		Version:    *release.TagName,
		AssetsName: *release.Assets[assetIndex].Name,
		AssetsURL:  *release.Assets[assetIndex].BrowserDownloadURL,
		PageURL:    *release.HTMLURL,
	}
	return
}

func getAssetIndex(release *github.RepositoryRelease) (int, error) {
	system, err := getSystem()
	if err != nil {
		return 0, err
	}
	arch, err := getArch()
	if err != nil {
		return 0, err
	}
	binName := fmt.Sprintf("dl-%s-%s-%s.tar.gz", *release.TagName, system, arch)
	for i, asset := range release.Assets {
		if binName == *asset.Name {
			return i, nil
		}
	}

	return 0, fmt.Errorf("error getting archive from release %s", *release.HTMLURL)
}

func getSystem() (string, error) {
	system := runtime.GOOS

	switch system {
	case "linux", "darwin":
		return system, nil
	}
	return "", fmt.Errorf("this installer does not support %s platform at this time", system)
}

func getArch() (string, error) {
	arch := runtime.GOARCH

	switch arch {
	case "amd64", "arm64":
		return arch, nil
	}
	return "", fmt.Errorf("your machine architecture %s is not currently supported", arch)
}
