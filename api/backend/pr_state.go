package backend

import (
	"fmt"
	"strings"
)

// terminalPRStates is the set of PR states that no longer accept mutating
// review/edit operations. Bitbucket Cloud returns HTTP 200 for participant
// approval regardless of state, so this guard must run client-side before the
// mutation request.
var terminalPRStates = map[string]struct{}{
	"DECLINED":   {},
	"MERGED":     {},
	"SUPERSEDED": {},
}

// ValidateMutablePRState returns a typed *DomainError (Kind ErrConflict) when
// the pull request is in a terminal state (DECLINED, MERGED, SUPERSEDED) and
// therefore cannot accept review or edit mutations. It returns nil for open or
// unknown states so callers may proceed.
//
// Call this from approve / unapprove / request-changes / edit BEFORE issuing
// the mutation, after fetching the PR via GetPR.
func ValidateMutablePRState(pr PullRequest) error {
	state := strings.ToUpper(strings.TrimSpace(pr.State))
	if _, terminal := terminalPRStates[state]; !terminal {
		return nil
	}
	return &DomainError{
		Kind:     ErrConflict,
		Code:     CodeInvalidRequest,
		Resource: "pull-request",
		ID:       fmt.Sprintf("%d", pr.ID),
		Message:  fmt.Sprintf("pull request #%d is %s; it cannot be modified", pr.ID, state),
	}
}
