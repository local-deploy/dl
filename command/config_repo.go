package command

import (
	"fmt"
	"log"

	"github.com/dixonwille/wmenu/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(configRepoCmd)
}

var configRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository source configuration",
	Long:  `Menu for setting up the images source repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		setRepo()
	},
	Hidden: false,
}

func setRepo() {
	menu := wmenu.NewMenu("Select application repository source:")
	menu.LoopOnInvalid()

	menu.Action(func(opts []wmenu.Opt) error { fmt.Println(opts[0].Value); return nil })

	locale := viper.GetString("repo")

	menu.Option("ghcr.io", "ghcr.io", locale == "ghcr.io", func(opt wmenu.Opt) error {
		saveRepoConfig(opt.Value)
		return nil
	})

	menu.Option("quay.io", "quay.io", locale == "quay.io", func(opt wmenu.Opt) error {
		saveRepoConfig(opt.Value)
		return nil
	})

	err := menu.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func saveRepoConfig(lang interface{}) {
	viper.Set("repo", lang)
	err := viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}

	switch lang {
	case "ghcr.io":
		fmt.Println("Selected ghcr.io")
	case "quay.io":
		fmt.Println("Selected quay.io")
	}
}
