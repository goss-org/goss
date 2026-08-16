package goss

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// AddResources is a simple wrapper to add multiple resources
func AddResources(ctx context.Context, fileName, resourceName string, keys []string, c *util.Config) error {
	if err := setLogLevel(c); err != nil {
		return err
	}
	format, err := getStoreFormatFromFileName(fileName)
	if err != nil {
		return err
	}
	setStoreFormat(format)

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
		if err := AddResource(ctx, fileName, gossConfig, resourceName, key, *c, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

// AddResource adds a single resource to fileName
func AddResource(ctx context.Context, fileName string, gossConfig GossConfig, resourceName, key string, config util.Config, sys *system.System) error {
	var err error
	var res resource.ResourceRead

	// Need to figure out a good way to refactor this
	switch resourceName {
	case resource.AddResourceName:
		res, err = gossConfig.Addrs.AppendSysResource(ctx, key, sys, config)
	case resource.CommandResourceName:
		res, err = gossConfig.Commands.AppendSysResource(ctx, key, sys, config)
	case resource.DNSResourceName:
		res, err = gossConfig.DNS.AppendSysResource(ctx, key, sys, config)
	case resource.FileResourceName:
		res, err = gossConfig.Files.AppendSysResource(ctx, key, sys, config)
	case resource.GroupResourceName:
		res, err = gossConfig.Groups.AppendSysResource(ctx, key, sys, config)
	case resource.PackageResourceName:
		res, err = gossConfig.Packages.AppendSysResource(ctx, key, sys, config)
	case resource.PortResourceName:
		res, err = gossConfig.Ports.AppendSysResource(ctx, key, sys, config)
	case resource.ProcessResourceName:
		res, err = gossConfig.Processes.AppendSysResource(ctx, key, sys, config)
	case resource.ServiceResourceName:
		res, err = gossConfig.Services.AppendSysResource(ctx, key, sys, config)
	case resource.UserResourceName:
		res, err = gossConfig.Users.AppendSysResource(ctx, key, sys, config)
	case resource.GossFileResourceName:
		res, err = gossConfig.Gossfiles.AppendSysResource(ctx, key, sys, config)
	case resource.KernelParamResourceName:
		res, err = gossConfig.KernelParams.AppendSysResource(ctx, key, sys, config)
	case resource.MountResourceName:
		res, err = gossConfig.Mounts.AppendSysResource(ctx, key, sys, config)
	case resource.InterfaceResourceName:
		res, err = gossConfig.Interfaces.AppendSysResource(ctx, key, sys, config)
	case resource.HTTPResourceName:
		res, err = gossConfig.HTTPs.AppendSysResource(ctx, key, sys, config)
	case resource.RegistryResourceName:
		res, err = gossConfig.Registries.AppendSysResource(ctx, key, sys, config)
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
func AutoAddResources(ctx context.Context, fileName string, keys []string, c *util.Config) error {
	format, err := getStoreFormatFromFileName(fileName)
	if err != nil {
		return err
	}
	setStoreFormat(format)

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
		if err := AutoAddResource(ctx, fileName, gossConfig, key, c, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

// AutoAddResource adds a single resource to fileName with automatic detection of the type of resource
func AutoAddResource(ctx context.Context, fileName string, gossConfig GossConfig, key string, c *util.Config, sys *system.System) error {
	// file
	if strings.Contains(key, "/") {
		res, _, ok, err := gossConfig.Files.AppendSysResourceIfExists(ctx, key, sys)
		if err != nil {
			return err
		}
		if ok {
			resourcePrint(fileName, res, c.AnnounceToCLI)
		}
	}

	// group
	if res, _, ok, err := gossConfig.Groups.AppendSysResourceIfExists(ctx, key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// package
	if res, _, ok, err := gossConfig.Packages.AppendSysResourceIfExists(ctx, key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// port
	if res, _, ok, err := gossConfig.Ports.AppendSysResourceIfExists(ctx, key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// process
	if res, sysres, ok, err := gossConfig.Processes.AppendSysResourceIfExists(ctx, key, sys); err != nil {
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
						if res, _, ok, err := gossConfig.Ports.AppendSysResourceIfExists(ctx, port, sys); err != nil {
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
	if res, _, ok, err := gossConfig.Services.AppendSysResourceIfExists(ctx, key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	// user
	if res, _, ok, err := gossConfig.Users.AppendSysResourceIfExists(ctx, key, sys); err != nil {
		return err
	} else if ok {
		resourcePrint(fileName, res, c.AnnounceToCLI)
	}

	return nil
}
