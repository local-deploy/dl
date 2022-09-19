package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func completionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "completion [bash|zsh]",
		Short:                 "Generate completion script",
		Long:                  fmt.Sprintf(completionDesc, rootCmd.Root().Name()),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh"},
		Run: func(cmd *cobra.Command, args []string) {
			runCompletion(cmd, args)
		},
	}
	return cmd
}

func runCompletion(cmd *cobra.Command, shell []string) {
	if len(shell) > 0 {
		switch shell[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		}
	} else {
		fmt.Printf(completionDesc, rootCmd.Root().Name())
	}
}

var completionDesc = `To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  $ echo "\nsource <(%[1]s completion bash)" >> ~/.bashrc

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

You will need to start a new shell for this setup to take effect.
`
