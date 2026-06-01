package enumflag

import (
	"strings"
	"testing"
)

func TestSet_AcceptsAllowedValue(t *testing.T) {
	var target string
	v := New([]string{"open", "closed", "merged"}, &target, false)
	if err := v.Set("closed"); err != nil {
		t.Fatalf("Set(closed) returned error: %v", err)
	}
	if target != "closed" {
		t.Fatalf("target = %q, want closed", target)
	}
	if v.String() != "closed" {
		t.Fatalf("String() = %q, want closed", v.String())
	}
}

func TestSet_RejectsOffEnum(t *testing.T) {
	var target string
	v := New([]string{"open", "closed"}, &target, false)
	err := v.Set("INVALID")
	if err == nil {
		t.Fatal("Set(INVALID) returned nil, want error")
	}
	if !strings.Contains(err.Error(), "open, closed") {
		t.Fatalf("error %q does not list allowed values", err.Error())
	}
}

func TestSet_RejectsEmptyString(t *testing.T) {
	var target string
	v := New([]string{"open", "closed"}, &target, false)
	if err := v.Set(""); err == nil {
		t.Fatal("Set(\"\") returned nil, want error")
	}
}

func TestSet_CaseInsensitiveMatchPreservesInput(t *testing.T) {
	var target string
	v := New([]string{"open", "merged"}, &target, true)
	if err := v.Set("MERGED"); err != nil {
		t.Fatalf("Set(MERGED) returned error: %v", err)
	}
	if target != "MERGED" {
		t.Fatalf("target = %q, want original MERGED preserved", target)
	}
}

func TestSet_CaseSensitiveRejectsWrongCase(t *testing.T) {
	var target string
	v := New([]string{"open"}, &target, false)
	if err := v.Set("OPEN"); err == nil {
		t.Fatal("Set(OPEN) with caseInsensitive=false returned nil, want error")
	}
}

func TestType(t *testing.T) {
	var target string
	if got := New(nil, &target, false).Type(); got != "string" {
		t.Fatalf("Type() = %q, want string", got)
	}
}
