package system

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSplitMountInfo(t *testing.T) {
	in := "rw,context=\"system_u:object_r:container_file_t:s0:c174,c741\",size=65536k,mode=755"
	want := []string{
		"rw",
		"context=\"system_u:object_r:container_file_t:s0:c174,c741\"",
		"size=65536k",
		"mode=755",
	}

	got := splitMountInfo(in)

	assert.DeepEqual(t, got, want)
}
