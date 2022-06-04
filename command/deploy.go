package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy()
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

func deploy() error {
	ctx := context.Background()
	err := progress.Run(ctx, func(ctx context.Context) error {
		return deployService(ctx)
	})
	if err != nil {
		return err
	}

	fmt.Println("All done")

	return nil
}

func deployService(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	project.LoadEnv()

	var err error
	sshClient, err = project.NewClient(&project.Server{
		Server:  project.Env.GetString("SERVER"),
		Key:     project.Env.GetString("SSH_KEY"),
		User:    project.Env.GetString("USER_SRV"),
		Port:    project.Env.GetUint("PORT_SRV"),
		Catalog: project.Env.GetString("CATALOG_SRV"),
	})

	if err != nil {
		w.Event(progress.ErrorMessageEvent("Client", "Failed connect to ssh"))
		os.Exit(1)
	}

	// Defer closing the network connection.
	defer func(client *project.SshClient) {
		err = client.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(sshClient)

	sshClient.Server.FwType, err = detectFw()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Detect FW", "Failed determine the FW. Please specify accesses manually."))
		os.Exit(1)
	}

	if !database && !files {
		database = true
		files = true
	}

	if files == true {
		pullWaitGroup.Add(1)
		go startFiles(ctx)
	}

	if database == true {
		err = upDbContainer()
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Import failed", fmt.Sprint(err)))
			os.Exit(1)
		}
		pullWaitGroup.Add(1)
		go startDump(ctx)
	}

	pullWaitGroup.Wait()

	return err
}

func startFiles(ctx context.Context) {
	defer pullWaitGroup.Done()
	sshClient.CopyFiles(ctx, override)
}

func startDump(ctx context.Context) {
	defer pullWaitGroup.Done()
	sshClient.DumpDb(ctx)
}

func detectFw() (string, error) {
	ls := strings.Join([]string{"cd", sshClient.Server.Catalog, "&&", "ls"}, " ")
	out, err := sshClient.Run(ls)
	if err != nil {
		fmt.Println(err)
	}

	if strings.Contains(string(out), "bitrix") {
		fmt.Println("Bitrix CMS detected")
		return "bitrix", nil
	}

	if strings.Contains(string(out), "wp-config.php") {
		fmt.Println("WordPress CMS detected")
		return "wordpress", nil
	}

	if strings.Contains(string(out), "artisan") {
		fmt.Println("Laravel FW detected")
		return "laravel", nil
	}

	return "", err
}

// upDbContainer Run db container before dump
func upDbContainer() error {
	ctx := context.Background()
	w := progress.ContextWriter(ctx)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed to connect to socket", fmt.Sprint(err)))
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

		w.Event(progress.StartingEvent("Starting db container"))

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

		w.Event(progress.StartedEvent("Starting db container"))

		return nil
	}
	return nil
}
