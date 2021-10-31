package command

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"net"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVarP(&projectContainer, "container", "c", "", "Start single container")
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up project",
	Long:  `Up project.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
	Example: "dl up\ndl up -c db",
}

func up() {
	helper.LoadEnv()

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	var singleContainer = ""
	if len(projectContainer) > 0 {
		singleContainer = projectContainer
	}

	cmdCompose := &exec.Cmd{
		Path: compose,
		Dir:  helper.ProjectEnv.GetString("PWD"),
		Args: []string{compose, "-p", helper.ProjectEnv.GetString("NETWORK_NAME"), "up", "-d", singleContainer},
		Env:  helper.CmdEnv(),
	}

	startProject, _ := pterm.DefaultSpinner.Start("Starting project")
	if err := cmdCompose.Run(); err != nil {
		//TODO: add errors
		startProject.Fail(err)
		return
	}
	startProject.UpdateText("Project has been successfully started")
	startProject.Success()

	p := helper.ProjectEnv.GetString("APP_NAME")
	h := getLocalIp()
	pterm.FgCyan.Println()
	panels := pterm.Panels{
		{{Data: pterm.FgYellow.Sprintf("nip.io\nlocal")},
			{Data: pterm.FgYellow.Sprintf("http://%s.%s.nip.io/\nhttp://%s.localhost/", p, h, p)}},
	}

	_ = pterm.DefaultPanel.WithPanels(panels).WithPadding(5).Render()
}

func getLocalIp() string {
	//name, _ := os.Hostname()
	address, err := net.LookupHost("localhost")
	if err != nil {
		fmt.Printf("%v\n", err)
		return ""
	}

	return address[0]
}
