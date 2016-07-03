package goss

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/urfave/cli"
)

func AddResource(fileName, resourceName, key string, c *cli.Context) error {
	setStoreFormatFromFileName(fileName)
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

	// Need to figure out a good way to refactor this
	switch resourceName {
	case "Addr":
		res, err := gossConfig.Addrs.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Command":
		res, err := gossConfig.Commands.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "DNS":
		res, err := gossConfig.DNS.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "File":
		res, err := gossConfig.Files.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Group":
		res, err := gossConfig.Groups.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Package":
		res, err := gossConfig.Packages.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Port":
		res, err := gossConfig.Ports.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Process":
		res, err := gossConfig.Processes.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Service":
		res, err := gossConfig.Services.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "User":
		res, err := gossConfig.Users.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Gossfile":
		res, err := gossConfig.Gossfiles.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "KernelParam":
		res, err := gossConfig.KernelParams.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	case "Mount":
		res, err := gossConfig.Mounts.AppendSysResource(key, sys, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resourcePrint(fileName, res)
	}

	WriteJSON(fileName, gossConfig)

	return nil
}

func AutoAddResource(fileName, key string, c *cli.Context) error {
	setStoreFormatFromFileName(fileName)
	var gossConfig GossConfig
	if _, err := os.Stat(fileName); err == nil {
		gossConfig = ReadJSON(fileName)
	} else {
		gossConfig = *NewGossConfig()
	}

	sys := system.New(c)

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

	WriteJSON(fileName, gossConfig)

	return nil
}
