package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type DockerContainer struct {
	Title         string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta          meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	ContainerName string  `json:"-" yaml:"-"`
	Running       matcher `json:"running" yaml:"running"`
}

func (cnt *DockerContainer) ID() string      { return cnt.ContainerName }
func (cnt *DockerContainer) SetID(id string) { cnt.ContainerName = id }

func (cnt *DockerContainer) GetTitle() string { return cnt.Title }
func (cnt *DockerContainer) GetMeta() meta    { return cnt.Meta }

func (cnt *DockerContainer) Validate(sys *system.System) []TestResult {
	skip := false
	sysDockerContainer := sys.NewDockerContainer(cnt.ContainerName, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(cnt, "running", cnt.Running, sysDockerContainer.Running, skip))
	return results
}

func NewDockerContainer(sysDockerContainer system.DockerContainer, config util.Config) (*DockerContainer, error) {
	container_name := sysDockerContainer.ContainerName()
	running, _ := sysDockerContainer.Running()
	return &DockerContainer{
		ContainerName: container_name,
		Running:       running,
	}, nil
}
