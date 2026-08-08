package matchers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSliceToPatterns_FlaggedRegex verifies that patterns of the form /regex/flags
// are recognised as regex patterns and not treated as literal strings.
// These tests reproduce the failures documented in GOSS_BUG.md.
func TestSliceToPatterns_FlaggedRegex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// patterns with /i flag must be compiled as regex, not string-contains
		{
			name:    "case-insensitive flag recognised as regex",
			pattern: "/loglevel (verbose|info)/i",
			wantErr: false,
		},
		{
			name:    "negated case-insensitive flag recognised as regex",
			pattern: "!/loglevel debug/i",
			wantErr: false,
		},
		{
			name:    "cipher pattern with /i flag recognised as regex",
			pattern: "/ciphers.*aes256-gcm@openssh\\.com/i",
			wantErr: false,
		},
		{
			name:    "macs pattern with /i flag recognised as regex",
			pattern: "/macs.*hmac-sha2-512/i",
			wantErr: false,
		},
		{
			name:    "kex pattern with /i flag recognised as regex",
			pattern: "/kexalgorithms.*ecdh-sha2-nistp521/i",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pats, err := sliceToPatterns([]string{tt.pattern})
			require.NoError(t, err)
			require.Len(t, pats, 1)
			// Must be a regexPattern, not a stringPattern.
			// A stringPattern would do strings.Contains, which won't honour
			// the alternation group or case-insensitive flag.
			_, isRegex := pats[0].(*regexPattern)
			assert.True(t, isRegex, "pattern %q must be parsed as a regex, got %T", tt.pattern, pats[0])
		})
	}
}

// TestNewRegexPattern_CaseInsensitiveFlag verifies that the /i flag is translated
// into a Go inline (?i) flag and that matching is actually case-insensitive.
func TestNewRegexPattern_CaseInsensitiveFlag(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		// --- cases from GOSS_BUG.md failure 3 (5.1.14) ---
		// sshd -T outputs "loglevel INFO" (key lowercase, value uppercase)
		// /loglevel (verbose|info)/i must match despite uppercase INFO
		{
			name:    "5.1.14: /i flag matches uppercase value INFO",
			pattern: "/loglevel (verbose|info)/i",
			input:   "loglevel INFO",
			want:    true,
		},
		{
			name:    "5.1.14: /i flag matches lowercase value info",
			pattern: "/loglevel (verbose|info)/i",
			input:   "loglevel info",
			want:    true,
		},
		{
			name:    "5.1.14: /i flag matches VERBOSE",
			pattern: "/loglevel (verbose|info)/i",
			input:   "loglevel VERBOSE",
			want:    true,
		},
		{
			name:    "5.1.14: negative pattern !/loglevel debug/i does not match INFO line",
			pattern: "!/loglevel debug/i",
			input:   "loglevel INFO",
			want:    false, // inverse=true, underlying regex does NOT match → treated as "not found" → correct
		},

		// --- cases from GOSS_BUG.md failure 1 (5.1.6) ---
		// sshd -T outputs "ciphers aes256-gcm@openssh.com,..."
		{
			name:    "5.1.6: cipher pattern with /i matches lowercase output",
			pattern: "/ciphers.*aes256-gcm@openssh\\.com/i",
			input:   "ciphers aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr",
			want:    true,
		},
		{
			name:    "5.1.6: cipher pattern with /i matches uppercase output",
			pattern: "/ciphers.*aes256-gcm@openssh\\.com/i",
			input:   "CIPHERS AES256-GCM@OPENSSH.COM,AES128-GCM@OPENSSH.COM",
			want:    true,
		},

		// --- cases from GOSS_BUG.md failure 4 (5.1.15) ---
		// sshd -T outputs "macs hmac-sha2-512,hmac-sha2-256"
		{
			name:    "5.1.15: macs pattern with /i matches output",
			pattern: "/macs.*hmac-sha2-512/i",
			input:   "macs hmac-sha2-512,hmac-sha2-256",
			want:    true,
		},
		{
			name:    "5.1.15: macs pattern with /i matches uppercase output",
			pattern: "/macs.*hmac-sha2-512/i",
			input:   "MACS HMAC-SHA2-512,HMAC-SHA2-256",
			want:    true,
		},

		// --- cases from GOSS_BUG.md failure 2 (5.1.12) ---
		{
			name:    "5.1.12: kex pattern with /i matches output",
			pattern: "/kexalgorithms.*ecdh-sha2-nistp521/i",
			input:   "kexalgorithms ecdh-sha2-nistp521,ecdh-sha2-nistp384",
			want:    true,
		},

		// --- ensure plain /pattern/ (no flag) still works ---
		{
			name:    "plain regex without flag still matches",
			pattern: "/^moo.*w$/",
			input:   "moo cow",
			want:    true,
		},
		{
			name:    "plain regex without flag does not match wrong case",
			pattern: "/loglevel info/",
			input:   "loglevel INFO",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pat, err := newRegexPattern(tt.pattern)
			require.NoError(t, err, "newRegexPattern(%q) should not return error", tt.pattern)
			got := pat.Match(tt.input)
			assert.Equal(t, tt.want, got,
				"pattern %q .Match(%q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
		})
	}
}

// TestHavePatternsMatcher_CaseInsensitiveFlag is an end-to-end test through the
// full HavePatternsMatcher, reproducing the exact failure scenarios in GOSS_BUG.md.
func TestHavePatternsMatcher_CaseInsensitiveFlag(t *testing.T) {
	tests := []struct {
		name     string
		actual   string  // simulated command stdout
		patterns []interface{}
		wantOK   bool
	}{
		// 5.1.14: loglevel INFO must match /loglevel (verbose|info)/i
		{
			name:     "5.1.14 loglevel INFO matches /i pattern",
			actual:   "loglevel INFO\n",
			patterns: []interface{}{"/loglevel (verbose|info)/i"},
			wantOK:   true,
		},
		// 5.1.14: both positive and negative patterns
		{
			name:   "5.1.14 loglevel INFO full pattern set",
			actual: "loglevel INFO\n",
			patterns: []interface{}{
				"/loglevel (verbose|info)/i",
				"!/loglevel debug/i",
			},
			wantOK: true,
		},
		// 5.1.6: cipher line must match all strong cipher patterns
		{
			name:   "5.1.6 cipher line matches all strong cipher patterns",
			actual: "ciphers aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr\n",
			patterns: []interface{}{
				"/ciphers.*aes256-gcm@openssh\\.com/i",
				"/ciphers.*aes128-gcm@openssh\\.com/i",
				"/ciphers.*aes256-ctr/i",
				"/ciphers.*aes192-ctr/i",
				"/ciphers.*aes128-ctr/i",
			},
			wantOK: true,
		},
		// 5.1.15: macs line must match strong mac patterns
		{
			name:   "5.1.15 macs line matches strong mac patterns",
			actual: "macs hmac-sha2-512,hmac-sha2-256\n",
			patterns: []interface{}{
				"/macs.*hmac-sha2-512/i",
				"/macs.*hmac-sha2-256/i",
			},
			wantOK: true,
		},
		// 5.1.12: kex line must match strong kex patterns
		{
			name:   "5.1.12 kex line matches strong kex patterns",
			actual: "kexalgorithms ecdh-sha2-nistp521,ecdh-sha2-nistp384,ecdh-sha2-nistp256,diffie-hellman-group16-sha512,diffie-hellman-group-exchange-sha256\n",
			patterns: []interface{}{
				"/kexalgorithms.*ecdh-sha2-nistp521/i",
				"/kexalgorithms.*ecdh-sha2-nistp384/i",
				"/kexalgorithms.*ecdh-sha2-nistp256/i",
				"/kexalgorithms.*diffie-hellman-group16-sha512/i",
				"/kexalgorithms.*diffie-hellman-group-exchange-sha256/i",
			},
			wantOK: true,
		},
		// sanity: /i flag must NOT match when truly absent
		{
			name:     "5.1.14 loglevel DEBUG fails positive pattern",
			actual:   "loglevel DEBUG\n",
			patterns: []interface{}{"/loglevel (verbose|info)/i"},
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := HavePatterns(tt.patterns)
			ok, err := m.Match(strings.NewReader(tt.actual))
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok,
				"HavePatterns(%v).Match(%q) = %v, want %v", tt.patterns, tt.actual, ok, tt.wantOK)
		})
	}
}
