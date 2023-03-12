package utils

import (
	"time"

	"github.com/local-deploy/dl/utils/github"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// CheckUpdates checking for updates every 24 hours
func CheckUpdates() {
	now := time.Now()
	lastCheck := viper.GetTime("check-updates")

	if lastCheck.Add(24 * time.Hour).After(now) {
		return
	}

	if isAvailableNewVersion() {
		printNotice()
	} else {
		viper.Set("check-updates", now)
		err := viper.WriteConfig()
		if err != nil {
			return
		}
	}
}

func isAvailableNewVersion() bool {
	currentVersion := viper.GetString("version")
	release, err := github.GetLatestRelease("local-deploy", "dl")
	if err != nil {
		// we don't want an error on a bad request
		return false
	}

	return currentVersion != release.Version
}

func printNotice() {
	pterm.FgGreen.Printfln("New version is available! Please update: dl self-update")
}
