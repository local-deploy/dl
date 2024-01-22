package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"golang.org/x/sync/errgroup"
)

// PullRequiredImages check for the presence of the image in the system and download if necessary
// Deprecated
func (cli *Client) PullRequiredImages(ctx context.Context, containers Containers) error {

	return progress.Run(ctx, func(ctx context.Context) error {
		w := progress.ContextWriter(ctx)
		eg, ctx := errgroup.WithContext(ctx)

		for _, con := range containers {
			container := con
			imageFiler := filters.NewArgs(filters.Arg("reference", container.Image+":"+container.Version))
			isImageExists, _ := cli.DockerCli.Client().ImageList(ctx, types.ImageListOptions{All: true, Filters: imageFiler})

			if len(isImageExists) == 0 {
				eg.Go(func() error {
					w.Event(progress.Event{ID: container.Name, Status: progress.Working, Text: "Pulling"})

					stream, err := cli.DockerCli.Client().ImagePull(ctx, container.Image+":"+container.Version, types.ImagePullOptions{})
					if err != nil {
						w.TailMsgf(fmt.Sprint(err))
						w.Event(progress.ErrorEvent(container.Name))
						return nil
					}

					dec := json.NewDecoder(stream)
					for {
						var jm jsonmessage.JSONMessage
						if err := dec.Decode(&jm); err != nil {
							if errors.Is(err, io.EOF) {
								break
							}
							return err
						}
						if jm.Error != nil {
							return err
						}
						toPullProgressEvent(container.Name, jm, w)
					}

					w.Event(progress.Event{ID: container.Name, Status: progress.Done, Text: "Pulled"})
					return err
				})
			}
		}

		err := eg.Wait()
		if err != nil {
			return err
		}
		return err
	}, cli.DockerCli.Err())
}

func toPullProgressEvent(parent string, jm jsonmessage.JSONMessage, w progress.Writer) {
	if jm.ID == "" || jm.Progress == nil {
		return
	}

	var (
		text   string
		status = progress.Working
	)

	text = jm.Progress.String()

	if jm.Status == "Pull complete" ||
		jm.Status == "Already exists" ||
		strings.Contains(jm.Status, "Image is up to date") ||
		strings.Contains(jm.Status, "Downloaded newer image") {
		status = progress.Done
	}

	if jm.Error != nil {
		status = progress.Error
		text = jm.Error.Message
	}

	w.Event(progress.Event{
		ID:         jm.ID,
		ParentID:   parent,
		Text:       jm.Status,
		Status:     status,
		StatusText: text,
	})
}
