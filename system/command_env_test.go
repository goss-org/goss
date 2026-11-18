package system

import (
	"reflect"
	"testing"
)

func TestCommandEnv(t *testing.T) {
	t.Parallel()

	base := []string{"PATH=/bin", "HOME=/root"}
	got := commandEnv(base, map[string]string{"ZED": "z", "HOME": "/tmp", "ALPHA": "a"})
	want := []string{"PATH=/bin", "HOME=/root", "ALPHA=a", "HOME=/tmp", "ZED=z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("commandEnv() = %v, want %v", got, want)
	}
	if len(base) != 2 {
		t.Errorf("base was modified: %v", base)
	}
}
