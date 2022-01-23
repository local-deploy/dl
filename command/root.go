package command

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dl",
	Short: "Deploy Local",
	Long: `Deploy Local — site deployment assistant locally.
Complete documentation is available at https://local-deploy.github.io/`,
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Execute root command
func Execute() {
	// pterm.Info.Prefix = pterm.Prefix{
	//	Text: "",
	// }
	// pterm.Success.Prefix = pterm.Prefix{
	//	Text: "",
	// }
	// pterm.Error.Prefix = pterm.Prefix{
	//	Text: "",
	// }
	// pterm.Warning.Prefix = pterm.Prefix{
	//	Text: "",
	// }

	usageTemplate := usageTemplate()

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.DisableAutoGenTag = true
	// rootCmd.CompletionOptions.DisableDefaultCmd = true

	cobra.CheckErr(rootCmd.Execute())
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
