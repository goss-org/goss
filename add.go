package goss

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

// AddResources is a simple wrapper to add multiple resources
func AddResources(fileName, resourceName string, keys []string, c *util.Config) error {
	var err error
	outStoreFormat, err = getStoreFormatFromFileName(fileName)
	if err != nil {
		return err
	}

	var gossConfig GossConfig
	if _, err := os.Stat(fileName); err == nil {
		gossConfig, err = ReadJSON(fileName)
		if err != nil {
			return err
		}
	} else {
		gossConfig = *NewGossConfig()
	}

	sys := system.New(c.PackageManager)

	for _, key := range keys {
		if err := AddResource(fileName, gossConfig, resourceName, key, *c, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

// AddResource adds a single resource to fileName
func AddResource(fileName string, gossConfig GossConfig, resourceName, key string, config util.Config, sys *system.System) error {
	v := reflect.ValueOf(gossConfig)
	f := v.FieldByName(resourceName)
	fun := f.MethodByName("AppendSysResource")
	res := fun.Call([]reflect.Value{reflect.ValueOf(key), reflect.ValueOf(sys), reflect.ValueOf(config)})
	if err, ok := res[1].Interface().(error); ok && err != nil {
		return err
	}
	resourcePrint(fileName, res[0].Interface().(resource.ResourceRead), config.AnnounceToCLI)

	return nil
}

// AutoAddResources is a simple wrapper to add multiple resources
func AutoAddResources(fileName string, keys []string, c *util.Config) error {
	var err error
	outStoreFormat, err = getStoreFormatFromFileName(fileName)
	if err != nil {
		return err
	}

	var gossConfig GossConfig
	if _, err = os.Stat(fileName); err == nil {
		gossConfig, err = ReadJSON(fileName)
		if err != nil {
			return err
		}
	} else {
		gossConfig = *NewGossConfig()
	}

	sys := system.New(c.PackageManager)

	for _, key := range keys {
		if err := AutoAddResource(fileName, gossConfig, key, c, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

// AutoAddResource adds a single resource to fileName with automatic detection of the type of resource
func AutoAddResource(fileName string, gossConfig GossConfig, key string, c *util.Config, sys *system.System) error {
	// file
	if strings.Contains(key, "/") {
		if res, _, ok := gossConfig.Files.AppendSysResourceIfExists(key, sys); ok {
			resourcePrint(fileName, res, c.AnnounceToCLI)
		}
	}

	// group
	if res, _, ok := gossConfig.Groups.AppendSysResourceIfExists(key, sys); ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// package
	if res, _, ok := gossConfig.Packages.AppendSysResourceIfExists(key, sys); ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// port
	if res, _, ok := gossConfig.Ports.AppendSysResourceIfExists(key, sys); ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// process
	res, sysres, ok, err := gossConfig.Processes.AppendSysResourceIfExists(key, sys)
	if err != nil {
		return err
	}
	if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
		ports := system.GetPorts(true)
		pids, _ := sysres.Pids()
		for _, pid := range pids {
			pidS := strconv.Itoa(pid)
			for port, entries := range ports {
				for _, entry := range entries {
					if entry.Pid == pidS {
						// port
						if res, _, ok := gossConfig.Ports.AppendSysResourceIfExists(port, sys); ok {
							resourcePrint(fileName, res, c.AnnounceToCLI)
						}
					}
				}
			}
		}
	}

	// Service
	if res, _, ok := gossConfig.Services.AppendSysResourceIfExists(key, sys); ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// user
	if res, _, ok := gossConfig.Users.AppendSysResourceIfExists(key, sys); ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	return nil
}
