// +build integration

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	err := os.Chdir("../..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
	build := exec.Command("go", "build", "./cmd/goss")
	err = build.Run()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", gossName, err)
		os.Exit(1)
	}
	err = os.Chdir(filepath.Join("cmd", "goss"))
	if err != nil {
		fmt.Printf("could not cd back to cmd/goss to run test suite: %v", err)
	}
	os.Exit(m.Run())
}

func TestGossSuccess(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		args []string
	}{
		"no_args": {
			args: []string{},
		},
		"help": {
			args: []string{"--help"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			c := command(tc.args)
			c.Run()
			assert.True(t, c.Success())
			stdoutGolden := golden("goss", name, "stdout", true)
			stderrGolden := golden("goss", name, "stderr", true)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				ioutil.WriteFile(stdoutGolden, []byte(c.Stdout()), 0644)
				ioutil.WriteFile(stderrGolden, []byte(c.Stderr()), 0644)
			}
			assert.Equal(t, read(t, stdoutGolden), c.Stdout())
			assert.Equal(t, read(t, stderrGolden), c.Stderr())
		})
	}
}

func command(args []string) *testcli.Cmd {
	return testcli.Command(filepath.Join("..", "../", "goss"), args...)
}

func golden(command string, caseName string, streamName string, isPass bool) string {
	outcome := "fail"
	if isPass {
		outcome = "pass"
	}
	return filepath.Join("testdata", command, fmt.Sprintf("%v.%v.%v.golden", caseName, streamName, outcome))
}

func read(t *testing.T, golden string) string {
	content, err := ioutil.ReadFile(golden)
	if err != nil {
		if os.IsNotExist(err) {
			return "" // this is fine; failure will happen on assert.
		}
		t.Fatalf("Could not read golden from %v: %v", golden, err)
	}
	return string(content)
}
