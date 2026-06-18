package testhelpers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// specPath returns the absolute path to the vendored OpenAPI spec for the given backend.
// backend must be "server" or "cloud".
func specPath(backend string) (string, error) {
	switch backend {
	case "server", "cloud":
		// __FILE__ is test/testhelpers/openapi.go; go up two dirs to reach the module root.
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return "", fmt.Errorf("openapi: runtime.Caller failed")
		}
		moduleRoot := filepath.Join(filepath.Dir(filename), "..", "..")
		return filepath.Join(moduleRoot, "api", backend, "gen", "openapi.yaml"), nil
	default:
		return "", fmt.Errorf("openapi: unknown backend %q (must be \"server\" or \"cloud\")", backend)
	}
}

// ValidateAgainstSchemaErr validates value (marshaled to JSON) against the named
// component schema in the vendored OpenAPI spec for the given backend
// ("server" or "cloud"). Returns a non-nil error on validation failure or
// unknown schema/backend. Does not call t.
func ValidateAgainstSchemaErr(schemaName string, value any, backend string) error {
	path, err := specPath(backend)
	if err != nil {
		return err
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("openapi: load spec %s: %w", path, err)
	}

	schemaRef, ok := doc.Components.Schemas[schemaName]
	if !ok {
		return fmt.Errorf("openapi: schema %q not found in %s spec", schemaName, backend)
	}

	// Marshal then unmarshal to get a plain any (map/slice/scalar) for VisitJSON.
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("openapi: marshal value: %w", err)
	}

	var jsonValue any
	if err := json.Unmarshal(raw, &jsonValue); err != nil {
		return fmt.Errorf("openapi: unmarshal value: %w", err)
	}

	return schemaRef.Value.VisitJSON(jsonValue)
}

// ValidateAgainstSchema validates value (marshaled to JSON) against the named
// component schema in the vendored OpenAPI spec for the given backend
// ("server" or "cloud"). Calls t.Fatal on validation error or unknown schema/backend.
func ValidateAgainstSchema(t testing.TB, schemaName string, value any, backend string) {
	t.Helper()
	if err := ValidateAgainstSchemaErr(schemaName, value, backend); err != nil {
		t.Fatalf("openapi schema validation failed for %s/%s: %v", backend, schemaName, err)
	}
}
