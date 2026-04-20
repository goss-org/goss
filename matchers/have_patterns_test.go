package matchers

import (
	"bytes"
	"strings"
	"testing"
)

// TestHavePatternsFailureResultActualIsString asserts that FailureResult returns
// the file content as a string, not the Go type name "object: *bytes.Reader".
func TestHavePatternsFailureResultActualIsString(t *testing.T) {
	content := "Banner /etc/issue.net\nLogLevel INFO\n"
	pattern := "nonexistent-pattern"

	m := HavePatterns([]interface{}{pattern})

	reader := strings.NewReader(content)
	success, err := m.Match(reader)
	if err != nil {
		t.Fatalf("Match returned unexpected error: %v", err)
	}
	if success {
		t.Fatal("expected Match to return false for missing pattern")
	}

	// reader is now consumed; FailureResult must still show the content
	result := m.FailureResult(reader)

	actual, ok := result.Actual.(string)
	if !ok {
		t.Fatalf("FailureResult.Actual must be a string, got %T: %v", result.Actual, result.Actual)
	}
	if strings.Contains(actual, "object:") {
		t.Errorf("FailureResult.Actual must not contain Go type repr, got: %q", actual)
	}
}

func TestHavePatternsFailureResultBytesReader(t *testing.T) {
	content := "pam_faillock.so preauth\n"
	pattern := "nonexistent-pattern"

	m := HavePatterns([]interface{}{pattern})

	reader := bytes.NewReader([]byte(content))
	success, err := m.Match(reader)
	if err != nil {
		t.Fatalf("Match returned unexpected error: %v", err)
	}
	if success {
		t.Fatal("expected Match to return false for missing pattern")
	}

	result := m.FailureResult(reader)

	actual, ok := result.Actual.(string)
	if !ok {
		t.Fatalf("FailureResult.Actual must be a string, got %T: %v", result.Actual, result.Actual)
	}
	if strings.Contains(actual, "object:") {
		t.Errorf("FailureResult.Actual must not contain Go type repr, got: %q", actual)
	}
}
