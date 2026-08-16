package goss

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The tests in this file are the ones a reviewer cannot do by eye across a whole
// module. They are deliberately static: a behavioural test proves something
// about the paths it drives, while "no library code writes to a process logger"
// is a statement about all of it.

// shippedFile is one parsed non-test Go file of the module.
type shippedFile struct {
	path string
	file *ast.File
	fset *token.FileSet
}

func shippedFiles(t *testing.T) []shippedFile {
	t.Helper()

	fset := token.NewFileSet()
	var out []shippedFile

	require.NoError(t, filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			// Dot directories are tooling, not shipped source. The walk root
			// is named "." and is not one of them.
			if path != "." && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			switch entry.Name() {
			case "vendor", "docs", "integration-tests", "extras", "development":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		out = append(out, shippedFile{path: filepath.ToSlash(path), file: parsed, fset: fset})

		return nil
	}))

	require.NotEmpty(t, out, "the walk should have found the module's source")

	return out
}

// importPath returns the import path of spec, unquoted.
func importPath(spec *ast.ImportSpec) string {
	return strings.Trim(spec.Path.Value, `"`)
}

// TestOnlyTerminalCLIPathsImportLog covers the whole module. Nothing in the
// library may reach the process-wide logger, and the file that still does holds
// the two terminal paths in package main, where no injected logger exists yet
// or can be trusted.
func TestOnlyTerminalCLIPathsImportLog(t *testing.T) {
	allowed := map[string]bool{"cmd/goss/goss.go": true}
	found := 0

	for _, shipped := range shippedFiles(t) {
		for _, spec := range shipped.file.Imports {
			if importPath(spec) != "log" {
				continue
			}
			found++
			require.True(t, allowed[shipped.path],
				"%s imports the log package; only the terminal CLI paths may", shipped.path)
		}
	}

	require.Equal(t, 1, found,
		"the two terminal diagnostics in cmd/goss are the only remaining users")
}

// TestLogutilsIsGone covers the dependency, not just its use. A filter that
// matches message prefixes cannot gate structured records, so leaving it behind
// would leave dead code claiming to do exactly that.
func TestLogutilsIsGone(t *testing.T) {
	for _, shipped := range shippedFiles(t) {
		for _, spec := range shipped.file.Imports {
			require.NotContains(t, importPath(spec), "logutils",
				"%s still imports logutils", shipped.path)
		}
	}

	for _, name := range []string{"go.mod", "go.sum"} {
		contents, err := os.ReadFile(name)
		require.NoError(t, err)
		require.NotContains(t, string(contents), "logutils", "%s still references logutils", name)
	}
}

// TestNoLibraryCodeCallsSlogDefault covers the gap sloglint's no-global setting
// leaves: it reports a package-level slog.Info but says nothing about
// slog.Default().Info. A global logger reintroduced that way would pass every
// other gate in this change.
func TestNoLibraryCodeCallsSlogDefault(t *testing.T) {
	for _, shipped := range shippedFiles(t) {
		slogNames := map[string]bool{}
		for _, spec := range shipped.file.Imports {
			if importPath(spec) != "log/slog" {
				continue
			}
			name := "slog"
			if spec.Name != nil {
				name = spec.Name.Name
			}
			slogNames[name] = true
		}
		if len(slogNames) == 0 {
			continue
		}

		ast.Inspect(shipped.file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || !slogNames[pkg.Name] {
				return true
			}
			require.NotEqual(t, "Default", sel.Sel.Name,
				"%s: goss has no default logger, by design", shipped.fset.Position(sel.Pos()))
			require.NotEqual(t, "SetDefault", sel.Sel.Name,
				"%s: goss does not touch the process-wide logger", shipped.fset.Position(sel.Pos()))

			return true
		})
	}
}

// TestNoConfiguredLevelStringIsInterpretedOutsideTheCLI keeps level parsing in
// one place. Under slog the level belongs to the handler, and a second
// interpretation of a level string in the library would put the two back into
// disagreement. That is how --debug and --log-level came to overlap in the
// first place.
func TestNoConfiguredLevelStringIsInterpretedOutsideTheCLI(t *testing.T) {
	// util/logging.go renders the name of goss's own level, which interprets no
	// configuration, and is why this check allows exactly that one.
	allowed := map[string]bool{"util/logging.go": true}

	for _, shipped := range shippedFiles(t) {
		if allowed[shipped.path] || strings.HasPrefix(shipped.path, "cmd/") {
			continue
		}

		ast.Inspect(shipped.file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			switch strings.Trim(lit.Value, `"`) {
			case "TRACE", "DEBUG", "INFO", "WARN", "ERROR":
				t.Errorf("%s: log level names are interpreted in cmd/goss only",
					shipped.fset.Position(lit.Pos()))
			}

			return true
		})
	}
}

// TestConfigHasNoLogLevelField pins the removal. The field is gone rather than
// deprecated: the one surviving use was a keyed write in a composite literal,
// so a deprecation marker would have given an embedder a clean build and
// silence, while removal gives them a compile error.
func TestConfigHasNoLogLevelField(t *testing.T) {
	for _, shipped := range shippedFiles(t) {
		ast.Inspect(shipped.file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.Field:
				for _, name := range node.Names {
					require.NotEqual(t, "LogLevel", name.Name,
						"%s: the field should not exist", shipped.fset.Position(name.Pos()))
				}
			case *ast.SelectorExpr:
				require.NotEqual(t, "LogLevel", node.Sel.Name,
					"%s: nothing should read or write the field",
					shipped.fset.Position(node.Pos()))
			case *ast.KeyValueExpr:
				if key, ok := node.Key.(*ast.Ident); ok {
					require.NotEqual(t, "LogLevel", key.Name,
						"%s: nothing should set the field",
						shipped.fset.Position(key.Pos()))
				}
			}

			return true
		})
	}
}

// TestCLIHandlerIsConstructedWithStderr is the static half of the stderr
// guarantee. A behavioural test can prove where a handler given a buffer
// writes; only this can prove which writer production hands it.
func TestCLIHandlerIsConstructedWithStderr(t *testing.T) {
	calls := 0

	for _, shipped := range shippedFiles(t) {
		ast.Inspect(shipped.file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok || fn.Name != "newCLIHandler" {
				return true
			}

			calls++
			require.NotEmpty(t, call.Args)
			require.Equal(t, "os.Stderr", render(shipped.fset, call.Args[0]),
				"%s: CLI log records go to stderr, never stdout",
				shipped.fset.Position(call.Pos()))

			return true
		})
	}

	require.Equal(t, 1, calls, "production should build the CLI handler in one place")
}

// TestAlphaGateOrdering is the control-flow half of the alpha gate: it decides,
// logs the complete diagnostic and exits, in that order, and it does all of
// that before any runtime configuration is built.
func TestAlphaGateOrdering(t *testing.T) {
	var gate, wrapper string

	for _, shipped := range shippedFiles(t) {
		for _, decl := range shipped.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			switch fn.Name.Name {
			case "fatalAlphaIfNeeded":
				gate = render(shipped.fset, fn)
			case "action":
				wrapper = render(shipped.fset, fn)
			}
		}
	}

	require.NotEmpty(t, gate, "fatalAlphaIfNeeded should exist")
	require.NotEmpty(t, wrapper, "the runtime-config action wrapper should exist")

	require.Regexp(t, `(?s)alphaRejected.*log\.Print\(alphaMessage.*os\.Exit\(1\)`, gate,
		"the gate should decide, log the whole message, then exit")
	require.Regexp(t, `(?s)fatalAlphaIfNeeded.*newRuntimeConfigFromCLI`, wrapper,
		"the platform gate runs before the level is validated")
}
