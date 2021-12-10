package command

import (
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

func init() {
	rootCmd.AddCommand(bashCmd)
	bashCmd.Flags().BoolVarP(&bashRoot, "root", "r", false, "Login as root")
}

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Login to PHP container",
	Long:  `Login to PHP container as www-data or root user and start bash shell.`,
	Run: func(cmd *cobra.Command, args []string) {
		bash()
	},
}

var bashRoot bool

func bash() {
	project.LoadEnv()

	bash, lookErr := exec.LookPath("bash")
	docker, lookErr := exec.LookPath("docker")
	if lookErr != nil {
		pterm.FgRed.Println(lookErr)
		return
	}

	site := project.Env.GetString("HOST_NAME")
	container := site + "_php"
	var root string

	if bashRoot {
		root = "--user root "
	}

	cmdCompose := &exec.Cmd{
		Path:   bash,
		Args:   []string{bash, "-c", docker + " exec -it " + root + container + " /bin/bash"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	_ = cmdCompose.Run()
}
