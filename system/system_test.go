package system

import (
	"reflect"
	"runtime"
	"testing"
)

type noInputs func() string

// test that a function with no inputs returns one of the expected strings
func testOutputs(f noInputs, validOutputs []string, t *testing.T) {
	output := f()
	// use reflect to get the name of the function
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	failed := true
	for _, valid := range validOutputs {
		if output == valid {
			failed = false
		}
	}
	if failed {
		t.Errorf("Function %v returned %v, which is not one of %v", name, output, validOutputs)
	}
}

func TestPackageManager(t *testing.T) {
	t.Parallel()
	testOutputs(DetectPackageManager, []string{"deb", "rpm", "apk", "pacman"}, t)
}

func TestDetectService(t *testing.T) {
	t.Parallel()
	testOutputs(DetectService, []string{"systemd", "init", "alpineinit", "upstart"}, t)
}

func TestDetectDistro(t *testing.T) {
	t.Parallel()
	testOutputs(
		DetectDistro,
		[]string{"ubuntu", "redhat", "alpine", "arch", "debian"},
		t,
	)
}

func TestHasCommand(t *testing.T) {
	t.Parallel()
	if !HasCommand("sh") {
		t.Error("System didn't have sh!")
	}
}
