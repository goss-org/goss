package system

import (
	"fmt"
	"os"

	"github.com/aelsabbahy/goss/util"

	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type DockerContainer interface {
	ContainerName() string
	Exists() (bool, error)
	Running() (bool, error)
	Containers() ([]string, error)
}
type DefDockerContainer struct {
	container_name string
	cntMap         map[string][]docker.Container
}

func NewDefDockerContainer(name string, system *System, config util.Config) DockerContainer {
	return &DefDockerContainer{
		container_name: name,
		cntMap:         system.DockerContainerMap(),
	}
}

func (cnt *DefDockerContainer) ContainerName() string {
	return cnt.container_name
}

func (cnt *DefDockerContainer) Exists() (bool, error) {
	if _, ok := cnt.cntMap[cnt.container_name]; ok {
		return true, nil
	}
	return false, nil
}

func (cnt *DefDockerContainer) Running() (bool, error) {
	if _, ok := cnt.cntMap[cnt.container_name]; ok {
		return true, nil
		// return (cnt.cntMap[cnt.container_name].State == "running"), nil
	}
	return false, nil
}

func (cnt *DefDockerContainer) Containers() ([]string, error) {
	var cnts []string
	for _, cnt := range cnt.cntMap[cnt.container_name] {
		for _, name := range cnt.Names {
			cnts = append(cnts, name)
		}
	}
	return cnts, nil
}

func GetContainers() map[string][]docker.Container {
	cmap := make(map[string][]docker.Container)
	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	containers, err := cli.ContainerList(context.Background(), docker.ContainerListOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, cnt := range containers {
		for _, name := range cnt.Names {
			cmap[name] = append(cmap[name], cnt)
		}
	}
	return cmap

}
