package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/lang"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "dl",
	Short: lang.Text("shortDl"),
	Long:  lang.Text("longDl"),
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	initConfig()

	//lang.SetLang(viper.GetString("lang"))
	//lang.SetLang("en")

	usageTemplate := UsageTemplate()
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := os.UserHomeDir()
	configDir := filepath.Join(home, ".dl")
	cobra.CheckErr(err)

	//viper.SetDefault("lang", "en")

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.Mkdir(configDir, 0755)
			viper.Set("version", version)

			viper.SafeWriteConfig()
		} else {
			panic(fmt.Errorf("Fatal error config file: %w \n", err))
		}
	}

	//env := viper.New()
	//
	//env.AddConfigPath("./")
	//env.SetConfigFile(".env")
	//env.SetConfigType("env")
	//env.ReadInConfig()
	//env.Debug()

	viper.AutomaticEnv()
}

// UsageTemplate returns usage template for the command.
func UsageTemplate() string {
	return lang.Text("rootUsage") + `:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + lang.Text("rootAliases") + `:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + lang.Text("rootExample") + `:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

` + lang.Text("rootAvailable") + `:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + lang.Text("rootFlags") + `:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + lang.Text("rootGlobalFlags") + `:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + lang.Text("rootAdditionalHelp") + `:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

` + lang.Text("rootUseHelp") + ` "{{.CommandPath}} [command] --help" ` + lang.Text("rootUseHelpMore") + `{{end}}
`
}
