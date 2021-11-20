package command

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
	"os"
	"strings"
	"sync"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Downloading db and files from the production server",
	Long:  `Downloading database and kernel files from the production server.`,
	Run: func(cmd *cobra.Command, args []string) {
		deploy()
	},
}

var (
	pullWaitGroup sync.WaitGroup
	sshClient     *project.SshClient
)

func deploy() {
	project.LoadEnv()

	var err error
	sshClient, err = project.NewClient(&project.Server{
		Server:  project.Env.GetString("SERVER"),
		Key:     project.Env.GetString("SSH_KEY"),
		User:    project.Env.GetString("USER_SRV"),
		Port:    project.Env.GetUint("PORT_SRV"),
		Catalog: project.Env.GetString("CATALOG_SRV"),
	})

	// Defer closing the network connection.
	defer func(client *project.SshClient) {
		err = client.Close()
		if err != nil {
			pterm.FgRed.Println(err)
			return
		}
	}(sshClient)

	sshClient.Server.FwType, err = detectFw()
	if err != nil {
		pterm.FgRed.Printfln("Failed to determine the FW. Please specify accesses manually.")
		os.Exit(1)
	}

	pullWaitGroup.Add(2)

	go startFiles()
	go startDump()

	pullWaitGroup.Wait()

	pterm.FgGreen.Println("All done")
}

func startFiles() {
	defer pullWaitGroup.Done()
	sshClient.CopyFiles()
}

func startDump() {
	defer pullWaitGroup.Done()
	sshClient.DumpDb()
}

func detectFw() (string, error) {
	ls := strings.Join([]string{"cd", sshClient.Server.Catalog, "&&", "ls"}, " ")
	out, err := sshClient.Run(ls)
	if err != nil {
		pterm.FgRed.Println(err)
	}

	if strings.Contains(string(out), "bitrix") {
		pterm.FgGreen.Println("Bitrix CMS detected")
		return "bitrix", nil
	}

	if strings.Contains(string(out), "wp-config.php") {
		pterm.FgGreen.Println("WordPress CMS detected")
		return "wordpress", nil
	}

	if strings.Contains(string(out), "artisan") {
		pterm.FgGreen.Println("Laravel FW detected")
		return "laravel", nil
	}

	return "", err
}
