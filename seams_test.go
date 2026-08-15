package goss

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSeamsHaveNoTestOnlyBranch guards the injection points. Every function
// named here exists partly so that a test can reach production behaviour
// without driving a whole binary, which is exactly the situation that tempts a
// "if testing then" shortcut. A seam with a test-only branch tests the branch
// instead of the product.
// The check is static and over the shipped tree, because the thing being ruled
// out is a branch no test would ever take.
func TestSeamsHaveNoTestOnlyBranch(t *testing.T) {
	t.Parallel()

	// Seam name to the file it must be declared in, so that a renamed or moved
	// seam fails here rather than silently dropping out of the inventory.
	seams := map[string]string{
		"newServer":               "serve.go",
		"newApp":                  "cmd/goss/goss.go",
		"alphaRejected":           "cmd/goss/goss.go",
		"alphaMessage":            "cmd/goss/goss.go",
		"newCLIHandler":           "cmd/goss/logging.go",
		"defaultOperations":       "cmd/goss/goss.go",
		"newRuntimeConfigFromCLI": "cmd/goss/goss.go",
	}

	// The runtime-config wrappers are methods on the dependency set, so they are
	// matched by receiver as well as by name.
	methodSeams := map[string]string{
		"action":    "cmd/goss/goss.go",
		"addAction": "cmd/goss/goss.go",
	}

	// Tokens that betray a runtime test check. testing.Testing() and flag.Lookup
	// are the supported ways to ask "am I under test", os.Args[0] inspection and
	// a GOSS_TEST style variable are the improvised ones.
	forbidden := []string{
		"testing.Testing",
		"test.v",
		"test.run",
		"go-build",
		"GOSS_TEST",
		"GOSS_TESTING",
		"IsTest",
		"isTest",
		"forTest",
		"ForTest",
		"testMode",
		"TestMode",
		"testHook",
		"TestHook",
	}

	found := map[string]bool{}

	for _, shipped := range shippedFiles(t) {
		for _, decl := range shipped.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			wantFile, isSeam := seams[fn.Name.Name]
			if fn.Recv != nil && receiverTypeName(fn) == "operations" {
				wantFile, isSeam = methodSeams[fn.Name.Name]
			}
			if !isSeam {
				continue
			}
			// A method of the same name on some other type is not the seam.
			if fn.Recv != nil && receiverTypeName(fn) != "operations" {
				continue
			}
			require.Equal(t, wantFile, shipped.path,
				"seam %s is declared somewhere unexpected", fn.Name.Name)
			found[fn.Name.Name] = true

			body := seamSource(t, shipped, fn)
			for _, token := range forbidden {
				require.NotContains(t, body, token,
					"seam %s looks like it branches on being under test", fn.Name.Name)
			}
		}
	}

	for name := range seams {
		require.True(t, found[name], "seam %s was not found in the shipped tree", name)
	}
	for name := range methodSeams {
		require.True(t, found[name], "seam operations.%s was not found in the shipped tree", name)
	}
}

// TestUTCHandlerHasNoTestOnlyBranch covers the remaining seam, which is a type
// rather than a function: the handler wrapper exists so a test can assert on
// rendered bytes, and its Handle must do the same thing in both cases.
func TestUTCHandlerHasNoTestOnlyBranch(t *testing.T) {
	t.Parallel()

	var checked int

	for _, shipped := range shippedFiles(t) {
		if shipped.path != "cmd/goss/logging.go" {
			continue
		}
		for _, decl := range shipped.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}
			if !strings.Contains(receiverTypeName(fn), "utcHandler") {
				continue
			}
			checked++
			body := seamSource(t, shipped, fn)
			for _, token := range []string{"testing.Testing", "GOSS_TEST", "testMode", "isTest"} {
				require.NotContains(t, body, token,
					"utcHandler.%s branches on being under test", fn.Name.Name)
			}
			// The wrapper's whole job is normalising to UTC, so a method that stopped
			// doing it would leave the UTC test resting on nothing.
			if fn.Name.Name == "Handle" {
				require.Contains(t, body, "UTC()",
					"utcHandler.Handle should be the thing that normalises the timestamp")
			}
		}
	}

	require.NotZero(t, checked, "the UTC handler wrapper was not found")
}

// seamSource returns the source text of fn, comments and all.
func seamSource(t *testing.T, shipped shippedFile, fn *ast.FuncDecl) string {
	t.Helper()

	start := shipped.fset.Position(fn.Pos())
	end := shipped.fset.Position(fn.End())

	body := readDoc(t, shipped.path)
	lines := strings.Split(body, "\n")
	require.LessOrEqual(t, end.Line, len(lines))

	return strings.Join(lines[start.Line-1:end.Line], "\n")
}

// receiverTypeName renders fn's receiver type, pointer star and all.
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	switch recv := fn.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := recv.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return recv.Name
	}

	return ""
}
