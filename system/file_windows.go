//go:build windows
// +build windows

package system

import "context"

func (f *DefFile) Mode(ctx context.Context) (string, error) {
	return "-1", nil // not applicable on Windows
}

func (f *DefFile) Owner(ctx context.Context) (string, error) {
	return "-1", nil // not applicable on Windows
}

func (f *DefFile) Group(ctx context.Context) (string, error) {
	return "-1", nil // not applicable on Windows
}
