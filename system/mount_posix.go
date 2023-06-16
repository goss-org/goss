//go:build linux || darwin || !windows
// +build linux darwin !windows

package system

import (
	"math"
	"syscall"
)

func getUsage(mountpoint string) (int, error) {
	statfsOut := &syscall.Statfs_t{}
	err := syscall.Statfs(mountpoint, statfsOut)
	if err != nil {
		return -1, err
	}

	percentageFree := float64(statfsOut.Bfree) / float64(statfsOut.Blocks)
	usage := math.Round((1 - percentageFree) * 100)

	return int(usage), nil
}
