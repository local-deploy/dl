package teleport

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/local-deploy/dl/utils/docker"
)

var pullWaitGroup sync.WaitGroup
var tsh string

// DeployTeleport Deploy using teleport
func DeployTeleport(ctx context.Context, database bool, files bool, override []string, tables []string) error {
	var err error
	w := progress.ContextWriter(ctx)

	tsh, err = teleportBin()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed deploy", "Teleport not installed"))
		return err
	}

	client, err := getClient()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed deploy", fmt.Sprint(err)))
		return err
	}

	if !database && !files {
		database = true
		files = true
	}

	if files {
		pullWaitGroup.Add(1)
		go startFiles(ctx, client, override)
	}

	if database {
		err = docker.UpDbContainer()
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Import failed", fmt.Sprint(err)))
			return err
		}
		pullWaitGroup.Add(1)
		go startDump(ctx, client, tables)
	}

	pullWaitGroup.Wait()

	return err
}

func startFiles(ctx context.Context, t *teleport, override []string) {
	defer pullWaitGroup.Done()
	copyFiles(ctx, t, override)
}

func startDump(ctx context.Context, t *teleport, tables []string) {
	defer pullWaitGroup.Done()
	dumpDB(ctx, t, tables)
}
