// +build genny

package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=$GOFILE -out=resource_list.go gen "ResourceType=Addr,Command,DNS,File,Gossfile,Group,Package,Port,Process,Service,User,KernelParam,Mount,Interface,HTTP,DiskUsage"
//go:generate sed -i -e "/^\\/\\/ +build genny/d" resource_list.go
//go:generate goimports -w resource_list.go resource_list.go

type ResourceType generic.Type

type ResourceTypeMap map[string]*ResourceType

func (r ResourceTypeMap) AppendSysResource(sr string, sys *system.System, config util.Config) (*ResourceType, error) {
	sysres := sys.NewResourceType(sr, sys, config)
	res, err := NewResourceType(sysres, config)
	if err != nil {
		return nil, err
	}
	if old_res, ok := r[res.ID()]; ok {
		res.Title = old_res.Title
		res.Meta = old_res.Meta
	}
	r[res.ID()] = res
	return res, nil
}

func (r ResourceTypeMap) AppendSysResourceIfExists(sr string, sys *system.System) (*ResourceType, system.ResourceType, bool) {
	sysres := sys.NewResourceType(sr, sys, util.Config{})
	// FIXME: Do we want to be silent about errors?
	res, _ := NewResourceType(sysres, util.Config{})
	if e, _ := sysres.Exists(); e != true {
		return res, sysres, false
	}
	if old_res, ok := r[res.ID()]; ok {
		res.Title = old_res.Title
		res.Meta = old_res.Meta
	}
	r[res.ID()] = res
	return res, sysres, true
}

func (ret *ResourceTypeMap) UnmarshalJSON(data []byte) error {
	// Curried json.Unmarshal
	unmarshal := func(i interface{}) error {
		if err := json.Unmarshal(data, i); err != nil {
			return err
		}
		return nil
	}

	// Validate configuration
	zero := ResourceType{}
	whitelist, err := util.WhitelistAttrs(zero, util.JSON)
	if err != nil {
		return err
	}
	if err := util.ValidateSections(unmarshal, zero, whitelist); err != nil {
		return err
	}

	var tmp map[string]*ResourceType
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	typ := reflect.TypeOf(zero)
	typs := strings.Split(typ.String(), ".")[1]
	for id, res := range tmp {
		if res == nil {
			return fmt.Errorf("Could not parse resource %s:%s", typs, id)
		}
		res.SetID(id)
	}

	*ret = tmp
	return nil
}

func (ret *ResourceTypeMap) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	// Validate configuration
	zero := ResourceType{}
	whitelist, err := util.WhitelistAttrs(zero, util.YAML)
	if err != nil {
		return err
	}
	if err := util.ValidateSections(unmarshal, zero, whitelist); err != nil {
		return err
	}

	var tmp map[string]*ResourceType
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	typ := reflect.TypeOf(zero)
	typs := strings.Split(typ.String(), ".")[1]
	for id, res := range tmp {
		if res == nil {
			return fmt.Errorf("Could not parse resource %s:%s", typs, id)
		}
		res.SetID(id)
	}

	*ret = tmp
	return nil
}
