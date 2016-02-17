// +build genny

package resource

import (
	"encoding/json"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=$GOFILE -out=resource_list.go gen "ResourceType=Addr,Command,DNS,File,Gossfile,Group,Package,Port,Process,Service,User"
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
	r[res.ID()] = res
	return res, sysres, true
}

func (r *ResourceTypeMap) UnmarshalJSON(data []byte) error {
	var tmp map[string]*ResourceType
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	for id, res := range tmp {
		res.SetID(id)
	}

	*r = tmp

	return nil
}

//func (r *ResourceTypeMap) UnmarshalYAML(data []byte) error {
func (r *ResourceTypeMap) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var tmp map[string]*ResourceType
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	for id, res := range tmp {
		res.SetID(id)
	}

	*r = tmp

	return nil
}
