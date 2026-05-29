package contract_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoServerMapLiteralBodies asserts that no call to putJSON, postJSON, or
// patchJSON (on the *Client receiver or any httpx.Transport) passes a map
// composite literal as the body argument (position 1, zero-indexed).
//
// Map literals lose static type safety and make OCC version field injection
// error-prone; named or anonymous structs must be used instead.
func TestNoServerMapLiteralBodies(t *testing.T) {
	t.Helper()

	// Locate api/server relative to the module root (two levels up from this
	// file's directory: test/contract/ → test/ → repo root).
	thisDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Walk up from test/contract to the repo root.
	repoRoot := filepath.Join(thisDir, "..", "..")
	serverDir := filepath.Join(repoRoot, "api", "server")

	fset := token.NewFileSet()

	entries, err := os.ReadDir(serverDir)
	if err != nil {
		t.Fatalf("readdir %s: %v", serverDir, err)
	}

	mutationFuncs := map[string]bool{
		"putJSON":   true,
		"postJSON":  true,
		"patchJSON": true,
		"PutJSON":   true,
		"PostJSON":  true,
		"PatchJSON": true,
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			// skip gen/ and any sub-directories
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == "client.go" {
			// Defines the helpers — not a call site.
			continue
		}

		fullPath := filepath.Join(serverDir, name)
		f, err := parser.ParseFile(fset, fullPath, nil, 0)
		if err != nil {
			t.Errorf("parse %s: %v", name, err)
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Extract the function/method name from the call expression.
			funcName := ""
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				funcName = fn.Name
			case *ast.SelectorExpr:
				funcName = fn.Sel.Name
			}

			if !mutationFuncs[funcName] {
				return true
			}

			// Body is argument at index 1 (0-indexed): (path, body, dest).
			if len(call.Args) < 2 {
				return true
			}
			bodyArg := call.Args[1]

			lit, ok := bodyArg.(*ast.CompositeLit)
			if !ok {
				return true
			}
			if _, isMap := lit.Type.(*ast.MapType); !isMap {
				return true
			}

			pos := fset.Position(call.Pos())
			t.Errorf(
				"%s:%d: %s called with map literal body — use a named or anonymous struct instead",
				pos.Filename, pos.Line, funcName,
			)
			return true
		})
	}
}
