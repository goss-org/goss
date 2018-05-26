package goss

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/urfave/cli"
)

// Simple wrapper to add multiple resources
func AddResources(fileName, resourceName string, keys []string, c *cli.Context) error {
	OutStoreFormat = getStoreFormatFromFileName(fileName)
	config := util.Config{
		IgnoreList:        c.GlobalStringSlice("exclude-attr"),
		Timeout:           int(c.Duration("timeout") / time.Millisecond),
		AllowInsecure:     c.Bool("insecure"),
		NoFollowRedirects: c.Bool("no-follow-redirects"),
		Server:            c.String("server"),
		Username:          c.String("username"),
		Password:          c.String("password"),
	}

	var gossConfig GossConfig
	if _, err := os.Stat(fileName); err == nil {
		gossConfig = ReadJSON(fileName)
	} else {
		gossConfig = *NewGossConfig()
	}

	sys := system.New(c)

	for _, key := range keys {
		if err := AddResource(fileName, gossConfig, resourceName, key, c, config, sys); err != nil {
			return err
		}
	}
	WriteJSON(fileName, gossConfig)

	return nil
}

func AddResource(fileName string, gossConfig GossConfig, resourceName, key string, c *cli.Context, config util.Config, sys *system.System) error {
	v := reflect.ValueOf(gossConfig)
	f := v.FieldByName(resourceName)
	fun := f.MethodByName("AppendSysResource")
	res := fun.Call([]reflect.Value{reflect.ValueOf(key), reflect.ValueOf(sys), reflect.ValueOf(config)})
	if err, ok := res[1].Interface().(error); ok && err != nil {
		return err
	}
	resourcePrint(fileName, res[0].Interface().(resource.ResourceRead))
	return nil
}

// Simple wrapper to add multiple resources
func AutoAddResources(fileName string, keys []string, c *cli.Context) error {
	OutStoreFormat = getStoreFormatFromFileName(fileName)
	config := util.Config{
		IgnoreList: c.GlobalStringSlice("exclude-attr"),
		Timeout:    int(c.Duration("timeout") / time.Millisecond),
	}

	var gossConfig GossConfig
	if _, err := os.Stat(fileName); err == nil {
		gossConfig = ReadJSON(fileName)
	} else {
		gossConfig = *NewGossConfig()
	}

	sys := system.New(c)

	for _, key := range keys {
		if err := AutoAddResource(fileName, gossConfig, key, c, config, sys); err != nil {
			return err
		}
	}
	WriteJSON(fileName, gossConfig)

	return nil
}

func AutoAddResource(fileName string, gossConfig GossConfig, key string, c *cli.Context, config util.Config, sys *system.System) error {
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
	if res, sysres, ok := gossConfig.Processes.AppendSysResourceIfExists(key, sys); ok == true {
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
