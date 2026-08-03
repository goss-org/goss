package goss

import (
	"fmt"
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

// Merge consumes all the resources in g2 into c. Duplicate resources in g2
// will overwrite the ones already present in c; a warning describing each
// duplicate is returned to the caller so it can be emitted at the
// appropriate architectural boundary (where a *util.Config and therefore a
// Logger is available).
//
// Merge itself performs no logging: keeping this function pure makes it
// trivially testable and decouples core config merging from the goss
// logging infrastructure.
func (c *GossConfig) Merge(g2 GossConfig) []string {
	var warnings []string
	collect := func(w string) {
		if w != "" {
			warnings = append(warnings, w)
		}
	}

	for k, v := range g2.Files {
		collect(mergeType(c.Files, "file", k, v))
	}
	for k, v := range g2.Packages {
		collect(mergeType(c.Packages, "package", k, v))
	}
	for k, v := range g2.Addrs {
		collect(mergeType(c.Addrs, "addr", k, v))
	}
	for k, v := range g2.Ports {
		collect(mergeType(c.Ports, "port", k, v))
	}
	for k, v := range g2.Services {
		collect(mergeType(c.Services, "service", k, v))
	}
	for k, v := range g2.Users {
		collect(mergeType(c.Users, "user", k, v))
	}
	for k, v := range g2.Groups {
		collect(mergeType(c.Groups, "group", k, v))
	}
	for k, v := range g2.Commands {
		collect(mergeType(c.Commands, "command", k, v))
	}
	for k, v := range g2.DNS {
		collect(mergeType(c.DNS, "dns", k, v))
	}
	for k, v := range g2.Processes {
		collect(mergeType(c.Processes, "process", k, v))
	}
	for k, v := range g2.KernelParams {
		collect(mergeType(c.KernelParams, "kernel-param", k, v))
	}
	for k, v := range g2.Mounts {
		collect(mergeType(c.Mounts, "mount", k, v))
	}
	for k, v := range g2.Interfaces {
		collect(mergeType(c.Interfaces, "interface", k, v))
	}
	for k, v := range g2.HTTPs {
		collect(mergeType(c.HTTPs, "http", k, v))
	}
	for k, v := range g2.Matchings {
		collect(mergeType(c.Matchings, "matching", k, v))
	}

	return warnings
}

// mergeType inserts v into m at key k, returning a non-empty warning string
// describing a duplicate if one was overwritten. This function performs no
// logging; the caller (ultimately a component at the edge layer that has
// access to a Logger) is responsible for emitting the returned warnings.
func mergeType[V any](m map[string]V, t, k string, v V) string {
	_, duplicate := m[k]
	m[k] = v
	if duplicate {
		return fmt.Sprintf("[WARN] Duplicate key detected: '%s: %s'. The value from a later-loaded goss file has overwritten the previous value.", t, k)
	}
	return ""
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

// mergeGoss merges g2 into g1, discarding g1.Gossfiles first so the
// recursion boundary in mergeJSONData is respected. Warnings about
// duplicate keys are returned rather than logged; the caller (which has a
// *util.Config in scope) is expected to emit them via c.Log().
func mergeGoss(g1, g2 GossConfig) (GossConfig, []string) {
	g1.Gossfiles = nil
	warnings := g1.Merge(g2)
	return g1, warnings
}
