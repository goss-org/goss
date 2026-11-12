package goss

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

func checkErr(t *testing.T, err error, format string, a ...any) {
	t.Helper()
	if err == nil {
		return
	}

	t.Fatalf(format+": "+err.Error(), a...)
}

func TestConfigMerge(t *testing.T) {
	var g1json = `file:
  /etc/passwd:
    exists: true
    mode: "0644"
    size: 1722
    owner: root
    group: root
    filetype: file
    contains: []`

	var g2json = `service:
  sshd:
    enabled: true
    running: true
`

	g1, err := ReadJSONData([]byte(g1json), true)
	checkErr(t, err, "reading g1 failed")
	_, ok := g1.Services["sshd"]
	if ok {
		t.Fatalf("did not expect sshd service")
	}

	g2, err := ReadJSONData([]byte(g2json), true)
	checkErr(t, err, "reading g1 failed")

	g1.Merge(g2)
	_, ok = g1.Files["/etc/passwd"]
	if !ok {
		t.Fatalf("expected passwd file, got none")
	}
	_, ok = g1.Services["sshd"]
	if !ok {
		t.Fatalf("expected sshd service, got none")
	}
}

func TestUseAsPackage(t *testing.T) {
	ctx := t.Context()
	output := &bytes.Buffer{}

	// temp spec file
	fh, err := os.CreateTemp("", "*.yaml")
	checkErr(t, err, "temp file failed")
	fh.Close()

	// new config that doesnt spam output etc
	cfg, err := util.NewConfig(util.WithFormatOptions("pretty"), util.WithResultWriter(output), util.WithSpecFile(fh.Name()))
	checkErr(t, err, "new config failed")

	// adds the os tmp dir to the goss spec file
	err = AddResources(ctx, fh.Name(), "File", []string{os.TempDir()}, cfg)
	checkErr(t, err, "could not add resource %q", os.TempDir())

	// validate and sanity check, compare structured vs direct results etc
	results, err := ValidateResults(ctx, cfg)
	checkErr(t, err, "check failed")

	found := 0
	passed := 0
	for rg := range results {
		for _, r := range rg {
			found++

			if r.Result == resource.SUCCESS {
				passed++
			}
		}
	}

	code, err := Validate(ctx, cfg)
	checkErr(t, err, "check failed")
	if code != 0 {
		t.Fatalf("check failed, expected 0 got %d", code)
	}

	res := &outputs.StructuredOutput{}
	err = json.Unmarshal(output.Bytes(), res)
	checkErr(t, err, "unmarshal failed")

	if res.Summary.Failed != 0 {
		t.Fatalf("expected 0 failed, got %d", res.Summary.Failed)
	}

	if len(res.Results) != found {
		t.Fatalf("expected %d results for %d", found, len(res.Results))
	}

	okcount := 0
	for _, r := range res.Results {
		if r.Result == resource.SUCCESS {
			okcount++
		}
	}

	if okcount != passed {
		t.Fatalf("expected %d passed but got %d", passed, okcount)
	}
}

func TestSkipResourcesByType(t *testing.T) {
	ctx := t.Context()
	output := &bytes.Buffer{}

	// temp spec file
	fh, err := os.CreateTemp("", "*.yaml")
	checkErr(t, err, "temp file failed")
	fh.Close()

	// new config that doesnt spam output etc
	cfg, err := util.NewConfig(util.WithFormatOptions("pretty"), util.WithResultWriter(output), util.WithSpecFile(fh.Name()), util.WithDisabledResourceTypes("file"))
	checkErr(t, err, "new config failed")

	// adds the os tmp dir to the goss spec file
	err = AddResources(ctx, fh.Name(), "File", []string{os.TempDir()}, cfg)
	checkErr(t, err, "could not add resource %q", os.TempDir())

	// validate and sanity check, compare structured vs direct results etc
	results, err := ValidateResults(ctx, cfg)
	checkErr(t, err, "check failed")

	skipped := 0
	for rg := range results {
		for _, r := range rg {
			if r.Skipped {
				skipped++
			}
		}
	}

	if skipped != 5 {
		t.Fatalf("Expected to skip 5 tests, skipped %d", skipped)
	}
}

// A max-concurrent below 1 used to start no workers, so no test ran and
// validation passed with "Count: 0".
func TestValidateRunsTestsWithNonPositiveMaxConcurrent(t *testing.T) {
	for _, mc := range []int{0, -1} {
		cfg, err := util.NewConfig(util.WithMaxConcurrency(mc))
		checkErr(t, err, "new config failed")

		specFile := filepath.Join(t.TempDir(), "goss.yaml")
		missing := filepath.Join(t.TempDir(), "does-not-exist")
		spec := "file:\n  " + missing + ":\n    exists: true\n"
		checkErr(t, os.WriteFile(specFile, []byte(spec), 0o644), "write spec failed")
		cfg.Spec = specFile

		results, err := ValidateResults(t.Context(), cfg)
		checkErr(t, err, "check failed")

		count, failed := 0, 0
		for rg := range results {
			for _, r := range rg {
				count++
				if r.Result == resource.FAIL {
					failed++
				}
			}
		}
		if count != 1 || failed != 1 {
			t.Errorf("max-concurrent %d: ran %d tests with %d failed, want 1 and 1", mc, count, failed)
		}
	}
}
