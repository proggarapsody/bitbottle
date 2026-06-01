# ADR 0001 — Promote `--start-at` to 3rd positional (Option A)

**Date:** 2026-06-02  
**Status:** Accepted  
**Scope:** REF-UX (#621)

Option A was chosen over Option B (defaulting to HEAD) because making `--start-at` an explicit positional `[START_AT]` preserves the invariant that every branch and tag creation is deliberately anchored to a known ref — defaulting silently to HEAD would create branches at an unexpected point when the user forgets the argument, which is harder to recover from than a clear "start-at is required" error. The positional form `bb branch create PROJECT/REPO NAME SHA` mirrors `git branch NAME START_POINT` ergonomics that users already know, while keeping the `--start-at` flag for backward compatibility and for callers who prefer named flags in scripts.
