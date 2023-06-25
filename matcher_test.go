package goss

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestMatchers(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "out_matching_*"))
	if err != nil {
		t.Fatal(err)
	}

	for _, outFile := range files {
		outFile := outFile
		parts := strings.Split(outFile, ".")
		tfName := fmt.Sprintf("%s.yaml", strings.TrimPrefix(parts[0], "testdata/out_"))
		tf := filepath.Join("testdata", tfName)
		outFormat := parts[2]
		wantCode, err := strconv.Atoi(parts[1])
		if err != nil {
			t.Fatal(err)
		}
		tn := outFile
		t.Run(tn, func(t *testing.T) {
			output := &bytes.Buffer{}

			cfg, err := util.NewConfig(
				util.WithOutputFormat(outFormat),
				util.WithResultWriter(output),
				util.WithSpecFile(tf),
				util.WithFormatOptions("sort", "pretty"),
			)
			if err != nil {
				t.Fatal(err)
			}
			exitCode, err := Validate(cfg)
			actualOut := output.String()
			actualOut = sanitizeOutput(actualOut)

			if *update {
				os.WriteFile(outFile, []byte(actualOut), 0644)
			}
			wantOutB, err := os.ReadFile(outFile)
			if err != nil {
				t.Fatal(err)
			}
			wantOut := string(wantOutB)
			if actualOut != wantOut {
				assert.Equal(t, wantOut, actualOut)
			}
			if exitCode != wantCode {
				assert.Equal(t, wantCode, exitCode)
			}
		})
	}
}

func sanitizeOutput(s string) string {
	// Remove duration time
	re := regexp.MustCompile(`\d\.\d\d\ds`)
	return re.ReplaceAllString(s, "")
}
