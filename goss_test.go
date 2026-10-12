package goss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestDiscoveryRetryEmitsOnlyFinalAttempt(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "first-attempt")
	command := []string{"sh", "-c", fmt.Sprintf("if [ -f %q ]; then exit 0; fi; touch %q; exit 1", marker, marker)}
	var output bytes.Buffer
	config, gossConfig := discoveryRetryConfig(command, &output, time.Second, 0)
	stderr, code, err := captureDiscoveryStderr(t, func() (int, error) {
		return ValidateConfig(t.Context(), config, gossConfig)
	})

	if err != nil {
		t.Fatalf("retry returned an error: %v", err)
	}
	if code != 0 {
		t.Fatalf("retry returned status %d, want 0", code)
	}
	if stderr == "" {
		t.Fatal("retry progress was not emitted separately from stdout")
	}
	if !bytes.HasSuffix(output.Bytes(), []byte("\n")) || bytes.Count(output.Bytes(), []byte("\n")) != 1 {
		t.Fatalf("retry output is not one newline-terminated document: %q", output.String())
	}
	var document map[string]any
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("final output is not one JSON document: %v", err)
	}
	for _, value := range document {
		variable := value.(map[string]any)
		if variable["outcome"] != resource.OutcomePass {
			t.Fatalf("final document outcome = %v, want pass", variable["outcome"])
		}
	}
	varsFile := filepath.Join(t.TempDir(), "retry-discovery.json")
	if err := os.WriteFile(varsFile, output.Bytes(), 0o600); err != nil {
		t.Fatalf("writing retry document: %v", err)
	}
	if _, err := loadVars([]string{varsFile}, ""); err != nil {
		t.Fatalf("retry document cannot be loaded by --vars: %v", err)
	}
}

func TestDiscoveryFormatDocumentation(t *testing.T) {
	for _, path := range []string{"README.md", "docs/cli.md"} {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading format documentation: %v", err)
			}
			for _, detail := range []string{"discovery", "--vars", "meta.register", "pretty", "sort", "exclude_raw"} {
				if !strings.Contains(string(content), detail) {
					t.Errorf("%s does not describe %s", path, detail)
				}
			}
		})
	}
}

func TestDiscoveryRetryTimeoutPreservesFinalDocumentAndStatus(t *testing.T) {
	var output bytes.Buffer
	config, gossConfig := discoveryRetryConfig([]string{"sh", "-c", "exit 1"}, &output, time.Second, 2*time.Second)
	stderr, code, err := captureDiscoveryStderr(t, func() (int, error) {
		return ValidateConfig(t.Context(), config, gossConfig)
	})

	if err != nil {
		t.Fatalf("discovery timeout returned an error that the CLI would print to stdout: %v", err)
	}
	if code != 1 {
		t.Fatalf("timeout returned status %d, want final discovery status 1", code)
	}
	if stderr == "" {
		t.Fatal("stderr progress did not explain the timeout")
	}
	if !bytes.HasSuffix(output.Bytes(), []byte("\n")) || bytes.Count(output.Bytes(), []byte("\n")) != 1 {
		t.Fatalf("timeout output is not one newline-terminated document: %q", output.String())
	}
	var document map[string]any
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("timeout output is not one JSON document: %v", err)
	}
}

func TestDiscoveryRetryDoesNotRepeatContractFailures(t *testing.T) {
	data := []byte(`{"command":{"retry":{"exec":["sh","-c","exit 0"],"exit-status":0,"meta":{"register":7}}}}`)
	gossConfig, parseErr := ReadJSONData(data, true)
	if parseErr != nil {
		t.Fatalf("building invalid-register config: %v", parseErr)
	}
	var output bytes.Buffer
	config := &util.Config{OutputFormat: "discovery", OutputWriter: &output, RetryTimeout: time.Second, MaxConcurrent: 1}
	_, code, err := captureDiscoveryStderr(t, func() (int, error) {
		return ValidateConfig(t.Context(), config, &gossConfig)
	})

	if err != nil {
		t.Fatalf("contract failure returned unexpected error: %v", err)
	}
	if code != 2 {
		t.Fatalf("contract failure returned status %d, want 2", code)
	}
	if output.Len() != 0 {
		t.Fatalf("contract failure wrote stdout: %s", output.String())
	}
}

type discoveryRetryShortWriter struct{}

func (discoveryRetryShortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return 1, nil
}

type discoveryRetryFailWriter struct{}

func (discoveryRetryFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("final destination unavailable")
}

func TestDiscoveryRetryReportsFinalOutputFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		writer io.Writer
	}{
		{name: "short write", writer: discoveryRetryShortWriter{}},
		{name: "explicit error", writer: discoveryRetryFailWriter{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, gossConfig := discoveryRetryConfig([]string{"sh", "-c", "exit 0"}, test.writer, time.Second, 0)
			stderr, code, err := captureDiscoveryStderr(t, func() (int, error) {
				return ValidateConfig(t.Context(), config, gossConfig)
			})

			if err != nil {
				t.Fatalf("output failure returned unexpected error: %v", err)
			}
			if code != 2 {
				t.Fatalf("output failure returned status %d, want 2", code)
			}
			if stderr == "" {
				t.Fatal("output failure did not produce a diagnostic")
			}
		})
	}
}

func discoveryRetryConfig(command []string, writer io.Writer, timeout, sleep time.Duration) (*util.Config, *GossConfig) {
	gossConfig := NewGossConfig()
	commandResource := &resource.Command{
		Exec:       &util.ExecCommand{CmdSlice: command},
		ExitStatus: 0,
	}
	commandResource.SetID("retry")
	gossConfig.Commands["retry"] = commandResource
	config := &util.Config{
		OutputFormat:  "discovery",
		OutputWriter:  writer,
		RetryTimeout:  timeout,
		Sleep:         sleep,
		MaxConcurrent: 1,
	}
	return config, gossConfig
}

func captureDiscoveryStderr(t *testing.T, run func() (int, error)) (string, int, error) {
	t.Helper()
	originalStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stderr capture: %v", err)
	}
	os.Stderr = writer
	defer func() { os.Stderr = originalStderr }()

	code, runErr := run()
	if err := writer.Close(); err != nil {
		t.Fatalf("closing stderr writer: %v", err)
	}
	os.Stderr = originalStderr
	diagnostic, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading stderr: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("closing stderr reader: %v", err)
	}
	return string(diagnostic), code, runErr
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
