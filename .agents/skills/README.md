# Matt Pocock skills

Project-local copies of all 27 stable engineering and productivity skills from
[mattpocock/skills](https://github.com/mattpocock/skills), pinned to commit
`4588b32ecab9ecc9fc8cc6b6c5e7d675b6004b0d` on 2026-10-06.
Experimental and miscellaneous skills are excluded. The stable set includes
cross-skill dependencies such as `grilling` and `writing-for-agents`.

Installed using Codex's `install-skill-from-github.py` with `--dest .agents/skills`
and the pinned commit as `--ref`. Complete skill folders and Codex invocation
metadata are preserved. [LICENSE](LICENSE) contains the upstream MIT notice.
`pr/CREDITS.md` preserves that skill's additional attribution.

These are development skills. The root `skills/` directory is the Bitbottle
consumer skill shipped to users; keep the two separate.

Skills marked `disable-model-invocation: true` require an explicit user invocation.
Installation does not authorize issue creation, messages, merges, or releases.
Follow repository conventions and the user's current scope when using a skill.
Tracker, triage-label, and domain conventions are configured in `docs/agents/`; see [the project knowledge map](../../docs/agents/README.md).

For updates, inspect a new upstream commit in a temporary directory, compare
against `sources.json`, preserve intentional local edits, and replace selected
folders explicitly. This installation does not use a skills.sh update lockfile.
Do not run upstream's global maintainer linking script.
