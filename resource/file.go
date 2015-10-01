package resource

import "github.com/aelsabbahy/goss/system"

type File struct {
	Path     string   `json:"path"`
	Exists   bool     `json:"exists"`
	Mode     string   `json:"mode,omitempty"`
	Owner    string   `json:"owner,omitempty"`
	Group    string   `json:"group,omitempty"`
	LinkedTo string   `json:"linked-to,omitempty"`
	Filetype string   `json:"filetype,omitempty"`
	Contains []string `json:"contains"`
}

func (f *File) Validate(sys *system.System) []TestResult {
	sysFile := sys.NewFile(f.Path, sys)

	var results []TestResult

	results = append(results, ValidateValue(f.Path, "exists", f.Exists, sysFile.Exists))
	if !f.Exists {
		return results
	}

	if f.Mode != "" {
		results = append(results, ValidateValue(f.Path, "mode", f.Mode, sysFile.Mode))
	}

	if f.Owner != "" {
		results = append(results, ValidateValue(f.Path, "owner", f.Owner, sysFile.Owner))
	}

	if f.Group != "" {
		results = append(results, ValidateValue(f.Path, "group", f.Group, sysFile.Group))
	}

	if f.LinkedTo != "" {
		results = append(results, ValidateValue(f.Path, "linkedto", f.LinkedTo, sysFile.LinkedTo))
	}

	if f.Filetype != "" {
		results = append(results, ValidateValue(f.Path, "filetype", f.Filetype, sysFile.Filetype))
	}

	if len(f.Contains) != 0 {
		results = append(results, ValidateContains(f.Path, "contains", f.Contains, sysFile.Contains))
	}

	return results
}

func NewFile(sysFile system.File) *File {
	path := sysFile.Path()
	mode, _ := sysFile.Mode()
	owner, _ := sysFile.Owner()
	group, _ := sysFile.Group()
	linkedTo, _ := sysFile.LinkedTo()
	filetype, _ := sysFile.Filetype()
	exists, _ := sysFile.Exists()
	return &File{
		Path:     path,
		Mode:     mode.(string),
		Owner:    owner.(string),
		Group:    group.(string),
		LinkedTo: linkedTo.(string),
		Filetype: filetype.(string),
		Contains: []string{},
		Exists:   exists.(bool),
	}
}
