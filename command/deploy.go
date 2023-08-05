package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/local-deploy/dl/project"
	"github.com/local-deploy/dl/utils/client"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/local-deploy/dl/utils/teleport"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	database      bool
	files         bool
	override      []string
	tables        []string
	pullWaitGroup sync.WaitGroup
	sshClient     *client.Client
)

func deployCommand() *cobra.Command {
	cmd := &cobra.Command{
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
			return deployRun()
		},
		Example:   "dl deploy\ndl deploy -d\ndl deploy -d -t b_user,b_file\ndl deploy -f\ndl deploy -f -o bitrix,upload",
		ValidArgs: []string{"--database", "--files", "--override"},
	}
	cmd.Flags().BoolVarP(&database, "database", "d", false, "Dump only database from server")
	cmd.Flags().BoolVarP(&files, "files", "f", false, "Download only files from server")
	cmd.Flags().StringSliceVarP(&override, "override", "o", nil, "Override downloaded files (comma separated values)")
	cmd.Flags().StringSliceVarP(&tables, "tables", "t", nil, "Dump only specified tables (comma separated values)")
	return cmd
}

func deployRun() error {
	ctx := context.Background()
	err := progress.Run(ctx, deployService)
	if err != nil {
		fmt.Println("Something went wrong...")
		return nil
	}

	fmt.Println("All done")

	showSpecificInfo()

	return nil
}

// showProjectInfo Display specific FW info
func showSpecificInfo() {
	if sshClient.Config.FwType == "wordpress" {
		n := project.Env.GetString("NIP_DOMAIN")
		pterm.Println()
		pterm.FgYellow.Println("Please specify the domain in the wp-config.php file:")
		pterm.FgDefault.Printfln("define('WP_HOME', 'http://%s');\ndefine('WP_SITEURL', 'http://%s');", n, n)
	}
}

func deployService(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	project.LoadEnv()

	var err error

	if len(project.Env.GetString("TELEPORT")) > 0 {
		sshClient = &client.Client{Config: &client.Config{FwType: "bitrix"}}
		fmt.Println("Deploy using Teleport")
		return teleport.DeployTeleport(ctx, database, files, override, tables)
	}

	sshClient, err = getClient()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed to connect", fmt.Sprint(err)))
		return err
	}

	// Defer closing the network connection.
	defer func(client *client.Client) {
		err = client.Close()
		if err != nil {
			return
		}
	}(sshClient)

	sshClient.Config.FwType, err = detectFw()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Detect FW", fmt.Sprint(err)))
		return err
	}

	if !database && !files {
		database = true
		files = true
	}

	if files {
		pullWaitGroup.Add(1)
		go startFiles(ctx, sshClient)
	}

	if database {
		err = docker.UpDbContainer()
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Import failed", fmt.Sprint(err)))
			return err
		}
		pullWaitGroup.Add(1)
		go startDump(ctx, sshClient)
	}

	pullWaitGroup.Wait()

	return err
}

func getClient() (c *client.Client, err error) {
	server := &client.Config{
		Addr:             project.Env.GetString("SERVER"),
		Key:              project.Env.GetString("SSH_KEY"),
		UseKeyPassphrase: project.Env.GetBool("ASK_KEY_PASSPHRASE"),
		UsePassword:      project.Env.GetBool("USE_SSH_PASS"),
		User:             project.Env.GetString("USER_SRV"),
		Port:             project.Env.GetUint("PORT_SRV"),
		Catalog:          project.Env.GetString("CATALOG_SRV"),
	}
	logrus.Infof("SSH client connect %v", fmt.Sprint(server))
	c, err = client.NewClient(server)
	return
}

func startFiles(ctx context.Context, c *client.Client) {
	defer pullWaitGroup.Done()
	project.CopyFiles(ctx, c, override)
}

func startDump(ctx context.Context, c *client.Client) {
	defer pullWaitGroup.Done()
	project.DumpDb(ctx, c, tables)
}

func detectFw() (string, error) {
	ls := strings.Join([]string{"cd", sshClient.Config.Catalog, "&&", "ls"}, " ")
	out, err := sshClient.Run(ls)
	if err != nil {
		return "", err
	}

	logrus.Info("Detect Framework")
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

	return "", errors.New("failed determine the Framework, please specify accesses manually https://clck.ru/uAGwX")
}
