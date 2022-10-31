package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

func execCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [command]",
		Short:   "Executing a command in a PHP container",
		Long:    `Running bash command in PHP container as user www-data`,
		Example: "dl exec composer install\ndl exec \"ls -la\"",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runExec(args)
		},
	}
	return cmd
}

func runExec(args []string) {
	project.LoadEnv()

	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println(err)
		return
	}

	command := []string{docker, "exec", "-t", "-w", "/var/www/html"}

	site := project.Env.GetString("HOST_NAME")
	command = append(command, site+"_php")
	command = append(command, args...)

	logrus.Infof("Use container name %s", site+"_php")
	logrus.Infof("Command execution %s", strings.Join(args, " "))

	out, err := exec.Command("bash", "-c", strings.Join(command, " ")).CombinedOutput()
	if err != nil {
		logrus.Error(err)
	}
	fmt.Println(string(out))
}
