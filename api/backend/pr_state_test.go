package backend

import (
	"errors"
	"testing"
)

func TestValidateMutablePRState_AllowsOpen(t *testing.T) {
	for _, st := range []string{"OPEN", "open", "", "DRAFT", "weird"} {
		if err := ValidateMutablePRState(PullRequest{ID: 1, State: st}); err != nil {
			t.Errorf("state %q: got error %v, want nil", st, err)
		}
	}
}

func TestValidateMutablePRState_RejectsTerminal(t *testing.T) {
	for _, st := range []string{"DECLINED", "MERGED", "SUPERSEDED", "declined", " merged "} {
		err := ValidateMutablePRState(PullRequest{ID: 7, State: st})
		if err == nil {
			t.Errorf("state %q: got nil, want conflict error", st)
			continue
		}
		if !errors.Is(err, ErrConflict) {
			t.Errorf("state %q: error Kind = %v, want ErrConflict", st, err)
		}
		var de *DomainError
		if !errors.As(err, &de) {
			t.Errorf("state %q: error is not *DomainError", st)
			continue
		}
		if de.Resource != "pull-request" {
			t.Errorf("state %q: Resource = %q, want pull-request", st, de.Resource)
		}
	}
}
