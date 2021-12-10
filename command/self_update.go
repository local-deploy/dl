package command

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"github.com/google/go-github/v41/github"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/helper"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
	selfUpdateCmd.Flags().BoolVarP(&noConfig, "no-overwrite", "n", false, "Do not overwrite configuration files")
}

var selfUpdateCmd = &cobra.Command{
	Use:     "self-update",
	Aliases: []string{"upgrade"},
	Short:   "Update dl",
	Long:    `Downloading the latest version of the app.`,
	Run: func(cmd *cobra.Command, args []string) {
		selfUpdate()
	},
}

var noConfig bool

func selfUpdate() {
	client := github.NewClient(nil)

	spinnerRelease, _ := pterm.DefaultSpinner.Start("Getting the latest release")
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "local-deploy", "dl")
	if err != nil {
		spinnerRelease.Fail("Failed to get release: %v", err)
		return
	}

	time.Sleep(time.Second)
	tmpPath := filepath.Join(os.TempDir(), *release.Assets[0].Name)

	spinnerRelease.UpdateText("Downloading release")
	err = downloadRelease(tmpPath, *release.Assets[0].BrowserDownloadURL)
	if err != nil {
		spinnerRelease.Fail("Failed to download release: %v", err)
		return
	}
	spinnerRelease.Success()

	spinnerExtract, _ := pterm.DefaultSpinner.Start("Unpacking archive")
	err = extractArchive(tmpPath)
	if err != nil {
		spinnerExtract.Fail("Extract archive failed: %s", err)
		os.Exit(1)
	}
	spinnerExtract.Success()

	spinnerFiles, _ := pterm.DefaultSpinner.Start("Copying files")
	err = copyBin()
	if err != nil {
		spinnerFiles.Fail("Failed: %v", err)
		return
	}

	if noConfig == false {
		err = copyConfigFiles()
		if err != nil {
			spinnerFiles.Fail("Failed: %v", err)
			return
		}
	}
	spinnerFiles.Success()

	spinnerTmp, _ := pterm.DefaultSpinner.Start("Cleaning up temporary directory")
	err = os.RemoveAll(filepath.Join(os.TempDir(), "dl"))
	if err != nil {
		spinnerTmp.Fail("Failed: %v", err)
		return
	}
	spinnerTmp.Success()

	viper.Set("version", *release.TagName)
	err = viper.WriteConfig()
	if err != nil {
		pterm.FgRed.Println(err)
	}

	pterm.DefaultSection.Printfln("DL has been successfully updated to version %s", *release.TagName)
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
	if err := os.Mkdir(tmpPath, 0755); err != nil {
		return err
	}

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
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
			return errors.New(pterm.Sprint("extract archive failed. Unknown type %s in %s", header.Typeflag, header.Name))
		}
	}

	err = os.Remove(archivePath)
	if err != nil {
		return err
	}

	return nil
}

func copyBin() error {
	var err error

	binPath, _ := helper.BinPath()
	//TODO: Darwin
	tmpLinuxBin := filepath.Join(os.TempDir(), "dl", "bin", "dl_linux_amd64")

	if helper.IsBinFileExists() {
		err = os.Remove(binPath)
		if err != nil {
			return err
		}
	}

	bytesRead, err := ioutil.ReadFile(tmpLinuxBin)
	err = ioutil.WriteFile(binPath, bytesRead, 0775)
	if err != nil {
		return err
	}
	return nil
}

func copyConfigFiles() error {
	confDir, _ := helper.ConfigDir()

	tmpConfigFiles := filepath.Join(os.TempDir(), "dl", "config-files")
	configFilesDir := filepath.Join(confDir, "config-files")

	rm := os.RemoveAll(configFilesDir)
	if rm != nil {
		return rm
	}

	mdir := os.Mkdir(configFilesDir, 0775)
	if mdir != nil {
		return mdir
	}

	var err = filepath.Walk(tmpConfigFiles, func(path string, info os.FileInfo, err error) error {
		var relPath = strings.Replace(path, tmpConfigFiles, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(configFilesDir, relPath), 0755)
		} else {
			var data, err = ioutil.ReadFile(filepath.Join(tmpConfigFiles, relPath))
			if err != nil {
				return err
			}
			return ioutil.WriteFile(filepath.Join(configFilesDir, relPath), data, 0644)
		}
	})

	return err
}
