package goss

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoggingDocumentationCoversTheContract checks that the page says
// something. A site build proves the markdown parses, not that it has content,
// so the six subjects the page owes a reader are checked individually. Each is
// matched on its heading, because a passing mention of "migration" somewhere in
// prose is not a migration section.
func TestLoggingDocumentationCoversTheContract(t *testing.T) {
	t.Parallel()

	body := readDoc(t, "docs/logging.md")

	// The subject, and the heading that carries it. Headings are
	// matched leniently on a substring so that wording can change without
	// breaking the gate, but the level is pinned: these are sections, not
	// asides.
	sections := []struct {
		subject string
		heading string
	}{
		{"levels", "## Levels"},
		{"record schema", "## Record schema"},
		{"embedding", "## Embedding"},
		{"TRACE hook composition", "## Rendering TRACE"},
		{"sensitive fields", "## Sensitive fields"},
		{"migration", "## Migrating"},
	}

	for _, section := range sections {
		require.Contains(t, body, section.heading,
			"the %s section is missing from docs/logging.md", section.subject)
	}

	// The sections have to be about their subject, so each one is required to
	// name the API or attributes it exists to describe. Without this the
	// headings alone would pass on an empty document.
	contents := map[string][]string{
		"levels":                 {"TRACE", "util.LevelTrace", "GOSS_LOGLEVEL"},
		"record schema":          {"snake_case", "command output", "validation result", "request complete", "response_body"},
		"embedding":              {"util.WithLogger", "system.WithLogger", "OutputConfig", "slog.Default"},
		"TRACE hook composition": {"util.ReplaceTraceLevel", "ReplaceAttr", "util.Trace"},
		"sensitive fields":       {"output", "expected", "actual", "response_body", "resource_id"},
		"migration":              {"util.Config.LogLevel", "system.New", "GossConfig.Merge", "render"},
	}

	for subject, required := range contents {
		for _, needle := range required {
			require.Contains(t, body, needle,
				"the %s section should mention %q", subject, needle)
		}
	}
}

// TestLoggingDocumentationIsReachable is the navigation half. The site nav is
// generated from the docs tree, so the file being present is what puts it in
// the menu; what is not automatic is a reader of the CLI or migration pages
// ever finding out the page exists.
func TestLoggingDocumentationIsReachable(t *testing.T) {
	t.Parallel()

	require.FileExists(t, "docs/logging.md", "the page has to exist to be navigable")

	cli := readDoc(t, "docs/cli.md")
	require.Contains(t, cli, "logging.md",
		"the --loglevel documentation should link to the logging page")

	migrations := readDoc(t, "docs/migrations.md")
	require.Contains(t, migrations, "logging.md",
		"the migration guide should link to the logging page")
	require.Contains(t, migrations, "log/slog",
		"the migration guide should describe the logging change itself")
}

// TestCLIDocumentationDoesNotPromiseAPostKeywordFlag guards a claim that costs
// readers time: the documented "may be given either before or after the command
// name" is false. --log-level is a parse error after the command name, and -l
// there binds --listen-addr under serve. Changing the parser is a separate
// question, so the documentation must not assert it.
func TestCLIDocumentationDoesNotPromiseAPostKeywordFlag(t *testing.T) {
	t.Parallel()

	cli := readDoc(t, "docs/cli.md")
	require.NotContains(t, cli, "either\n    before or after the command name")
	require.NotContains(t, cli, "before or after the command name")
}

func readDoc(t *testing.T, path string) string {
	t.Helper()

	body, err := os.ReadFile(path)
	require.NoError(t, err)

	return strings.ReplaceAll(string(body), "\r\n", "\n")
}
