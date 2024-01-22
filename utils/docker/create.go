package docker

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
	"golang.org/x/exp/slices"
)

type createOptions struct { //nolint:maligned
	Build         bool
	noBuild       bool
	Pull          string
	pullChanged   bool
	removeOrphans bool
	ignoreOrphans bool
	forceRecreate bool
	noRecreate    bool
	recreateDeps  bool
	noInherit     bool
	timeChanged   bool
	timeout       int
	quietPull     bool
	scale         []string
}

func (opts createOptions) apply(project *types.Project) error {
	if opts.pullChanged {
		for i, service := range project.Services {
			service.PullPolicy = opts.Pull
			project.Services[i] = service
		}
	}
	if opts.Build {
		for i, service := range project.Services {
			if service.Build == nil {
				continue
			}
			service.PullPolicy = types.PullPolicyBuild
			project.Services[i] = service
		}
	}
	if opts.noBuild {
		for i, service := range project.Services {
			service.Build = nil
			if service.Image == "" {
				service.Image = api.GetImageNameOrDefault(service, project.Name)
			}
			project.Services[i] = service
		}
	}

	return nil
}

func (opts createOptions) isPullPolicyValid() bool {
	pullPolicies := []string{types.PullPolicyAlways, types.PullPolicyNever, types.PullPolicyBuild,
		types.PullPolicyMissing, types.PullPolicyIfNotPresent}
	return slices.Contains(pullPolicies, opts.Pull)
}
