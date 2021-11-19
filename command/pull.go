package command

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
	"sync"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Downloading db and files from the production server",
	Long:  `Downloading database and kernel files from the production server.`,
	Run: func(cmd *cobra.Command, args []string) {
		pull()
	},
}

var pullWaitGroup sync.WaitGroup

func pull() {
	project.LoadEnv()

	pterm.FgGreen.Println("Create and download database dump")

	go startDump()

	pullWaitGroup.Add(1)
	pullWaitGroup.Wait()

	pterm.FgGreen.Println("All done")
}

func startDump() {
	defer pullWaitGroup.Done()
	project.DumpDb()
}
