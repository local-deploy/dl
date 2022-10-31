package command

import (
	"os"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/utils"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dl",
		Short: "Deploy Local",
		Long: `Deploy Local — site deployment assistant locally.
Complete documentation is available at https://local-deploy.github.io/`,
	}
	debug bool
)

// Execute root command
func Execute() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableSorting:         true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	})
	logrus.SetLevel(logrus.FatalLevel)

	usageTemplate := usageTemplate()
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.DisableAutoGenTag = true
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show more output")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
			progress.Mode = "plain"
		}
	}

	rootCmd.AddCommand(
		envCommand(),
		psCommand(),
		bashCommand(),
		execCommand(),
		completionCommand(),
		configCommand(),
		deployCommand(),
		docsCommand(),
		upCommand(),
		downCommand(),
		recreateCommand(),
		serviceCommand(),
		selfUpdateCommand(),
		versionCommand(),
	)

	cobra.CheckErr(rootCmd.Execute())

	// check for new version
	utils.CheckUpdates()
}

// usageTemplate returns usage template for the command.
func usageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
