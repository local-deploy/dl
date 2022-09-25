package command

import (
	"log"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configLangCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lang",
		Short: "Language configuration",
		Long:  `Menu for setting up the language.`,
		Run: func(cmd *cobra.Command, args []string) {
			configLangRun()
		},
		Hidden: true,
	}
	return cmd
}

func configLangRun() {
	currentRepo := viper.GetString("locale")
	options := []string{"en", "ru"}
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultOption(currentRepo).
		Show("Select application language")
	pterm.Printfln("Selected lang: %s", pterm.Green(selectedOption))

	saveLangConfig(selectedOption)
}

func saveLangConfig(lang interface{}) {
	viper.Set("locale", lang)
	err := viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
}
