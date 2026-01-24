//go:build linux || darwin || !windows
// +build linux darwin !windows

package system

import (
	"context"

	"github.com/goss-org/goss/util"
)

// NewServiceWindows stub for non Windows platforms.
// This is needed for compilation and should never be called since DetectService() only
// returns "windows" on Windows platforms.
func NewServiceWindows(ctx context.Context, service string, system *System, config util.Config) Service {
	return NewServiceInit(ctx, service, system, config)
}