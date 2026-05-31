package repoarg

// SplitLeadingRepo separates an optional leading repository positional from
// a fixed number of trailing positionals.
//
// Commands shaped like `cmd [PROJECT/REPO] TRAILING...` accept the repo
// either as the first positional OR via -R/--repo / BB_REPO / pinned
// defaults (resolved later through factory.ResolveTarget). This helper
// discriminates purely by arg count so the repo positional can be omitted
// whenever -R supplies it:
//
//	len(args) == trailing      → no repo positional; repoArgs is nil
//	len(args) == trailing + 1  → args[0] is the repo; repoArgs = args[:1]
//
// Pair it with cobra.RangeArgs(trailing, trailing+1). The returned
// repoArgs is meant to be passed straight to factory.ResolveTarget, which
// falls back to f.BaseRepo() when it is empty.
func SplitLeadingRepo(args []string, trailing int) (repoArgs, rest []string) {
	if len(args) > trailing {
		return args[:1], args[1:]
	}
	return nil, args
}

// SplitTrailingRepo is the mirror of SplitLeadingRepo for commands shaped
// like `cmd LEADING... [PROJECT/REPO]`, where the optional repo positional
// comes last (e.g. `commit files HASH [PROJECT/REPO]`).
//
//	len(args) == leading       → no repo positional; repoArgs is nil
//	len(args) == leading + 1   → last arg is the repo; repoArgs = args[leading:]
func SplitTrailingRepo(args []string, leading int) (repoArgs, rest []string) {
	if len(args) > leading {
		return args[leading:], args[:leading]
	}
	return nil, args
}
