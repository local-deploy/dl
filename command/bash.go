package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

var bashRoot bool

func bashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Login to PHP container",
		Long:  `Login to PHP container as www-data or root user and start bash shell.`,
		Run: func(cmd *cobra.Command, args []string) {
			runBash()
		},
		ValidArgs: []string{"--root"},
	}
	cmd.Flags().BoolVarP(&bashRoot, "root", "r", false, "Login as root")
	return cmd
}

func runBash() {
	project.LoadEnv()

	bash, lookErr := exec.LookPath("bash")
	docker, lookErr := exec.LookPath("docker")
	if lookErr != nil {
		fmt.Println(lookErr)
		return
	}

	site := project.Env.GetString("HOST_NAME")
	container := site + "_php"
	var root string

	if bashRoot {
		root = "--user root "
	}

	// TODO: rewrite to api
	cmdCompose := &exec.Cmd{
		Path:   bash,
		Args:   []string{bash, "-c", docker + " exec -it " + root + container + " /bin/bash"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	_ = cmdCompose.Run()
}
