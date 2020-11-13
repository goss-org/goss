package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type File struct {
	Title    string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta     meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path     string  `json:"-" yaml:"-"`
	Exists   matcher `json:"exists" yaml:"exists"`
	Mode     matcher `json:"mode,omitempty" yaml:"mode,omitempty"`
	Size     matcher `json:"size,omitempty" yaml:"size,omitempty"`
	Owner    matcher `json:"owner,omitempty" yaml:"owner,omitempty"`
	Group    matcher `json:"group,omitempty" yaml:"group,omitempty"`
	LinkedTo matcher `json:"linked-to,omitempty" yaml:"linked-to,omitempty"`
	Filetype matcher `json:"filetype,omitempty" yaml:"filetype,omitempty"`
	Contains matcher `json:"contains" yaml:"contains"`
	Md5      matcher `json:"md5,omitempty" yaml:"md5,omitempty"`
	Sha256   matcher `json:"sha256,omitempty" yaml:"sha256,omitempty"`
	Skip     bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (f *File) ID() string      { return f.Path }
func (f *File) SetID(id string) { f.Path = id }

func (f *File) GetTitle() string { return f.Title }
func (f *File) GetMeta() meta    { return f.Meta }

func (f *File) Validate(sys *system.System) []TestResult {
	skip := false
	sysFile := sys.NewFile(f.Path, sys, util.Config{})

	if f.Skip {
		skip = true
	}

	var results []TestResult
	results = append(results, ValidateValue(f, "exists", f.Exists, sysFile.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if f.Mode != nil {
		results = append(results, ValidateValue(f, "mode", f.Mode, sysFile.Mode, skip))
	}
	if f.Owner != nil {
		results = append(results, ValidateValue(f, "owner", f.Owner, sysFile.Owner, skip))
	}
	if f.Group != nil {
		results = append(results, ValidateValue(f, "group", f.Group, sysFile.Group, skip))
	}
	if f.LinkedTo != nil {
		results = append(results, ValidateValue(f, "linkedto", f.LinkedTo, sysFile.LinkedTo, skip))
	}
	if f.Filetype != nil {
		results = append(results, ValidateValue(f, "filetype", f.Filetype, sysFile.Filetype, skip))
	}
	if isSet(f.Contains) {
		results = append(results, ValidateValue(f, "contains", f.Contains, sysFile.Contains, skip))
	}
	if f.Size != nil {
		results = append(results, ValidateValue(f, "size", f.Size, sysFile.Size, skip))
	}
	if f.Md5 != nil {
		results = append(results, ValidateValue(f, "md5", f.Md5, sysFile.Md5, skip))
	}
	if f.Sha256 != nil {
		results = append(results, ValidateValue(f, "sha256", f.Sha256, sysFile.Sha256, skip))
	}
	return results
}

func NewFile(sysFile system.File, config util.Config) (*File, error) {
	path := sysFile.Path()
	exists, err := sysFile.Exists()
	if err != nil {
		return nil, err
	}
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
