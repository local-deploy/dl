package helper

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
)

// WpdeployCheck Legacy app check
func WpdeployCheck() bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		pterm.FgRed.Println("Failed to connect to socket")
		return false
	}

	containerFilter := filters.NewArgs(filters.Arg("label", "com.docker.compose.project=local-services"))
	isExists, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: containerFilter})
	if err != nil {
		pterm.FgRed.Println(err)
		return false
	}
	if len(isExists) > 0 {
		err := wpdeployDown()
		if err != nil {
			pterm.FgRed.Println(err)
			return false
		}
		return false
	}
	return true
}

func wpdeployDown() error {
	dir, _ := os.Getwd()
	wpdeploy, _ := exec.LookPath("wpdeploy")
	reader := bufio.NewReader(os.Stdin)

	pterm.FgYellow.Print("An old version of wpdeploy is running. Do you want to stop it (Y/n)? ")

	a, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	//goland:noinspection GoErrorStringFormat
	errorMsg := errors.New("Stop wpdeploy local services first: wpdeploy local-services down")

	a = strings.TrimSpace(a)
	if strings.ToLower(a) == "y" || a == "" {
		cmdDown := &exec.Cmd{
			Path:   wpdeploy,
			Dir:    dir,
			Args:   []string{wpdeploy, "local-services", "down"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
		err = cmdDown.Run()
		if err != nil {
			return err
		}
		return nil
	}
	return errorMsg
}
