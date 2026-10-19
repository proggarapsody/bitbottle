// Package qamatrix emits scenario claims that the QA runner binds to actual
// go test -json pass events. A claim alone is never evidence of a passing test.
package qamatrix

import (
	"encoding/json"
	"strings"
	"testing"
)

const Prefix = "QA_MATRIX "

// Scenario identifies an observed public operation outcome. Flags also holds
// MCP input parameter names; NotApplicable gives a reason for any reviewed
// important input that cannot apply to this scenario.
type Scenario struct {
	Operation     string            `json:"operation"`
	Backend       string            `json:"backend"`
	Outcome       string            `json:"outcome"`
	Assertions    []string          `json:"assertions"`
	Flags         []string          `json:"flags,omitempty"`
	NotApplicable map[string]string `json:"not_applicable,omitempty"`
}

// Record emits one bounded JSON event; the runner accepts it only when this
// precise package and test subsequently pass in the same go test invocation.
func Record(t testing.TB, s Scenario) {
	t.Helper()
	if strings.TrimSpace(s.Operation) == "" || strings.TrimSpace(s.Backend) == "" ||
		strings.TrimSpace(s.Outcome) == "" || len(s.Assertions) == 0 {
		t.Fatalf("qa matrix scenario requires operation, backend, outcome, and assertions")
	}
	b, err := json.Marshal(s)
	if err != nil || len(b) > 8192 {
		t.Fatalf("qa matrix scenario cannot be encoded within 8192 bytes: %v", err)
	}
	t.Logf("%s%s", Prefix, b)
}
