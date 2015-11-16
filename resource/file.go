package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type File struct {
	Path     string   `json:"-"`
	Exists   bool     `json:"exists"`
	Mode     string   `json:"mode,omitempty"`
	Owner    string   `json:"owner,omitempty"`
	Group    string   `json:"group,omitempty"`
	LinkedTo string   `json:"linked-to,omitempty"`
	Filetype string   `json:"filetype,omitempty"`
	Contains []string `json:"contains"`
}

func (f *File) ID() string      { return f.Path }
func (f *File) SetID(id string) { f.Path = id }

func (f *File) Validate(sys *system.System) []TestResult {
	sysFile := sys.NewFile(f.Path, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(f, "exists", f.Exists, sysFile.Exists))

	if f.Mode != "" {
		results = append(results, ValidateValue(f, "mode", f.Mode, sysFile.Mode))
	}

	if f.Owner != "" {
		results = append(results, ValidateValue(f, "owner", f.Owner, sysFile.Owner))
	}

	if f.Group != "" {
		results = append(results, ValidateValue(f, "group", f.Group, sysFile.Group))
	}

	if f.LinkedTo != "" {
		results = append(results, ValidateValue(f, "linkedto", f.LinkedTo, sysFile.LinkedTo))
	}

	if f.Filetype != "" {
		results = append(results, ValidateValue(f, "filetype", f.Filetype, sysFile.Filetype))
	}

	if len(f.Contains) != 0 {
		results = append(results, ValidateContains(f, "contains", f.Contains, sysFile.Contains))
	}

	return results
}

func NewFile(sysFile system.File, config util.Config) (*File, error) {
	path := sysFile.Path()
	exists, _ := sysFile.Exists()
	f := &File{
		Path:     path,
		Exists:   exists.(bool),
		Contains: []string{},
	}
	if !contains(config.IgnoreList, "mode") {
		mode, _ := sysFile.Mode()
		f.Mode = mode.(string)
	}
	if !contains(config.IgnoreList, "owner") {
		owner, _ := sysFile.Owner()
		f.Owner = owner.(string)
	}
	if !contains(config.IgnoreList, "group") {
		group, _ := sysFile.Group()
		f.Group = group.(string)
	}
	if !contains(config.IgnoreList, "linked-to") {
		linkedTo, _ := sysFile.LinkedTo()
		f.LinkedTo = linkedTo.(string)
	}
	if !contains(config.IgnoreList, "filetype") {
		filetype, _ := sysFile.Filetype()
		f.Filetype = filetype.(string)
	}
	return f, nil
}
