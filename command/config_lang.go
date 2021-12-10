package command

import (
	"fmt"
	"log"

	"github.com/dixonwille/wmenu/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(configLangCmd)
}

var configLangCmd = &cobra.Command{
	Use:   "lang",
	Short: "Language configuration",
	Long:  `Menu for setting up the language.`,
	Run: func(cmd *cobra.Command, args []string) {
		setLang()
	},
	Hidden: true,
}

func setLang() {
	menu := wmenu.NewMenu("Select application language:")
	menu.LoopOnInvalid()

	menu.Action(func(opts []wmenu.Opt) error { fmt.Println(opts[0].Value); return nil })

	locale := viper.GetString("locale")

	menu.Option("English", "en", locale == "en", func(opt wmenu.Opt) error {
		saveLangConfig(opt.Value)
		return nil
	})
	menu.Option("Russian", "ru", locale == "ru", func(opt wmenu.Opt) error {
		saveLangConfig(opt.Value)
		return nil
	})

	err := menu.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func saveLangConfig(lang interface{}) {
	viper.Set("locale", lang)
	err := viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}

	switch lang {
	case "ru":
		fmt.Println("Выбран русский язык")
	case "en":
		fmt.Println("English language selected")
	}
}
