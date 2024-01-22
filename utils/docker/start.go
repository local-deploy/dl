package docker

import (
	"context"
	"time"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/api"
)

type composeOptions struct {
	*compose.ProjectOptions
}

type upOptions struct { //nolint:maligned
	*composeOptions
	Detach             bool
	noStart            bool
	noDeps             bool
	cascadeStop        bool
	exitCodeFrom       string
	noColor            bool
	noPrefix           bool
	attachDependencies bool
	attach             []string
	noAttach           []string
	timestamp          bool
	wait               bool
	waitTimeout        int
}

// StartContainers running docker containers
func (cli *Client) StartContainers(ctx context.Context, project *types.Project, recreate bool) error {
	var (
		services []string
		consumer api.LogConsumer
	)

	up := upOptions{}
	opt := createOptions{}

	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:    project.Name,
			api.ServiceLabel:    s.Name,
			api.VersionLabel:    api.ComposeVersion,
			api.WorkingDirLabel: project.WorkingDir,
			api.OneoffLabel:     "False", // default, will be overridden by `run` command
		}
		project.Services[i] = s
	}

	err := opt.apply(project)
	if err != nil {
		return err
	}

	err = up.apply(project, services)
	if err != nil {
		return err
	}

	create := api.CreateOptions{
		Services:             services,
		RemoveOrphans:        false,
		IgnoreOrphans:        false,
		Recreate:             api.RecreateDiverged,
		RecreateDependencies: api.RecreateDiverged,
		Inherit:              true,
		QuietPull:            false,
	}

	if recreate {
		create.Recreate = api.RecreateForce
	}

	timeout := time.Duration(up.waitTimeout) * time.Second
	start := api.StartOptions{
		Project:      project,
		Attach:       consumer,
		ExitCodeFrom: up.exitCodeFrom,
		CascadeStop:  up.cascadeStop,
		Wait:         up.wait,
		WaitTimeout:  timeout,
		Services:     services,
	}

	err = cli.Backend.Up(ctx, project, api.UpOptions{
		Create: create,
		Start:  start,
	})
	if err != nil {
		return err
	}

	return nil
}

func (opts upOptions) apply(project *types.Project, services []string) error {
	if opts.noDeps {
		err := project.ForServices(services, types.IgnoreDependencies)
		if err != nil {
			return err
		}
	}

	if opts.exitCodeFrom != "" {
		_, err := project.GetService(opts.exitCodeFrom)
		if err != nil {
			return err
		}
	}

	return nil
}
