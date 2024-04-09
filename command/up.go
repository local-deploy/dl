package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/local-deploy/dl/project"
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func upCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Up project",
		Long: `Start project containers. On completion, displays the local links to the project.
Analogue of the "docker-compose up -d" command.`,
		Run: func(cmd *cobra.Command, args []string) {
			upRun()
			// check for new version
			utils.CheckUpdates()
		},
	}
	return cmd
}

func upRun() {
	project.LoadEnv()

	if !utils.WpdeployCheck() {
		return
	}

	ctx := context.Background()
	cli, err := docker.NewClient()
	if err != nil {
		pterm.FgRed.Printfln("Failed to connect to socket")
		return
	}

	if !cli.IsServiceRunning(ctx) {
		err := startLocalServices()
		if err != nil {
			pterm.FgRed.Println(err)
			return
		}
	}

	pterm.FgGreen.Printfln("Starting project...")

	if viper.GetBool("ca") {
		pterm.FgGreen.Printfln("SSL certificate enabled")
		project.CreateCert()
	}

	bin, option := utils.GetCompose()
	Args := []string{bin}
	preArgs := []string{"-p", project.Env.GetString("NETWORK_NAME"), "--project-directory", project.Env.GetString("PWD"), "up", "-d"}

	if len(option) > 0 {
		Args = append(Args, option)
	}

	Args = append(Args, preArgs...)

	logrus.Infof("Run command: %s, args: %s", bin, Args)
	cmdCompose := &exec.Cmd{
		Path:   bin,
		Dir:    project.Env.GetString("PWD"),
		Args:   Args,
		Env:    project.CmdEnv(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err = cmdCompose.Run()
	if err != nil {
		pterm.FgRed.Printfln(fmt.Sprint(err))
		return
	}
	pterm.FgGreen.Printfln("Project has been successfully started")

	showProjectInfo()
}

func startLocalServices() error {
	reader := bufio.NewReader(os.Stdin)

	pterm.FgYellow.Print("Local services are not running. Would you like to launch (Y/n)? ")

	a, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	a = strings.TrimSpace(a)
	if strings.ToLower(a) == "y" || a == "" {
		ctx := context.Background()
		err := upServiceRun(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	//goland:noinspection GoErrorStringFormat
	return errors.New("start local services first: dl service up")
}

// showProjectInfo Display project links
func showProjectInfo() {
	l := project.Env.GetString("LOCAL_DOMAIN")
	n := project.Env.GetString("NIP_DOMAIN")

	schema := "http"

	if viper.GetBool("ca") {
		schema = "https"
	}

	pterm.FgCyan.Println()
	panels := pterm.Panels{
		{{Data: pterm.FgYellow.Sprintf("nip.io\nlocal")},
			{Data: pterm.FgYellow.Sprintf(schema+"://%s/\n"+schema+"://%s/", n, l)}},
	}

	_ = pterm.DefaultPanel.WithPanels(panels).WithPadding(5).Render()
}
