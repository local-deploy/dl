package command

import (
	"log"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configRepoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository source configuration",
		Long:  `Menu for setting up the images source repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			configRepoRun()
		},
		Hidden: false,
	}
	return cmd
}

func configRepoRun() {
	currentRepo := viper.GetString("repo")
	options := []string{"ghcr.io", "quay.io"}
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultOption(currentRepo).
		Show("Select application repository source")
	pterm.Printfln("Selected repo: %s", pterm.Green(selectedOption))

	saveRepoConfig(selectedOption)
}

func saveRepoConfig(lang interface{}) {
	viper.Set("repo", lang)
	err := viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
}
