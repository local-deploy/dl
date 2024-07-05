package command

import (
	"log"

	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Additional service containers",
		Long:  `Menu for managing the launch of additional containers (portainer and mailhog).`,
		Run: func(_ *cobra.Command, _ []string) {
			configServiceRun()
		},
		Hidden: false,
	}
	return cmd
}

func configServiceRun() {
	hasKeys := viper.IsSet("services")
	currentServices := viper.GetStringSlice("services")
	if !hasKeys {
		currentServices = append(currentServices, "portainer", "mail")
	}
	options := []string{"portainer", "mail"}
	selectedOption, _ := pterm.DefaultInteractiveMultiselect.
		WithOptions(options).
		WithFilter(false).
		WithKeySelect(keys.Space).
		WithKeyConfirm(keys.Enter).
		WithDefaultOptions(currentServices).
		Show("Select the services that should be started with the 'dl service up' command")
	pterm.Printfln("Selected services: %s", pterm.Green(selectedOption))

	saveServiceConfig(selectedOption)
}

func saveServiceConfig(lang interface{}) {
	viper.Set("services", lang)
	err := viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
}
