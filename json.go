package goss

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/codegangsta/cli"
)

type ConfigJSON struct {
	Files     []*resource.File     `json:"files,omitempty"`
	Packages  []*resource.Package  `json:"packages,omitempty"`
	Addrs     []*resource.Addr     `json:"addrs,omitempty"`
	Ports     []*resource.Port     `json:"ports,omitempty"`
	Services  []*resource.Service  `json:"services,omitempty"`
	Users     []*resource.User     `json:"users,omitempty"`
	Groups    []*resource.Group    `json:"groups,omitempty"`
	Commands  []*resource.Command  `json:"commands,omitempty"`
	DNS       []*resource.DNS      `json:"dns,omitempty"`
	Processes []*resource.Process  `json:"processes,omitempty"`
	Gossfiles []*resource.Gossfile `json:"gossfiles,omitempty"`
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

// FIXME: This code made me throw up a little in my mouth.. need to fix.
func mergeGoss(g1, g2 ConfigJSON) ConfigJSON {
	var configJSON ConfigJSON

	for _, x := range g1.Files {
		configJSON.Files = append(configJSON.Files, x)
	}
	for _, x := range g2.Files {
		configJSON.Files = append(configJSON.Files, x)
	}

	for _, x := range g1.Packages {
		configJSON.Packages = append(configJSON.Packages, x)
	}
	for _, x := range g2.Packages {
		configJSON.Packages = append(configJSON.Packages, x)
	}

	for _, x := range g1.Addrs {
		configJSON.Addrs = append(configJSON.Addrs, x)
	}
	for _, x := range g2.Addrs {
		configJSON.Addrs = append(configJSON.Addrs, x)
	}

	for _, x := range g1.Ports {
		configJSON.Ports = append(configJSON.Ports, x)
	}
	for _, x := range g2.Ports {
		configJSON.Ports = append(configJSON.Ports, x)
	}

	for _, x := range g1.Services {
		configJSON.Services = append(configJSON.Services, x)
	}
	for _, x := range g2.Services {
		configJSON.Services = append(configJSON.Services, x)
	}

	for _, x := range g1.Users {
		configJSON.Users = append(configJSON.Users, x)
	}
	for _, x := range g2.Users {
		configJSON.Users = append(configJSON.Users, x)
	}

	for _, x := range g1.Groups {
		configJSON.Groups = append(configJSON.Groups, x)
	}
	for _, x := range g2.Groups {
		configJSON.Groups = append(configJSON.Groups, x)
	}

	for _, x := range g1.Commands {
		configJSON.Commands = append(configJSON.Commands, x)
	}
	for _, x := range g2.Commands {
		configJSON.Commands = append(configJSON.Commands, x)
	}

	for _, x := range g1.DNS {
		configJSON.DNS = append(configJSON.DNS, x)
	}
	for _, x := range g2.DNS {
		configJSON.DNS = append(configJSON.DNS, x)
	}

	for _, x := range g1.Processes {
		configJSON.Processes = append(configJSON.Processes, x)
	}
	for _, x := range g2.Processes {
		configJSON.Processes = append(configJSON.Processes, x)
	}

	return configJSON
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
		sysResource := sys.NewAddr(key, sys)
		resource := resource.NewAddr(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Addrs = append(configJSON.Addrs, resource)
	case "Command":
		sysResource := sys.NewCommand(key, sys)
		resource := resource.NewCommand(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Commands = append(configJSON.Commands, resource)
	case "DNS":
		sysResource := sys.NewDNS(key, sys)
		resource := resource.NewDNS(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.DNS = append(configJSON.DNS, resource)
	case "File":
		sysResource := sys.NewFile(key, sys)
		resource := resource.NewFile(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Files = append(configJSON.Files, resource)
	case "Group":
		sysResource := sys.NewGroup(key, sys)
		resource := resource.NewGroup(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Groups = append(configJSON.Groups, resource)
	case "Package":
		sysResource := sys.NewPackage(key, sys)
		resource := resource.NewPackage(sysResource)
		resourcePrint(fileName, resource)
		configJSON.Packages = append(configJSON.Packages, resource)
	case "Port":
		sysResource := sys.NewPort(key, sys)
		resource := resource.NewPort(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Ports = append(configJSON.Ports, resource)
	case "Process":
		sysResource := sys.NewProcess(key, sys)
		resource := resource.NewProcess(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Processes = append(configJSON.Processes, resource)
	case "Service":
		sysResource := sys.NewService(key, sys)
		resource := resource.NewService(sysResource)
		resourcePrint(fileName, resource)
		configJSON.Services = append(configJSON.Services, resource)
	case "User":
		sysResource := sys.NewUser(key, sys)
		resource := resource.NewUser(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Users = append(configJSON.Users, resource)
	case "Gossfile":
		sysResource := sys.NewGossfile(key, sys)
		resource := resource.NewGossfile(*sysResource)
		resourcePrint(fileName, resource)
		configJSON.Gossfiles = append(configJSON.Gossfiles, resource)
	}

	WriteJSON(fileName, configJSON)

	return nil
}

func resourcePrint(fileName string, resource interface{}) {
	out, _ := json.MarshalIndent(resource, "", "    ")
	fmt.Printf("Adding to '%s':\n\n%s\n\n", fileName, string(out))
}
