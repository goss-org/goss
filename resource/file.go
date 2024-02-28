package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type File struct {
	Title    string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta     meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id       string  `json:"-" yaml:"-"`
	Path     string  `json:"path,omitempty" yaml:"path,omitempty"`
	Exists   matcher `json:"exists" yaml:"exists"`
	Mode     matcher `json:"mode,omitempty" yaml:"mode,omitempty"`
	Size     matcher `json:"size,omitempty" yaml:"size,omitempty"`
	Owner    matcher `json:"owner,omitempty" yaml:"owner,omitempty"`
	Group    matcher `json:"group,omitempty" yaml:"group,omitempty"`
	LinkedTo matcher `json:"linked-to,omitempty" yaml:"linked-to,omitempty"`
	Filetype matcher `json:"filetype,omitempty" yaml:"filetype,omitempty"`
	Contains matcher `json:"contains,omitempty" yaml:"contains,omitempty"`
	Contents matcher `json:"contents" yaml:"contents"`
	Md5      matcher `json:"md5,omitempty" yaml:"md5,omitempty"`
	Sha256   matcher `json:"sha256,omitempty" yaml:"sha256,omitempty"`
	Sha512   matcher `json:"sha512,omitempty" yaml:"sha512,omitempty"`
	Skip     bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	FileResourceKey  = "file"
	FileResourceName = "File"
)

func init() {
	registerResource(FileResourceKey, &File{})
}

func (f *File) ID() string {
	if f.Path != "" && f.Path != f.id {
		return fmt.Sprintf("%s: %s", f.id, f.Path)
	}
	return f.id
}
func (f *File) SetID(id string)  { f.id = id }
func (f *File) SetSkip()         { f.Skip = true }
func (f *File) TypeKey() string  { return FileResourceKey }
func (f *File) TypeName() string { return FileResourceName }

func (f *File) GetTitle() string { return f.Title }
func (f *File) GetMeta() meta    { return f.Meta }
func (f *File) GetPath() string {
	if f.Path != "" {
		return f.Path
	}
	return f.id
}

func (f *File) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), "id", f.ID())
	skip := f.Skip
	sysFile, _ := sys.NewFile(ctx, f.GetPath(), sys, util.Config{})

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
		fmt.Fprintf(os.Stderr, "DEPRECATION WARNING: file.contains has been renamed to file.contents\n")
		results = append(results, ValidateValue(f, "contains", f.Contains, sysFile.Contents, skip))
	}
	if isSet(f.Contents) {
		results = append(results, ValidateValue(f, "contents", f.Contents, sysFile.Contents, skip))
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
	if f.Sha512 != nil {
		results = append(results, ValidateValue(f, "sha512", f.Sha512, sysFile.Sha512, skip))
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
		id:       path,
		Exists:   exists,
		Contents: []string{},
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
