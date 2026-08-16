package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// flagBaseline is every flag name and alias the CLI has, keyed by the command
// path it appears under. It is a literal rather than a golden file because a
// golden file invites regeneration, and the point is that this list does not
// change in this migration: no new flag, no new alias, no rename.
// Adding a flag is a deliberate act, so an intended addition means editing this
// fixture in the same commit and explaining it in the message.
var flagBaseline = map[string][]string{
	// -l is bound twice on purpose and stays that way: globally it is a
	// --log-level alias, and under serve it is --listen-addr, which wins there.
	"goss":                  {"L", "g", "gossfile", "l", "log-level", "loglevel", "package", "vars", "vars-inline"},
	"goss add":              {"exclude-attr"},
	"goss add addr":         {"timeout"},
	"goss add command":      {"timeout"},
	"goss add dns":          {"server", "timeout"},
	"goss add file":         {},
	"goss add goss":         {},
	"goss add group":        {},
	"goss add http":         {"insecure", "k", "no-follow-redirects", "p", "password", "proxy", "r", "timeout", "u", "username", "x"},
	"goss add interface":    {},
	"goss add kernel-param": {},
	"goss add mount":        {"timeout"},
	"goss add package":      {},
	"goss add port":         {},
	"goss add process":      {},
	"goss add registry":     {},
	"goss add service":      {},
	"goss add user":         {},
	"goss autoadd":          {},
	"goss render":           {"d", "debug"},
	"goss serve":            {"c", "cache", "e", "endpoint", "f", "format", "format-options", "l", "listen-addr", "max-concurrent", "o"},
	"goss validate":         {"color", "f", "format", "format-options", "max-concurrent", "no-color", "o", "r", "retry-timeout", "s", "sleep"},
}

// TestCLIAddsNoFlag pins the inventory. --log-level is not new: it already
// exists under the same name and aliases, and this change only alters what
// reading it does.
func TestCLIAddsNoFlag(t *testing.T) {
	t.Parallel()

	got := inventoryFlags(newApp(defaultOperations()))

	require.Equal(t, sortedBaseline(), got)
}

// TestFlagInventoryDetectsAnAddition proves the comparison above can fail. A
// fixture-versus-reflection test passes just as well when the reflection half is
// broken and returns nothing at all, so the walk is driven over a command tree
// with a known extra flag.
func TestFlagInventoryDetectsAnAddition(t *testing.T) {
	t.Parallel()

	app := newApp(defaultOperations())
	// Nested, because a walk that only visits the root would still pass on this.
	for _, command := range app.Commands {
		if command.Name == "serve" {
			command.Flags = append(command.Flags,
				&cli.StringFlag{Name: "log-format", Aliases: []string{"F"}})
		}
	}

	got := inventoryFlags(app)

	require.NotEqual(t, sortedBaseline(), got)
	require.Contains(t, got["goss serve"], "log-format")
	require.Contains(t, got["goss serve"], "F")
}

// inventoryFlags collects flag names and aliases per command path, over the whole
// command tree. Help and version are excluded: urfave adds them itself, and
// whether they are materialised depends on when the walk happens rather than on
// anything goss declares.
func inventoryFlags(command *cli.Command) map[string][]string {
	out := map[string][]string{}
	walkCommand(command, "", out)

	return out
}

func walkCommand(command *cli.Command, prefix string, out map[string][]string) {
	path := strings.TrimSpace(prefix + " " + command.Name)

	names := []string{}
	for _, flag := range command.Flags {
		for _, name := range flag.Names() {
			switch name {
			case "help", "h", "version", "v":
				continue
			}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	out[path] = names

	for _, sub := range command.Commands {
		walkCommand(sub, path, out)
	}
}

func sortedBaseline() map[string][]string {
	out := make(map[string][]string, len(flagBaseline))
	for path, names := range flagBaseline {
		sorted := append([]string{}, names...)
		sort.Strings(sorted)
		out[path] = sorted
	}

	return out
}
