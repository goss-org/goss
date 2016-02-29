package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type File struct {
	Title    string   `json:"title,omitempty" yaml:"title,omitempty"`
	Meta     meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path     string   `json:"-" yaml:"-"`
	Exists   bool     `json:"exists" yaml:"exists"`
	Mode     matcher  `json:"mode,omitempty" yaml:"mode,omitempty"`
	Owner    matcher  `json:"owner,omitempty" yaml:"owner,omitempty"`
	Group    matcher  `json:"group,omitempty" yaml:"group,omitempty"`
	LinkedTo matcher  `json:"linked-to,omitempty" yaml:"linked-to,omitempty"`
	Filetype matcher  `json:"filetype,omitempty" yaml:"filetype,omitempty"`
	Contains []string `json:"contains" yaml:"contains"`
}

func (f *File) ID() string      { return f.Path }
func (f *File) SetID(id string) { f.Path = id }

func (f *File) GetTitle() string { return f.Title }
func (f *File) GetMeta() meta    { return f.Meta }

func (f *File) Validate(sys *system.System) []TestResult {
	sysFile := sys.NewFile(f.Path, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(f, "exists", f.Exists, sysFile.Exists))

	if f.Mode != nil {
		results = append(results, ValidateValue(f, "mode", f.Mode, sysFile.Mode))
	}

	if f.Owner != nil {
		results = append(results, ValidateValue(f, "owner", f.Owner, sysFile.Owner))
	}

	if f.Group != nil {
		results = append(results, ValidateValue(f, "group", f.Group, sysFile.Group))
	}

	if f.LinkedTo != nil {
		results = append(results, ValidateValue(f, "linkedto", f.LinkedTo, sysFile.LinkedTo))
	}

	if f.Filetype != nil {
		results = append(results, ValidateValue(f, "filetype", f.Filetype, sysFile.Filetype))
	}

	if len(f.Contains) > 0 {
		results = append(results, ValidateContains(f, "contains", f.Contains, sysFile.Contains))
	}

	return results
}

func NewFile(sysFile system.File, config util.Config) (*File, error) {
	path := sysFile.Path()
	exists, _ := sysFile.Exists()
	f := &File{
		Path:     path,
		Exists:   exists,
		Contains: []string{},
	}
	if !contains(config.IgnoreList, "mode") {
		if mode, err := sysFile.Mode(); err == nil {
			f.Mode = mode
		}
	}
	if !contains(config.IgnoreList, "owner") {
		if owner, err := sysFile.Owner(); err == nil {
			f.Owner = owner
		}
	}
	if !contains(config.IgnoreList, "group") {
		if group, err := sysFile.Group(); err == nil {
			f.Group = group
		}
	}
	if !contains(config.IgnoreList, "linked-to") {
		if linkedTo, err := sysFile.LinkedTo(); err == nil {
			f.LinkedTo = linkedTo
		}
	}
	if !contains(config.IgnoreList, "filetype") {
		if filetype, err := sysFile.Filetype(); err == nil {
			f.Filetype = filetype
		}
	}
	return f, nil
}
