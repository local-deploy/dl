package command

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
	"net"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up project",
	Long:  `Up project.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func up() {
	project.LoadEnv()

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	cmdCompose := &exec.Cmd{
		Path: compose,
		Dir:  project.Env.GetString("PWD"),
		Args: []string{compose, "-p", project.Env.GetString("NETWORK_NAME"), "up", "-d"},
		Env:  project.CmdEnv(),
	}

	startProject, _ := pterm.DefaultSpinner.Start("Starting project")
	output, err := cmdCompose.CombinedOutput()
	if err != nil {
		startProject.Fail(fmt.Sprint(err) + ": " + string(output))
		return
	}
	startProject.UpdateText("Project has been successfully started")
	startProject.Success()

	showProjectInfo()
}

//showProjectInfo Display project links
func showProjectInfo() {
	p := project.Env.GetString("APP_NAME")
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
