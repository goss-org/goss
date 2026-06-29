package outputs

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteNagiosResult_Table(t *testing.T) {
	tests := []struct {
		name            string
		specFile        string
		testCount       int
		failed          int
		skipped         int
		duration        time.Duration
		perfdata        bool
		verbose         bool
		summary         map[int]string
		wantContains    []string
		wantNotContains []string
		wantRet         int
	}{
		{
			name:            "ok no perf no verbose",
			specFile:        "goss.yaml",
			testCount:       3,
			failed:          0,
			skipped:         0,
			duration:        1500 * time.Millisecond,
			perfdata:        false,
			verbose:         false,
			summary:         nil,
			wantContains:    []string{"GOSS-goss.yaml OK", "Count: 3", "Failed: 0"},
			wantNotContains: []string{"|total=", "Fail "},
			wantRet:         0,
		},
		{
			name:            "critical with perf",
			specFile:        "GossFilename.yaml",
			testCount:       5,
			failed:          2,
			skipped:         1,
			duration:        2 * time.Second,
			perfdata:        true,
			verbose:         false,
			summary:         nil,
			wantContains:    []string{"GOSS-GossFilename.yaml CRITICAL", "|total=5 failed=2 skipped=1"},
			wantNotContains: []string{"Fail "},
			wantRet:         2,
		},
		{
			name:      "critical verbose",
			specFile:  "gossConfAbc.yaml",
			testCount: 2,
			failed:    1,
			skipped:   0,
			duration:  500 * time.Millisecond,
			perfdata:  false,
			verbose:   true,
			summary: map[int]string{
				0: "Fail 1 - something went wrong\n",
			},
			wantContains:    []string{"GOSS-gossConfAbc.yaml CRITICAL", "Fail 1 - something went wrong"},
			wantNotContains: []string{"|total="},
			wantRet:         2,
		},
		{
			name:      "critical perf and verbose",
			specFile:  "gossConfFile.yaml",
			testCount: 4,
			failed:    2,
			skipped:   1,
			duration:  1250 * time.Millisecond,
			perfdata:  true,
			verbose:   true,
			summary: map[int]string{
				0: "Fail 1 - a\n",
				1: "Fail 2 - b\n",
			},
			wantContains:    []string{"GOSS-gossConfFile.yaml CRITICAL", "|total=4 failed=2 skipped=1", "Fail 1 - a", "Fail 2 - b"},
			wantNotContains: []string{},
			wantRet:         2,
		},
		{
			name:            "checkNoPanic1 - missing summary text",
			specFile:        "gossConfFile.yaml",
			testCount:       4,
			failed:          2,
			skipped:         1,
			duration:        1250 * time.Millisecond,
			perfdata:        true,
			verbose:         true,
			summary:         map[int]string{},
			wantContains:    []string{"GOSS-gossConfFile.yaml CRITICAL", "|total=4 failed=2 skipped=1"},
			wantNotContains: []string{},
			wantRet:         2,
		},
		{
			name:            "checkNoPanic2 - verbose but missing",
			specFile:        "gossConfFile.yaml",
			testCount:       4,
			failed:          2,
			skipped:         1,
			duration:        1250 * time.Millisecond,
			perfdata:        false,
			verbose:         true,
			summary:         map[int]string{},
			wantContains:    []string{"GOSS-gossConfFile.yaml CRITICAL", "Count: 4", "Failed: 2", "Skipped: 1"},
			wantNotContains: []string{},
			wantRet:         2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rtn := writeNagiosResult(&buf, tt.specFile, tt.testCount, tt.failed, tt.skipped, tt.duration, tt.perfdata, tt.verbose, tt.summary)
			out := buf.String()

			assert.Equal(t, tt.wantRet, rtn, "Return code mismatch")

			for _, want := range tt.wantContains {
				assert.Contains(t, out, want)
			}
			for _, not := range tt.wantNotContains {
				assert.NotContains(t, out, not)
			}
		})
	}
}
