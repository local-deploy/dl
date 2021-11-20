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

var (
	pullWaitGroup sync.WaitGroup
	sshClient     *project.SshClient
)

func pull() {
	project.LoadEnv()

	var err error
	sshClient, err = project.NewClient(&project.Server{
		Server:  project.Env.GetString("SERVER"),
		Key:     project.Env.GetString("SSH_KEY"),
		User:    project.Env.GetString("USER_SRV"),
		Catalog: project.Env.GetString("CATALOG_SRV"),
		Port:    project.Env.GetUint("PORT_SRV"),
	})

	// Defer closing the network connection.
	defer func(client *project.SshClient) {
		err = client.Close()
		if err != nil {
			pterm.FgRed.Println(err)
			return
		}
	}(sshClient)

	pterm.FgGreen.Println("Create and download database dump")
	go startDump()

	pullWaitGroup.Add(1)
	pullWaitGroup.Wait()

	pterm.FgGreen.Println("All done")
}

func startDump() {
	defer pullWaitGroup.Done()
	sshClient.DumpDb()
}
