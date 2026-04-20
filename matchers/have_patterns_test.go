package matchers

import (
	"bytes"
	"strings"
	"testing"
)

// TestHavePatternsMatchString verifies that HavePatterns correctly matches
// string content (the type it receives after the validate-layer materialization fix).
func TestHavePatternsMatchString(t *testing.T) {
	content := "Banner /etc/issue.net\nLogLevel INFO\n"

	m := HavePatterns([]interface{}{"Banner /etc/issue.net"})
	success, err := m.Match(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success {
		t.Fatal("expected Match to return true")
	}
}

// TestHavePatternsFailureResultString verifies that FailureResult returns the
// actual content as a string — not a Go type repr — when given string input.
// After the validate-layer fix, FailureResult always receives a materialized string.
func TestHavePatternsFailureResultString(t *testing.T) {
	content := "Banner /etc/issue.net\nLogLevel INFO\n"
	pattern := "nonexistent-pattern"

	m := HavePatterns([]interface{}{pattern})
	success, err := m.Match(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if success {
		t.Fatal("expected Match to return false for missing pattern")
	}

	result := m.FailureResult(content)

	actual, ok := result.Actual.(string)
	if !ok {
		t.Fatalf("FailureResult.Actual must be a string, got %T: %v", result.Actual, result.Actual)
	}
	if strings.Contains(actual, "object:") {
		t.Errorf("FailureResult.Actual must not contain Go type repr, got: %q", actual)
	}
	if !strings.Contains(actual, "Banner") {
		t.Errorf("FailureResult.Actual must contain the file content, got: %q", actual)
	}
}

// TestHavePatternsMatchBytesReader verifies that HavePatterns still accepts
// io.Reader directly (backwards-compat; the validate layer now sends strings).
func TestHavePatternsMatchBytesReader(t *testing.T) {
	content := "pam_faillock.so preauth\n"

	m := HavePatterns([]interface{}{"pam_faillock.so"})
	reader := bytes.NewReader([]byte(content))
	success, err := m.Match(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success {
		t.Fatal("expected Match to return true")
	}
}
