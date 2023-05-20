package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/local-deploy/dl/utils"
	"github.com/pkg/sftp"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

// NewSftp returns new sftp client and error if any.
func (c Client) NewSftp(opts ...sftp.ClientOption) (*sftp.Client, error) {
	return sftp.NewClient(c.Client, opts...)
}

// CleanRemote Deleting file on the server
func (c Client) CleanRemote(remotePath string) (err error) {
	ftp, err := c.NewSftp()
	if err != nil {
		return err
	}

	defer func(ftp *sftp.Client) {
		err := ftp.Close()
		if err != nil {
			pterm.FgRed.Println(err)
		}
	}(ftp)

	logrus.Infof("Delete file: %s", remotePath)
	err = ftp.Remove(remotePath)

	return err
}

// Download file from remote server
//
//goland:noinspection GoUnhandledErrorResult
func (c Client) Download(ctx context.Context, remotePath, localPath string) (err error) {
	// w := progress.ContextWriter(ctx)
	local, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer local.Close()

	ftp, err := c.NewSftp()
	if err != nil {
		return
	}
	defer ftp.Close()

	fileInfo, err := ftp.Stat(remotePath)
	if err != nil {
		return
	}
	defer ftp.Close()

	localDisk := utils.FreeSpaceHome()
	if fileInfo.Size() > int64(localDisk.Free) {
		remoteSize := utils.HumanSize(float64(fileInfo.Size()))
		localSize := utils.HumanSize(float64(localDisk.Free))
		return errors.New(fmt.Sprintf("No disk space. Filesize %s, free space %s", remoteSize, localSize))
	}

	remote, err := ftp.Open(remotePath)
	if err != nil {
		return
	}
	defer remote.Close()

	if _, err = io.Copy(local, remote); err != nil {
		return
	}

	return local.Sync()
}

// Upload a local file to remote server!
func (c Client) Upload(localPath string, remotePath string) (err error) {
	local, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer local.Close()

	ftp, err := c.NewSftp()
	if err != nil {
		return
	}
	defer ftp.Close()

	remote, err := ftp.Create(remotePath)
	if err != nil {
		return
	}
	defer remote.Close()

	_, err = io.Copy(remote, local)
	return
}

// Close client net connection.
func (c Client) Close() error {
	return c.Client.Close()
}
