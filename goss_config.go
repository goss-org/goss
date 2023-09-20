package goss

import (
	"log"
	"reflect"

	"github.com/goss-org/goss/resource"
)

type GossConfig struct {
	Files        resource.FileMap        `json:"file,omitempty" yaml:"file,omitempty"`
	Packages     resource.PackageMap     `json:"package,omitempty" yaml:"package,omitempty"`
	Addrs        resource.AddrMap        `json:"addr,omitempty" yaml:"addr,omitempty"`
	Ports        resource.PortMap        `json:"port,omitempty" yaml:"port,omitempty"`
	Services     resource.ServiceMap     `json:"service,omitempty" yaml:"service,omitempty"`
	Users        resource.UserMap        `json:"user,omitempty" yaml:"user,omitempty"`
	Groups       resource.GroupMap       `json:"group,omitempty" yaml:"group,omitempty"`
	Commands     resource.CommandMap     `json:"command,omitempty" yaml:"command,omitempty"`
	DNS          resource.DNSMap         `json:"dns,omitempty" yaml:"dns,omitempty"`
	Processes    resource.ProcessMap     `json:"process,omitempty" yaml:"process,omitempty"`
	Gossfiles    resource.GossfileMap    `json:"gossfile,omitempty" yaml:"gossfile,omitempty"`
	KernelParams resource.KernelParamMap `json:"kernel-param,omitempty" yaml:"kernel-param,omitempty"`
	Mounts       resource.MountMap       `json:"mount,omitempty" yaml:"mount,omitempty"`
	Interfaces   resource.InterfaceMap   `json:"interface,omitempty" yaml:"interface,omitempty"`
	HTTPs        resource.HTTPMap        `json:"http,omitempty" yaml:"http,omitempty"`
	Matchings    resource.MatchingMap    `json:"matching,omitempty" yaml:"matching,omitempty"`
}

func NewGossConfig() *GossConfig {
	return &GossConfig{
		Files:        make(resource.FileMap),
		Packages:     make(resource.PackageMap),
		Addrs:        make(resource.AddrMap),
		Ports:        make(resource.PortMap),
		Services:     make(resource.ServiceMap),
		Users:        make(resource.UserMap),
		Groups:       make(resource.GroupMap),
		Commands:     make(resource.CommandMap),
		DNS:          make(resource.DNSMap),
		Processes:    make(resource.ProcessMap),
		Gossfiles:    make(resource.GossfileMap),
		KernelParams: make(resource.KernelParamMap),
		Mounts:       make(resource.MountMap),
		Interfaces:   make(resource.InterfaceMap),
		HTTPs:        make(resource.HTTPMap),
		Matchings:    make(resource.MatchingMap),
	}
}

// Merge consumes all the resources in g2 into c, duplicate resources
// will be overwritten with the ones in g2
func (c *GossConfig) Merge(g2 GossConfig) {
	for k, v := range g2.Files {
		mergeType(c.Files, "file", k, v)
	}

	for k, v := range g2.Packages {
		mergeType(c.Packages, "package", k, v)
	}

	for k, v := range g2.Addrs {
		mergeType(c.Addrs, "addr", k, v)
	}

	for k, v := range g2.Ports {
		mergeType(c.Ports, "port", k, v)
	}

	for k, v := range g2.Services {
		mergeType(c.Services, "service", k, v)
	}

	for k, v := range g2.Users {
		mergeType(c.Users, "user", k, v)
	}

	for k, v := range g2.Groups {
		mergeType(c.Groups, "group", k, v)
	}

	for k, v := range g2.Commands {
		mergeType(c.Commands, "command", k, v)
	}

	for k, v := range g2.DNS {
		mergeType(c.DNS, "dns", k, v)
	}

	for k, v := range g2.Processes {
		mergeType(c.Processes, "process", k, v)
	}

	for k, v := range g2.KernelParams {
		mergeType(c.KernelParams, "kernel-param", k, v)
	}

	for k, v := range g2.Mounts {
		mergeType(c.Mounts, "mount", k, v)
	}

	for k, v := range g2.Interfaces {
		mergeType(c.Interfaces, "interface", k, v)
	}

	for k, v := range g2.HTTPs {
		mergeType(c.HTTPs, "http", k, v)
	}

	for k, v := range g2.Matchings {
		mergeType(c.Matchings, "matching", k, v)
	}
}

func mergeType[V any](m map[string]V, t, k string, v V) {
		if _, ok := m[k]; ok {
			log.Printf("[WARN] Duplicate key detected: '%s: %s'. The value from a later-loaded goss file has overwritten the previous value.", t, k)
		}
		m[k] = v
}


func (c *GossConfig) Resources() []resource.Resource {
	var tests []resource.Resource

	gm := genericConcatMaps(c.Commands,
		c.HTTPs,
		c.Addrs,
		c.DNS,
		c.Packages,
		c.Services,
		c.Files,
		c.Processes,
		c.Users,
		c.Groups,
		c.Ports,
		c.KernelParams,
		c.Mounts,
		c.Interfaces,
		c.Matchings,
	)

	for _, m := range gm {
		for _, t := range m {
			// FIXME: Can this be moved to a safer compile-time check?
			tests = append(tests, t.(resource.Resource))
		}
	}

	return tests
}

func genericConcatMaps(maps ...any) (ret []map[string]any) {
	for _, slice := range maps {
		im := interfaceMap(slice)
		ret = append(ret, im)
	}
	return ret
}

func interfaceMap(slice any) map[string]any {
	m := reflect.ValueOf(slice)
	if m.Kind() != reflect.Map {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make(map[string]any)

	for _, k := range m.MapKeys() {
		ret[k.Interface().(string)] = m.MapIndex(k).Interface()
	}

	return ret
}

func mergeGoss(g1, g2 GossConfig) GossConfig {
	g1.Gossfiles = nil

	g1.Merge(g2)

	return g1
}
