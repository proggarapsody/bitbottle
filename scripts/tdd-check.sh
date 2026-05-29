#!/usr/bin/env bash
#
# tdd-check.sh — recurring-BLOCKER pre-push gate for the auto-iter TDD subagent.
#
# Codifies the design-judge BLOCKER patterns that have repeated across cycles
# (interface gate, layer discipline, --force vs --confirm, MCP test triplet).
# Runs locally inside the TDD worktree before the subagent commits + returns.
#
# Not wired into CI on purpose — TDD subagent is the only intended caller.
# Human-direct pushes (e.g. chore: BACKLOG edits) bypass this; that's fine,
# they're not the regression vector.
#
# Exit codes:
#   0 — all checks pass
#   1 — at least one BLOCKER fired
#
# Update when a new recurring BLOCKER class shows up in design-judge output.

set -euo pipefail

cd "$(dirname "$0")/.."

fail=0

# Resolve the base commit for diff-scoped checks.
BASE=$(git merge-base HEAD origin/main 2>/dev/null || echo origin/main)

echo "tdd-check: scanning diff vs $BASE"
echo

# ─── Rule 1: layer discipline ────────────────────────────────────────────────
# pkg/cmd/** must not import api/server or api/cloud. The factory package is
# the only legitimate wiring point.
violations=$(grep -rE '"github.com/proggarapsody/bitbottle/api/(server|cloud)"' pkg/cmd/ 2>/dev/null \
  | grep -v '/factory/' || true)
if [ -n "$violations" ]; then
  echo "BLOCKER: pkg/cmd imports api/server or api/cloud directly"
  echo "$violations" | sed 's/^/  /'
  echo "  → move the helper to api/backend, or import it via factory"
  echo
  fail=1
fi

# ─── Rule 2: optional-interface gate ─────────────────────────────────────────
# For every new As<X>Client helper added in api/backend, the corresponding
# Cloud client must NOT implement the interface methods (otherwise the type
# assertion in the helper succeeds and ErrUnsupportedOnHost gets deferred to
# call-time instead of being the gate result).
#
# Exception: interfaces where CloudSupport=true AND ServerSupport=true in
# AllFeatureSpecs are intentionally both-backend. List them here so the
# rule doesn't false-positive on them. Update when a new both-backend
# optional interface is added.
both_backend_ifaces="DiffClient RefComparer CommitFileClient DefaultReviewerClient PRCommitClient PRFileClient PRParticipantClient DeployKeyClient SSHKeyClient RepoEditor RepoForksLister RepoTransferClient RepoWatcherClient RepoLabelClient PRMergePreviewClient SourceWriter HostInfoClient"
new_helpers=$(git diff "$BASE"...HEAD -- api/backend/ 2>/dev/null \
  | grep -E '^\+func As[A-Z][A-Za-z]+Client\b' \
  | sed 's/.*func As\([A-Z][A-Za-z]*\)Client.*/\1/' | sort -u || true)
for helper in $new_helpers; do
  iface="${helper}Client"
  # Skip intentionally both-backend interfaces.
  if echo "$both_backend_ifaces" | grep -qw "$iface"; then
    continue
  fi
  # Extract the method names from the interface definition.
  methods=$(awk "/type ${iface} interface {/,/^}/" api/backend/*.go 2>/dev/null \
    | grep -E '^\s+[A-Z][A-Za-z]+\(' \
    | sed -E 's/^\s+([A-Z][A-Za-z]+)\(.*/\1/' || true)
  for m in $methods; do
    if grep -rE "func \(c \*Client\) ${m}\(" api/cloud/ >/dev/null 2>&1; then
      echo "BLOCKER: Cloud implements ${iface}.${m} — defeats the optional-interface gate"
      echo "  → remove the stub from api/cloud; AsXClient() will surface ErrUnsupportedOnHost"
      fail=1
    fi
  done
done

# ─── Rule 3: MCP handler triplet ─────────────────────────────────────────────
# Every new pkg/cmd/mcp/handlers_<feature>.go needs a sibling _test.go file.
new_handlers=$(git diff --name-only --diff-filter=A "$BASE"...HEAD pkg/cmd/mcp/ 2>/dev/null \
  | grep -E 'handlers_[^/]+\.go$' \
  | grep -v '_test\.go$' || true)
for h in $new_handlers; do
  test_file="${h%.go}_test.go"
  if [ ! -f "$test_file" ]; then
    echo "BLOCKER: $h missing sibling test file $test_file"
    fail=1
  fi
done

# ─── Rule 4: --force flag forbidden ──────────────────────────────────────────
# Destructive operations use --confirm (default-safe), never --force
# (default-yes-unless-vetoed). Recurring BLOCKER from cycle 21.
forced=$(grep -rnE '"--force"|StringP\("force"|BoolP\("force"' pkg/cmd/ 2>/dev/null || true)
if [ -n "$forced" ]; then
  echo "BLOCKER: --force flag in pkg/cmd (use --confirm for destructive ops)"
  echo "$forced" | sed 's/^/  /'
  fail=1
fi

# ─── Rule 5: strict gofmt residue check ──────────────────────────────────────
# gofmt -w ./... in the self-check leaves no diff IFF every file was already
# touched and reformatted. A non-empty `gofmt -l ./...` afterward means a file
# slipped through — typically a newly-created file the subagent forgot to
# format. Cycle 83 (JSON-STABILITY) shipped a `style: gofmt …` follow-up
# commit because the strict check was missing; CI lint catches it but only
# after a 1.5-minute CI cycle. Catch it here instead.
stray_fmt=$(gofmt -l ./... 2>/dev/null || true)
if [ -n "$stray_fmt" ]; then
  echo "BLOCKER: gofmt -l flagged files (run 'gofmt -w' and re-commit):"
  echo "$stray_fmt" | sed 's/^/  /'
  fail=1
fi

# ─── Rule 6: SKILL.md skim heuristic ─────────────────────────────────────────
# New command files in pkg/cmd/ should map to entries in skills/SKILL.md.
# This is a hint, not a hard fail — perfect coverage is hard to autodetect.
new_cmd_dirs=$(git diff --name-only --diff-filter=A "$BASE"...HEAD pkg/cmd/ 2>/dev/null \
  | grep -E '\.go$' | grep -v '_test\.go$' \
  | grep -v 'handlers_\|tools_\|/factory/\|/internal/' \
  | xargs -I {} dirname {} 2>/dev/null | sort -u || true)
hint_skills=""
for d in $new_cmd_dirs; do
  # Last path segment is usually the command name.
  cmd_name=$(basename "$d")
  if [ -n "$cmd_name" ] && ! grep -q "$cmd_name" skills/SKILL.md 2>/dev/null; then
    hint_skills+="  $d → maybe add '$cmd_name' to skills/SKILL.md\n"
  fi
done
if [ -n "$hint_skills" ]; then
  echo "HINT: new commands without obvious SKILL.md entry (verify manually):"
  printf "%b" "$hint_skills"
  echo "  (not a BLOCKER — false positives common; double-check skills/SKILL.md + skills/references/)"
  echo
fi

# ─── Final ───────────────────────────────────────────────────────────────────
if [ "$fail" -eq 0 ]; then
  echo "tdd-check: pass"
else
  echo "tdd-check: FAIL — fix BLOCKERs above and re-commit before pushing"
fi
exit "$fail"
