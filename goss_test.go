package goss

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/util"
)

func checkErr(t *testing.T, err error, format string, a ...interface{}) {
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
	output := &bytes.Buffer{}

	// temp spec file
	fh, err := ioutil.TempFile("", "*.yaml")
	checkErr(t, err, "temp file failed")
	fh.Close()

	// new config that doesnt spam output etc
	cfg, err := util.NewConfig(util.WithFormatOptions("pretty"), util.WithResultWriter(output), util.WithSpecFile(fh.Name()))
	checkErr(t, err, "new config failed")

	// adds the os tmp dir to the goss spec file
	err = AddResources(fh.Name(), "File", []string{os.TempDir()}, cfg)
	checkErr(t, err, "could not add resource %q", os.TempDir())

	// validate and sanity check, compare structured vs direct results etc
	results, err := ValidateResults(cfg)
	checkErr(t, err, "check failed")

	found := 0
	passed := 0
	for rg := range results {
		for _, r := range rg {
			found++

			if r.Successful {
				passed++
			}
		}
	}

	code, err := Validate(cfg, time.Now())
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
		if r.Successful {
			okcount++
		}
	}

	if okcount != passed {
		t.Fatalf("expected %d passed but got %d", passed, okcount)
	}
}
