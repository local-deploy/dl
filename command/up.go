package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"github.com/varrcan/dl/project"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up project",
	Long: `Start project containers. On completion, displays the local links to the project.
Analogue of the "docker-compose up -d" command.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func up() {
	project.LoadEnv()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		pterm.Fatal.Printfln("Failed to connect to socket")
		return
	}

	containerFilter := filters.NewArgs(filters.Arg("name", "traefik"))
	traefikExists, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: containerFilter})

	if len(traefikExists) == 0 {
		pterm.FgRed.Printfln("Start local services first: dl service up")
		return
	}

	pterm.FgGreen.Printfln("Starting project...")

	bin, option := helper.GetCompose()
	Args := []string{bin}
	preArgs := []string{"-p", project.Env.GetString("NETWORK_NAME"), "up", "-d"}

	if len(option) > 0 {
		Args = append(Args, option)
	}

	Args = append(Args, preArgs...)

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

// showProjectInfo Display project links
func showProjectInfo() {
	l := project.Env.GetString("LOCAL_DOMAIN")
	n := project.Env.GetString("NIP_DOMAIN")

	pterm.FgCyan.Println()
	panels := pterm.Panels{
		{{Data: pterm.FgYellow.Sprintf("nip.io\nlocal")},
			{Data: pterm.FgYellow.Sprintf("http://%s/\nhttp://%s/", n, l)}},
	}

	_ = pterm.DefaultPanel.WithPanels(panels).WithPadding(5).Render()
}
