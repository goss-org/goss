package goss

import (
	"reflect"

	"github.com/aelsabbahy/goss/resource"
)

type GossConfig struct {
	Files     resource.FileMap     `json:"file,omitempty" yaml:"file,omitempty"`
	Packages  resource.PackageMap  `json:"package,omitempty" yaml:"package,omitempty"`
	Addrs     resource.AddrMap     `json:"addr,omitempty" yaml:"addr,omitempty"`
	Ports     resource.PortMap     `json:"port,omitempty" yaml:"port,omitempty"`
	Services  resource.ServiceMap  `json:"service,omitempty" yaml:"service,omitempty"`
	Users     resource.UserMap     `json:"user,omitempty" yaml:"user,omitempty"`
	Groups    resource.GroupMap    `json:"group,omitempty" yaml:"group,omitempty"`
	Commands  resource.CommandMap  `json:"command,omitempty" yaml:"command,omitempty"`
	DNS       resource.DNSMap      `json:"dns,omitempty" yaml:"dns,omitempty"`
	Processes resource.ProcessMap  `json:"process,omitempty" yaml:"process,omitempty"`
	Gossfiles resource.GossfileMap `json:"gossfile,omitempty" yaml:"gossfile,omitempty"`
}

func NewGossConfig() *GossConfig {
	return &GossConfig{
		Files:     make(resource.FileMap),
		Packages:  make(resource.PackageMap),
		Addrs:     make(resource.AddrMap),
		Ports:     make(resource.PortMap),
		Services:  make(resource.ServiceMap),
		Users:     make(resource.UserMap),
		Groups:    make(resource.GroupMap),
		Commands:  make(resource.CommandMap),
		DNS:       make(resource.DNSMap),
		Processes: make(resource.ProcessMap),
		Gossfiles: make(resource.GossfileMap),
	}
}

func (c *GossConfig) Resources() []resource.Resource {
	var tests []resource.Resource

	gm := genericConcatMaps(c.Commands, c.Addrs, c.DNS, c.Packages, c.Services, c.Files, c.Processes, c.Users, c.Groups, c.Ports)
	for _, m := range gm {
		for _, t := range m {
			// FIXME: Can this be moved to a safer compile-time check?
			tests = append(tests, t.(resource.Resource))
		}
	}

	return tests
}

func genericConcatMaps(maps ...interface{}) (ret []map[string]interface{}) {
	for _, slice := range maps {
		im := interfaceMap(slice)
		ret = append(ret, im)
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

func mergeGoss(g1, g2 GossConfig) GossConfig {
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
