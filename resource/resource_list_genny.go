// +build genny

package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=$GOFILE -out=resource_list.go gen "ResourceType=Addr,Command,DNS,File,Gossfile,Group,Package,Port,Process,Service,User"
//go:generate sed -i -e "/^\\/\\/ +build genny/d" resource_list.go

type ResourceType generic.Type

type ResourceTypeMap map[string]*ResourceType

func (r ResourceTypeMap) AppendSysResource(sr string, sys *system.System) (*ResourceType, system.ResourceType) {
	sysres := sys.NewResourceType(sr, sys)
	res := NewResourceType(sysres)
	r[res.ID()] = res
	return res, sysres
}

func (r ResourceTypeMap) AppendSysResourceIfExists(sr string, sys *system.System) (*ResourceType, system.ResourceType, bool) {
	sysres := sys.NewResourceType(sr, sys)
	res := NewResourceType(sysres)
	if e, _ := sysres.Exists(); e != true {
		return res, sysres, false
	}
	r[res.ID()] = res
	return res, sysres, true
}
