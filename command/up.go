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

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	pterm.FgGreen.Printfln("Starting project...")

	cmdCompose := &exec.Cmd{
		Path:   compose,
		Dir:    project.Env.GetString("PWD"),
		Args:   []string{compose, "-p", project.Env.GetString("NETWORK_NAME"), "up", "-d"},
		Env:    project.CmdEnv(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err = cmdCompose.Run()
	if err != nil {
		pterm.FgGreen.Printfln(fmt.Sprint(err))
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
