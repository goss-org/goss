// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cgo,!windows,!plan9

package group

import (
	"fmt"
	"runtime"
)

func init() {
	implemented = false
}

func lookupGroup(groupname string) (*Group, error) {
	return nil, fmt.Errorf("user: LookupGroup not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func lookupGroupID(int) (*Group, error) {
	return nil, fmt.Errorf("user: LookupGroupID not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}
