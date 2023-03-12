package command

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/utils/github"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var noConfig bool
var tag string

func selfUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"upgrade"},
		Short:   "Update dl",
		Long:    `Downloading the latest version of the app.`,
		Example: "dl self-update\ndl self-update -n\ndl self-update --tag 0.5.2",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			tag, err := progress.RunWithStatus(ctx, selfUpdateRun)

			if err != nil {
				return err
			}
			if len(tag) > 0 {
				printVersion(tag)
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&noConfig, "no-overwrite", "n", false, "Do not overwrite configuration files")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Download the specified version")
	return cmd
}

func selfUpdateRun(ctx context.Context) (string, error) {
	w := progress.ContextWriter(ctx)

	w.Event(progress.Event{ID: "Update", Status: progress.Working})
	w.Event(progress.Event{ID: "Getting the release", ParentID: "Update", Status: progress.Working})

	var release *github.Release
	var err error
	if len(tag) > 0 {
		var rxTag, _ = regexp.MatchString("^\\d.\\d.\\d+$", tag)
		if !rxTag {
			w.Event(progress.ErrorMessageEvent("Getting the release", fmt.Sprint("Incorrect release format")))
			return "", nil
		}

		release, err = github.GetRelease("local-deploy", "dl", tag)
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Getting the release", fmt.Sprintf("Failed to get release: %s", err)))
			return "", nil
		}
	} else {
		release, err = github.GetLatestRelease("local-deploy", "dl")
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Getting the release", fmt.Sprintf("Failed to get release: %s", err)))
			return "", nil
		}
	}

	w.Event(progress.Event{ID: "Getting the release", ParentID: "Update", Status: progress.Done})

	time.Sleep(time.Second)
	tmpPath := filepath.Join(os.TempDir(), release.AssetsName)

	w.Event(progress.Event{ID: "Downloading release", ParentID: "Update", Status: progress.Working})
	err = downloadRelease(tmpPath, release.AssetsURL)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Downloading release", fmt.Sprintf("Failed to download release: %s", err)))
		return "", nil
	}
	w.Event(progress.Event{ID: "Downloading release", ParentID: "Update", Status: progress.Done})

	w.Event(progress.Event{ID: "Unpacking archive", ParentID: "Update", Status: progress.Working})
	err = extractArchive(tmpPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Unpacking archive", fmt.Sprintf("Extract archive failed: %s", err)))
		return "", nil
	}
	w.Event(progress.Event{ID: "Unpacking archive", ParentID: "Update", Status: progress.Done})
	w.Event(progress.Event{ID: "Copying files", ParentID: "Update", Status: progress.Working})

	err = copyBin()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Copying files", fmt.Sprintf("Failed: %s", err)))
		return "", nil
	}

	if !noConfig {
		err = copyConfigFiles()
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Copying files", fmt.Sprint(err)))
			return "", nil
		}
	}
	w.Event(progress.Event{ID: "Copying files", ParentID: "Update", Status: progress.Done})
	w.Event(progress.Event{ID: "Cleaning up temporary directory", ParentID: "Update", Status: progress.Working})

	err = os.RemoveAll(filepath.Join(os.TempDir(), "dl"))
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Cleaning up temporary directory", fmt.Sprint(err)))
		return "", nil
	}
	w.Event(progress.Event{ID: "Cleaning up temporary directory", ParentID: "Update", Status: progress.Done})

	viper.Set("version", release.Version)

	repo := viper.GetString("repo")
	if len(repo) == 0 {
		viper.Set("repo", "ghcr.io")
	}
	viper.Set("check-updates", time.Now())

	err = viper.WriteConfig()
	if err != nil {
		pterm.FgRed.Println(err)
	}

	w.Event(progress.Event{ID: "Update", Status: progress.Done})

	return release.Version, nil
}

func downloadRelease(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			pterm.FgRed.Printfln("Request failed: %v", err)
			os.Exit(1)
		}
	}(resp.Body)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			pterm.FgRed.Printfln("File creation error: %v", err)
			os.Exit(1)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractArchive(archivePath string) error {
	reader, err := os.Open(archivePath)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	tmpPath := filepath.Join(os.TempDir(), "dl")

	_, err = os.Stat(tmpPath)
	if err == nil {
		err = os.RemoveAll(tmpPath)
		if err != nil {
			return err
		}
	}

	if err := os.Mkdir(tmpPath, 0755); err != nil {
		return err
	}

	for {
		header, err := tarReader.Next()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		tmpFiles := filepath.Join(tmpPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(tmpFiles, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(tmpFiles)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			err = outFile.Close()
			if err != nil {
				pterm.FgRed.Println(err)
				return err
			}

		default:
			return fmt.Errorf("extract archive failed. Unknown type %x in %s", header.Typeflag, header.Name)
		}
	}

	err = os.Remove(archivePath)
	if err != nil {
		return err
	}

	return nil
}

func copyBin() error {
	var (
		err    error
		system string
		arch   string
	)

	binPath, _ := helper.BinPath()

	system, err = getSystem()
	if err != nil {
		return err
	}

	arch, err = getArch()
	if err != nil {
		return err
	}

	tmpLinuxBin := strings.Join([]string{"dl", system, arch}, "_")
	tmpBinPath := filepath.Join(os.TempDir(), "dl", "bin", tmpLinuxBin)

	if helper.IsBinFileExists() {
		err = os.Remove(binPath)
		if err != nil {
			return err
		}
	}

	bytesRead, err := os.ReadFile(tmpBinPath)
	err = os.WriteFile(binPath, bytesRead, 0775) //nolint:gosec
	if err != nil {
		return err
	}
	return nil
}

func getSystem() (string, error) {
	system := runtime.GOOS

	switch system {
	case "linux", "darwin":
		return system, nil
	}
	return "", fmt.Errorf("This installer does not support %s platform at this time", system)
}

func getArch() (string, error) {
	arch := runtime.GOARCH

	switch arch {
	case "amd64", "arm64":
		return arch, nil
	}
	return "", fmt.Errorf("Your machine architecture %s is not currently supported", arch)
}

func copyConfigFiles() error {
	confDir, _ := helper.ConfigDir()

	tmpConfigFiles := filepath.Join(os.TempDir(), "dl", "config-files")
	configFilesDir := filepath.Join(confDir, "config-files")

	rm := os.RemoveAll(configFilesDir)
	if rm != nil {
		return rm
	}

	mkdir := os.Mkdir(configFilesDir, 0775)
	if mkdir != nil {
		return mkdir
	}

	var err = filepath.Walk(tmpConfigFiles, func(path string, info os.FileInfo, err error) error {
		var relPath = strings.Replace(path, tmpConfigFiles, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(configFilesDir, relPath), 0755)
		} else {
			var data, err = os.ReadFile(filepath.Join(tmpConfigFiles, relPath))
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(configFilesDir, relPath), data, 0644) //nolint:gosec
		}
	})

	return err
}

func printVersion(tag string) {
	pterm.DefaultSection.Printfln("DL has been successfully updated to version %s", tag)
}
