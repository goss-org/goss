package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// fakeSysService is a minimal system.Service so Validate can run without
// touching the real init system.
type fakeSysService struct{}

func (f *fakeSysService) Service() string              { return "sshd" }
func (f *fakeSysService) Exists() (bool, error)        { return true, nil }
func (f *fakeSysService) Enabled() (bool, error)       { return true, nil }
func (f *fakeSysService) Running() (bool, error)       { return true, nil }
func (f *fakeSysService) RunLevels() ([]string, error) { return []string{"3"}, nil }

// service.runlevels used a plain != nil check, so `runlevels: []` asserted
// nothing and passed silently. It must now warn like file.contents does.
func TestServiceEmptyRunLevelsWarns(t *testing.T) {
	sys := &system.System{
		NewService: func(context.Context, string, *system.System, util.Config) system.Service {
			return &fakeSysService{}
		},
	}

	s := &Service{id: "sshd", RunLevels: []any{}}
	out := captureStderr(t, func() { s.Validate(sys) })

	if !strings.Contains(out, "WARNING:") || !strings.Contains(out, "service.runlevels") {
		t.Errorf("Validate with empty runlevels stderr = %q, want a WARNING naming service.runlevels", out)
	}
}
