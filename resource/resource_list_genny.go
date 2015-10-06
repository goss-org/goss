// +build genny

package resource

import (
	"reflect"

	"github.com/aelsabbahy/goss/system"
	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=$GOFILE -out=resource_list.go gen "ResourceType=Addr,Command,DNS,File,Gossfile,Group,Package,Port,Process,Service,User"
//go:generate sed -i -e "/^\\/\\/ +build genny/d" resource_list.go

type ResourceType generic.Type

type ResourceTypeSlice []*ResourceType

func (r *ResourceTypeSlice) Append(ne *ResourceType) bool {
	for _, ele := range *r {
		if reflect.DeepEqual(ele, ne) {
			return false
		}
	}
	*r = append(*r, ne)
	return true
}

func (r *ResourceTypeSlice) AppendSysResource(sr string, sys *system.System) (*ResourceType, system.ResourceType, bool) {
	sysres := sys.NewResourceType(sr, sys)
	res := NewResourceType(sysres)
	ok := r.Append(res)
	return res, sysres, ok
}

func (r *ResourceTypeSlice) AppendSysResourceIfExists(sr string, sys *system.System) (*ResourceType, system.ResourceType, bool) {
	sysres := sys.NewResourceType(sr, sys)
	res := NewResourceType(sysres)
	if e, _ := sysres.Exists(); e != true {
		return res, sysres, false
	}
	ok := r.Append(res)
	return res, sysres, ok
}
