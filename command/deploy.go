package command

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVarP(&database, "database", "d", false, "Dump only database from server")
	pullCmd.Flags().BoolVarP(&files, "files", "f", false, "Download only files from server")
	pullCmd.Flags().StringSliceVarP(&override, "override", "o", nil, "Override downloaded files (comma separated values)")
}

var pullCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Downloading db and files from the production server",
	Long: `Downloading database and kernel files from the production server.  
Without specifying the flag, files and the database are downloaded by default.  
If you specify a flag, for example -d, only the database will be downloaded.

Directories that are downloaded by default  
Bitrix CMS: "bitrix"  
WordPress: "wp-admin" and "wp-includes"  
Laravel: only the database is downloaded`,
	Run: func(cmd *cobra.Command, args []string) {
		deploy()
	},
	Example:   "dl deploy\ndl deploy -d\ndl deploy -f -o bitrix,upload",
	ValidArgs: []string{"--database", "--files", "--override"},
}

var (
	database      bool
	files         bool
	override      []string
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

	if !database && !files {
		database = true
		files = true
	}

	if files == true {
		pullWaitGroup.Add(1)
		go startFiles()
	}

	if database == true {
		err = upDbContainer()
		if err != nil {
			pterm.FgRed.Println("Import failed: ", err)
			os.Exit(1)
		}
		pullWaitGroup.Add(1)
		go startDump()
	}

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
		pterm.FgDefault.Println("Bitrix CMS detected")
		return "bitrix", nil
	}

	if strings.Contains(string(out), "wp-config.php") {
		pterm.FgDefault.Println("WordPress CMS detected")
		return "wordpress", nil
	}

	if strings.Contains(string(out), "artisan") {
		pterm.FgDefault.Println("Laravel FW detected")
		return "laravel", nil
	}

	return "", err
}

// upDbContainer Run db container before dump
func upDbContainer() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		pterm.Fatal.Printfln("Failed to connect to socket")
		return nil
	}

	site := project.Env.GetString("HOST_NAME")
	var siteDb = site + "_db"
	containerFilter := filters.NewArgs(filters.Arg("name", siteDb))
	containerExists, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: containerFilter})

	if len(containerExists) == 0 {
		compose, lookErr := exec.LookPath("docker-compose")
		if lookErr != nil {
			return lookErr
		}

		pterm.FgGreen.Printfln("Starting db container")

		cmdCompose := &exec.Cmd{
			Path: compose,
			Dir:  project.Env.GetString("PWD"),
			Args: []string{compose, "-p", project.Env.GetString("NETWORK_NAME"), "up", "-d", "db"},
			Env:  project.CmdEnv(),
		}

		err = cmdCompose.Run()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
