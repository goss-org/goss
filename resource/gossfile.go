package resource

import "github.com/aelsabbahy/goss/system"

type Gossfile struct {
	Path string `json:"path"`
}

func NewGossfile(sysGossfile system.Gossfile) *Gossfile {
	path := sysGossfile.Path()
	return &Gossfile{
		Path: path,
	}
}
