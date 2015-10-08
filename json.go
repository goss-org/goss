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
	Files     resource.FileMap     `json:"file,omitempty"`
	Packages  resource.PackageMap  `json:"package,omitempty"`
	Addrs     resource.AddrMap     `json:"addr,omitempty"`
	Ports     resource.PortMap     `json:"port,omitempty"`
	Services  resource.ServiceMap  `json:"service,omitempty"`
	Users     resource.UserMap     `json:"user,omitempty"`
	Groups    resource.GroupMap    `json:"group,omitempty"`
	Commands  resource.CommandMap  `json:"command,omitempty"`
	DNS       resource.DNSMap      `json:"dns,omitempty"`
	Processes resource.ProcessMap  `json:"processe,omitempty"`
	Gossfiles resource.GossfileMap `json:"gossfile,omitempty"`
}

func NewConfigJSON() *ConfigJSON {
	return &ConfigJSON{
		Files:     make(resource.FileMap),
		Packages:  make(resource.PackageMap),
		Addrs:     make(resource.AddrMap),
		Ports:     make(resource.PortMap),
		Services:  make(resource.ServiceMap),
		Users:     make(resource.UserMap),
		Groups:    make(resource.GroupMap),
		Commands:  make(resource.CommandMap),
		Processes: make(resource.ProcessMap),
		Gossfiles: make(resource.GossfileMap),
	}
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
	gs := genericConcatMaps(c.Commands, c.Addrs, c.DNS, c.Packages, c.Services, c.Files, c.Processes, c.Users, c.Groups, c.Ports)
	for _, t := range gs {
		tests = append(tests, t.(resource.Resource))
	}

	return tests
}

func genericConcatMaps(maps ...interface{}) []interface{} {
	var ret []interface{}
	for _, slice := range maps {
		is := interfaceMap(slice)
		for _, x := range is {
			ret = append(ret, x)
		}
	}
	return ret
}

func interfaceMap(slice interface{}) map[string]interface{} {
	m := reflect.ValueOf(slice)
	if m.Kind() != reflect.Map {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make(map[string]interface{})

	for _, k := range m.MapKeys() {
		ret[k.Interface().(string)] = m.MapIndex(k).Interface()
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
	configJSON := NewConfigJSON()
	err := json.Unmarshal(data, configJSON)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return *configJSON
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

	for k, v := range g2.Files {
		g1.Files[k] = v
	}

	for k, v := range g2.Packages {
		g1.Packages[k] = v
	}

	for k, v := range g2.Addrs {
		g1.Addrs[k] = v
	}

	for k, v := range g2.Ports {
		g1.Ports[k] = v
	}

	for k, v := range g2.Services {
		g1.Services[k] = v
	}

	for k, v := range g2.Users {
		g1.Users[k] = v
	}

	for k, v := range g2.Groups {
		g1.Groups[k] = v
	}

	for k, v := range g2.Commands {
		g1.Commands[k] = v
	}

	for k, v := range g2.DNS {
		g1.DNS[k] = v
	}

	for k, v := range g2.Processes {
		g1.Processes[k] = v
	}

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
		configJSON = *NewConfigJSON()
	}

	sys := system.New(c)

	// Need to figure out a good way to refactor this
	switch resourceName {
	case "Addr":
		res, _ := configJSON.Addrs.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Command":
		res, _ := configJSON.Commands.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "DNS":
		res, _ := configJSON.DNS.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "File":
		res, _ := configJSON.Files.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Group":
		res, _ := configJSON.Groups.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Package":
		res, _ := configJSON.Packages.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Port":
		res, _ := configJSON.Ports.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Process":
		res, _ := configJSON.Processes.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Service":
		res, _ := configJSON.Services.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "User":
		res, _ := configJSON.Users.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	case "Gossfile":
		res, _ := configJSON.Gossfiles.AppendSysResource(key, sys)
		resourcePrint(fileName, res)
	}

	WriteJSON(fileName, configJSON)

	return nil
}

func AutoAppendResource(fileName, key string, c *cli.Context) error {
	var configJSON ConfigJSON
	if _, err := os.Stat(fileName); err == nil {
		configJSON = ReadJSON(fileName)
	} else {
		configJSON = *NewConfigJSON()
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
