package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

var bashRoot bool

func bashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Login to PHP container",
		Long: `Login to PHP container as www-data or root user and start bash shell.
As the second parameter, you can specify the name or ID of another docker container.
Default is always the PHP container.`,
		Example: "dl bash\ndl bash -r\ndl bash site.com_db\ndl bash fcb13f1a3ea7",
		Run: func(cmd *cobra.Command, args []string) {
			runBash(args)
		},
		ValidArgs: []string{"--root"},
	}
	cmd.Flags().BoolVarP(&bashRoot, "root", "r", false, "Login as root")
	return cmd
}

func runBash(args []string) {
	project.LoadEnv()

	bash, err := exec.LookPath("bash")
	if err != nil {
		fmt.Println(err)
		return
	}
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println(err)
		return
	}

	var command []string

	if bashRoot {
		logrus.Info("Login as root user")
		command = append(command, "--user root")
	}

	if len(args) > 0 {
		command = append(command, args[0])
		logrus.Infof("Use container name %s", args[0])
	} else {
		command = append(command, "-w /var/www/html")

		site := project.Env.GetString("HOST_NAME")
		command = append(command, site+"_php")
		logrus.Infof("Use container name %s", site+"_php")
	}

	cmdCompose := &exec.Cmd{
		Path:   bash,
		Args:   []string{bash, "-c", docker + " exec -it " + strings.Join(command, " ") + " /bin/bash"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	err = cmdCompose.Run()
	if err != nil {
		logrus.Error(err)
	}
}
