package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// fakeSysMount is a minimal system.Mount so Validate can run without touching
// the real mount table.
type fakeSysMount struct{}

func (f *fakeSysMount) MountPoint() string          { return "/" }
func (f *fakeSysMount) Exists() (bool, error)       { return true, nil }
func (f *fakeSysMount) Opts() ([]string, error)     { return []string{"rw"}, nil }
func (f *fakeSysMount) VfsOpts() ([]string, error)  { return []string{"rw"}, nil }
func (f *fakeSysMount) Source() (string, error)     { return "/dev/sda1", nil }
func (f *fakeSysMount) Filesystem() (string, error) { return "ext4", nil }
func (f *fakeSysMount) Usage() (int, error)         { return 10, nil }

// mount.opts used a plain != nil check, so `opts: []` asserted nothing and
// passed silently. It must now warn like file.contents does.
func TestMountEmptyOptsWarns(t *testing.T) {
	sys := &system.System{
		NewMount: func(context.Context, string, *system.System, util.Config) system.Mount {
			return &fakeSysMount{}
		},
	}

	m := &Mount{id: "/", Exists: true, Opts: []any{}}
	out := captureStderr(t, func() { m.Validate(sys) })

	if !strings.Contains(out, "WARNING:") || !strings.Contains(out, "mount.opts") {
		t.Errorf("Validate with empty opts stderr = %q, want a WARNING naming mount.opts", out)
	}
}

// mount.vfs-opts had the same plain != nil check as mount.opts.
func TestMountEmptyVfsOptsWarns(t *testing.T) {
	sys := &system.System{
		NewMount: func(context.Context, string, *system.System, util.Config) system.Mount {
			return &fakeSysMount{}
		},
	}

	m := &Mount{id: "/", Exists: true, VfsOpts: []any{}}
	out := captureStderr(t, func() { m.Validate(sys) })

	if !strings.Contains(out, "WARNING:") || !strings.Contains(out, "mount.vfs-opts") {
		t.Errorf("Validate with empty vfs-opts stderr = %q, want a WARNING naming mount.vfs-opts", out)
	}
}
