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
	"time"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/github"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var noConfig bool
var tag string

const maxBinSize = 104857600 // 100 MB

func selfUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"upgrade"},
		Short:   "Update dl",
		Long:    `Downloading the latest version of the app (if installed via bash script).`,
		Example: "dl self-update\ndl self-update -n\ndl self-update --tag 0.5.2",
		RunE: func(_ *cobra.Command, _ []string) error {
			return selfUpdateRun()
		},
	}
	cmd.Flags().BoolVarP(&noConfig, "no-overwrite", "n", false, "Do not overwrite configuration files")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Download the specified version")
	return cmd
}

func selfUpdateRun() error {
	ctx := context.Background()
	err := progress.RunWithTitle(ctx, selfUpdateService, os.Stdout, "Update")
	if err != nil {
		fmt.Println("Something went wrong...")
		return nil
	}

	if len(viper.GetString("version")) > 0 {
		printVersion(viper.GetString("version"))
	}

	return nil
}

func selfUpdateService(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	if utils.IsAptInstall() {
		pterm.FgYellow.Println("Please use command:")
		pterm.Println()
		pterm.FgGreen.Println("sudo apt update\nsudo apt install dl")
		os.Exit(0)
	}

	w.Event(progress.Event{ID: "Update", Status: progress.Working})
	w.Event(progress.Event{ID: "Getting the release", ParentID: "Update", Status: progress.Working})

	var release *github.Release
	var err error
	if len(tag) > 0 {
		var rxTag, _ = regexp.MatchString("^\\d.\\d.\\d+$", tag)
		if !rxTag {
			w.Event(progress.ErrorMessageEvent("Getting the release", fmt.Sprintf("Incorrect release format: %s", tag)))
			return err
		}
	}

	release, err = github.GetRelease("local-deploy", "dl", tag)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Getting the release", fmt.Sprintf("Failed: %s", err)))
		return err
	}
	w.Event(progress.Event{ID: "Getting the release", ParentID: "Update", Status: progress.Done})
	w.Event(progress.Event{ID: "Downloading", ParentID: "Update", Status: progress.Working})
	tmpPath := filepath.Join(os.TempDir(), release.AssetsName)
	err = downloadRelease(tmpPath, release.AssetsURL)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Downloading", fmt.Sprintf("Failed to download release: %s", err)))
		return err
	}
	w.Event(progress.Event{ID: "Downloading", ParentID: "Update", Status: progress.Done})
	w.Event(progress.Event{ID: "Unpacking", ParentID: "Update", Status: progress.Working})

	err = extractArchive(tmpPath)
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Unpacking", fmt.Sprintf("Extract archive failed: %s", err)))
		return err
	}

	err = copyBin()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Unpacking", fmt.Sprintf("Failed: %s", err)))
		return err
	}

	if !noConfig {
		err = utils.CreateTemplates(true)
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Unpacking", fmt.Sprint(err)))
			return err
		}
	}

	err = os.RemoveAll(filepath.Join(os.TempDir(), "dl"))
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Unpacking", fmt.Sprint(err)))
		return err
	}

	postUpdate(release)

	w.Event(progress.Event{ID: "Unpacking", ParentID: "Update", Status: progress.Done})
	w.Event(progress.Event{ID: "Update", Status: progress.Done})

	return nil
}

func postUpdate(release *github.Release) {
	viper.Set("version", release.Version)

	repo := viper.GetString("repo")
	if len(repo) == 0 {
		viper.Set("repo", "ghcr.io")
	}
	viper.Set("check-updates", time.Now())

	err := viper.WriteConfig()
	if err != nil {
		pterm.FgRed.Println(err)
	}
}

func downloadRelease(filepath string, url string) error {
	resp, err := http.Get(url) //nolint:bodyclose,gosec
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

	uncompressedStream, err := gzip.NewReader(io.LimitReader(reader, maxBinSize))
	if err != nil {
		return err
	}

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

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		tmpFiles := filepath.Join(tmpPath, filepath.Clean(header.Name))

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
			for {
				_, err := io.CopyN(outFile, tarReader, maxBinSize)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
			}

			err = outFile.Close()
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown type %x in %s", header.Typeflag, header.Name)
		}
	}

	return os.Remove(archivePath)
}

func copyBin() error {
	binPath := utils.BinPath()

	tmpBinPath := filepath.Join(os.TempDir(), "dl", "dl")

	if utils.IsBinFileExists() {
		err := os.Remove(binPath)
		if err != nil {
			return err
		}
	}

	bytesRead, err := os.ReadFile(tmpBinPath)
	if err != nil {
		return err
	}

	return os.WriteFile(binPath, bytesRead, 0600)
}

func printVersion(tag string) {
	pterm.DefaultSection.Printfln("DL has been successfully updated to version %s", tag)
}
