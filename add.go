package goss

import (
	"fmt"
	"os"
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
	var res resource.ResourceRead
	var err error

	// Need to figure out a good way to refactor this
	switch resourceName {
	case "Addr":
		res, err = gossConfig.Addrs.AppendSysResource(key, sys, config)
	case "Command":
		res, err = gossConfig.Commands.AppendSysResource(key, sys, config)
	case "DNS":
		res, err = gossConfig.DNS.AppendSysResource(key, sys, config)
	case "File":
		res, err = gossConfig.Files.AppendSysResource(key, sys, config)
	case "Group":
		res, err = gossConfig.Groups.AppendSysResource(key, sys, config)
	case "Package":
		res, err = gossConfig.Packages.AppendSysResource(key, sys, config)
	case "Port":
		res, err = gossConfig.Ports.AppendSysResource(key, sys, config)
	case "Process":
		res, err = gossConfig.Processes.AppendSysResource(key, sys, config)
	case "Service":
		res, err = gossConfig.Services.AppendSysResource(key, sys, config)
	case "User":
		res, err = gossConfig.Users.AppendSysResource(key, sys, config)
	case "Gossfile":
		res, err = gossConfig.Gossfiles.AppendSysResource(key, sys, config)
	case "KernelParam":
		res, err = gossConfig.KernelParams.AppendSysResource(key, sys, config)
	case "Mount":
		res, err = gossConfig.Mounts.AppendSysResource(key, sys, config)
	case "Interface":
		res, err = gossConfig.Interfaces.AppendSysResource(key, sys, config)
	case "HTTP":
		res, err = gossConfig.HTTPs.AppendSysResource(key, sys, config)
	default:
		err = fmt.Errorf("undefined resource name: %s", resourceName)
	}

	if err != nil {
		return err
	}

	resourcePrint(fileName, res, config.AnnounceToCLI)

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
		res, _, ok, err := gossConfig.Files.AppendSysResourceIfExists(key, sys)
		if err != nil {
			return err
		}
		if ok {
			resourcePrint(fileName, res, c.AnnounceToCLI)
		}
	}

	// group
	if res, _, ok, err := gossConfig.Groups.AppendSysResourceIfExists(key, sys); err != nil {
		return err

	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// package
	if res, _, ok, err := gossConfig.Packages.AppendSysResourceIfExists(key, sys); err != nil {

		return err

	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// port
	if res, _, ok, err := gossConfig.Ports.AppendSysResourceIfExists(key, sys); err != nil {
		return err

	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// process
	if res, sysres, ok, err := gossConfig.Processes.AppendSysResourceIfExists(key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
		ports := system.GetPorts(true)
		pids, _ := sysres.Pids()
		for _, pid := range pids {
			pidS := strconv.Itoa(pid)
			for port, entries := range ports {
				for _, entry := range entries {
					if entry.Pid == pidS {
						// port
						if res, _, ok, err := gossConfig.Ports.AppendSysResourceIfExists(port, sys); err != nil {
							return err
						} else if ok {
							resourcePrint(fileName, res, c.AnnounceToCLI)
						}
					}
				}
			}
		}
	}

	// Service
	if res, _, ok, err := gossConfig.Services.AppendSysResourceIfExists(key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// user
	if res, _, ok, err := gossConfig.Users.AppendSysResourceIfExists(key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	return nil
}
