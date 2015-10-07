package goss

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/codegangsta/cli"
)

type ConfigJSON struct {
	Files     resource.FileSlice     `json:"files,omitempty"`
	Packages  resource.PackageSlice  `json:"packages,omitempty"`
	Addrs     resource.AddrSlice     `json:"addrs,omitempty"`
	Ports     resource.PortSlice     `json:"ports,omitempty"`
	Services  resource.ServiceSlice  `json:"services,omitempty"`
	Users     resource.UserSlice     `json:"users,omitempty"`
	Groups    resource.GroupSlice    `json:"groups,omitempty"`
	Commands  resource.CommandSlice  `json:"commands,omitempty"`
	DNS       resource.DNSSlice      `json:"dns,omitempty"`
	Processes resource.ProcessSlice  `json:"processes,omitempty"`
	Gossfiles resource.GossfileSlice `json:"gossfiles,omitempty"`
}

func (c *ConfigJSON) String() string {
	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}
	return string(jsonData)
}

func (c *ConfigJSON) Resources() []resource.Resource {
	var tests []resource.Resource
	gs := genericConcatSlices(c.Commands, c.Addrs, c.DNS, c.Packages, c.Services, c.Files, c.Processes, c.Users, c.Groups, c.Ports)
	for _, t := range gs {
		tests = append(tests, t.(resource.Resource))
	}
	return tests
}

func genericConcatSlices(slices ...interface{}) []interface{} {
	var ret []interface{}
	for _, slice := range slices {
		is := interfaceSlice(slice)
		for _, x := range is {
			ret = append(ret, x)
		}
	}
	return ret
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func ReadJSON(filePath string) ConfigJSON {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	return ReadJSONData(file)
}

func ReadJSONData(data []byte) ConfigJSON {
	var configJSON ConfigJSON
	err := json.Unmarshal(data, &configJSON)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return configJSON
}

func RenderJSON(filePath string) string {
	path := filepath.Dir(filePath)
	configJSON := mergeJSONData(ReadJSON(filePath), 0, path)

	return configJSON.String()
}

func mergeJSONData(configJSON ConfigJSON, depth int, path string) ConfigJSON {
	depth++
	if depth >= 50 {
		fmt.Println("Error: Max depth of 50 reached, possibly due to dependency loop in goss file")
		os.Exit(1)
	}

	for _, gossfile := range configJSON.Gossfiles {
		fpath := filepath.Join(path, gossfile.Path)
		fdir := filepath.Dir(fpath)
		j := mergeJSONData(ReadJSON(fpath), depth, fdir)
		configJSON = mergeGoss(configJSON, j)
	}
	return configJSON
}

func mergeGoss(g1, g2 ConfigJSON) ConfigJSON {
	g1.Gossfiles = nil

	g1.Files.Append(g2.Files...)

	g1.Packages.Append(g2.Packages...)

	g1.Addrs.Append(g2.Addrs...)

	g1.Ports.Append(g2.Ports...)

	g1.Services.Append(g2.Services...)

	g1.Users.Append(g2.Users...)

	g1.Groups.Append(g2.Groups...)

	g1.Commands.Append(g2.Commands...)

	g1.DNS.Append(g2.DNS...)

	g1.Processes.Append(g2.Processes...)

	return g1
}

func WriteJSON(filePath string, configJSON ConfigJSON) error {
	jsonData, err := json.MarshalIndent(configJSON, "", "    ")
	if err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}

	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}

	return nil
}

func AppendResource(fileName, resourceName, key string, c *cli.Context) error {
	var configJSON ConfigJSON
	if _, err := os.Stat(fileName); err == nil {
		configJSON = ReadJSON(fileName)
	} else {
		configJSON = ConfigJSON{}
	}

	sys := system.New(c)

	// Need to figure out a good way to refactor this
	switch resourceName {
	case "Addr":
		if res, _, ok := configJSON.Addrs.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Command":
		if res, _, ok := configJSON.Commands.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "DNS":
		if res, _, ok := configJSON.DNS.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "File":
		if res, _, ok := configJSON.Files.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Group":
		if res, _, ok := configJSON.Groups.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Package":
		if res, _, ok := configJSON.Packages.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Port":
		if res, _, ok := configJSON.Ports.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Process":
		if res, _, ok := configJSON.Processes.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Service":
		if res, _, ok := configJSON.Services.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "User":
		if res, _, ok := configJSON.Users.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	case "Gossfile":
		if res, _, ok := configJSON.Gossfiles.AppendSysResource(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	}

	WriteJSON(fileName, configJSON)

	return nil
}

func AutoAppendResource(fileName, key string, c *cli.Context) error {
	var configJSON ConfigJSON
	if _, err := os.Stat(fileName); err == nil {
		configJSON = ReadJSON(fileName)
	} else {
		configJSON = ConfigJSON{}
	}

	sys := system.New(c)

	// file
	if strings.Contains(key, "/") {
		if res, _, ok := configJSON.Files.AppendSysResourceIfExists(key, sys); ok == true {
			resourcePrint(fileName, res)
		}
	}

	// group
	if res, _, ok := configJSON.Groups.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// package
	if res, _, ok := configJSON.Packages.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// port
	if res, _, ok := configJSON.Ports.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// process
	if res, sysres, ok := configJSON.Processes.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
		ports := system.GetPorts(true)
		pids, _ := sysres.Pids()
		for _, pid := range pids {
			pidS := strconv.Itoa(pid)
			for port, entry := range ports {
				if entry.Pid == pidS {
					// port
					if res, _, ok := configJSON.Ports.AppendSysResourceIfExists(port, sys); ok == true {
						resourcePrint(fileName, res)
					}
				}
			}
		}
	}

	// Service
	if res, _, ok := configJSON.Services.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	// user
	if res, _, ok := configJSON.Users.AppendSysResourceIfExists(key, sys); ok == true {
		resourcePrint(fileName, res)
	}

	WriteJSON(fileName, configJSON)

	return nil
}

func resourcePrint(fileName string, resource interface{}) {
	out, _ := json.MarshalIndent(resource, "", "    ")
	fmt.Printf("Adding %T to '%s':\n\n%s\n\n", resource, fileName, string(out))
}
