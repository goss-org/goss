package goss

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

// Simple wrapper to add multiple resources
func AddResources(fileName, resourceName string, keys []string, c *RuntimeConfig) error {
	var err error
	outStoreFormat, err = getStoreFormatFromFileName(fileName)
	if err != nil {
		return err
	}

	config := util.Config{
		IgnoreList:        c.ExcludeAttributes,
		Timeout:           int(c.Timeout / time.Millisecond),
		AllowInsecure:     c.Insecure,
		NoFollowRedirects: c.NoFollowRedirects,
		Server:            c.Server,
		Username:          c.Username,
		Password:          c.Password,
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
		if err := AddResource(fileName, gossConfig, resourceName, key, config, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

func AddResource(fileName string, gossConfig GossConfig, resourceName, key string, config util.Config, sys *system.System) error {
	// Need to figure out a good way to refactor this
	switch resourceName {
	case "Addr":
		res, err := gossConfig.Addrs.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Command":
		res, err := gossConfig.Commands.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "DNS":
		res, err := gossConfig.DNS.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "File":
		res, err := gossConfig.Files.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Group":
		res, err := gossConfig.Groups.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Package":
		res, err := gossConfig.Packages.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Port":
		res, err := gossConfig.Ports.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Process":
		res, err := gossConfig.Processes.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Service":
		res, err := gossConfig.Services.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "User":
		res, err := gossConfig.Users.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Gossfile":
		res, err := gossConfig.Gossfiles.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "KernelParam":
		res, err := gossConfig.KernelParams.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Mount":
		res, err := gossConfig.Mounts.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "Interface":
		res, err := gossConfig.Interfaces.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	case "HTTP":
		res, err := gossConfig.HTTPs.AppendSysResource(key, sys, config)
		if err != nil {
			return err
		}
		resourcePrint(fileName, res)
	default:
		return fmt.Errorf("undefined resource name: %s", resourceName)
	}

	return nil
}

// Simple wrapper to add multiple resources
func AutoAddResources(fileName string, keys []string, c *RuntimeConfig) error {
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
		if err := AutoAddResource(fileName, gossConfig, key, sys); err != nil {
			return err
		}
	}

	return WriteJSON(fileName, gossConfig)
}

func AutoAddResource(fileName string, gossConfig GossConfig, key string, sys *system.System) error {
	// file
	if strings.Contains(key, "/") {
		if res, _, ok := gossConfig.Files.AppendSysResourceIfExists(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	}

	// group
	if res, _, ok := gossConfig.Groups.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// package
	if res, _, ok := gossConfig.Packages.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// port
	if res, _, ok := gossConfig.Ports.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// process
	res, sysres, ok, err := gossConfig.Processes.AppendSysResourceIfExists(key, sys)
	if err != nil {
		return err
	}
	if ok {
		resourcePrint(fileName, res)
		ports := system.GetPorts(true)
		pids, _ := sysres.Pids()
		for _, pid := range pids {
			pidS := strconv.Itoa(pid)
			for port, entries := range ports {
				for _, entry := range entries {
					if entry.Pid == pidS {
						// port
						if res, _, ok := gossConfig.Ports.AppendSysResourceIfExists(port, sys); ok == true {
							resourcePrint(fileName, res)
						}
					}
				}
			}
		}
	}

	// Service
	if res, _, ok := gossConfig.Services.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// user
	if res, _, ok := gossConfig.Users.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	return nil
}
