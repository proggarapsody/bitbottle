#!/usr/bin/env python3
"""
sync_help — drift detector for the bitbottle agent skill.

What this catches
-----------------
The skill is a static document; the CLI evolves. After two rounds of
"the skill says X but `-h` says Y" audits, we want a script that
fails fast when those drift apart. Cheap to run, no deps.

Checks performed
----------------
1. Version drift — every `Verified against bitbottle X.Y.Z` and
   `last verified against **X.Y.Z**` line in the skill must equal
   `bitbottle --version`.
2. Phantom-flag drift — the skill's "Flags that DO NOT exist" list
   must NOT include flags that actually appear in the relevant
   `bitbottle <cmd> -h` output.

Exit code is non-zero if any check fails. Output is plain text so it
shows up cleanly in CI logs.

Run from anywhere:

    python3 skills/scripts/sync_help.py
    python3 skills/scripts/sync_help.py --skill-dir /path/to/skills
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path
from typing import Iterable

# Commands whose `-h` we scan for phantom-flag claims. Matches the
# commands that the skill's flag-reality section talks about.
COMMANDS_TO_SCAN = [
    ["pr", "list"],
    ["pr", "create"],
    ["pr", "merge"],
    ["repo", "list"],
    ["branch", "list"],
    ["tag", "list"],
    ["pipeline", "list"],
    ["commit", "log"],
    ["api"],
]

VERSION_PATTERNS = [
    re.compile(r"[Vv]erified against bitbottle (\d+\.\d+\.\d+)"),
    re.compile(r"last verified against \*\*(\d+\.\d+\.\d+)\*\*"),
]

PHANTOM_LINE = re.compile(
    r"[Ff]lags that DO NOT exist[^:]*:\s*(.+)", re.IGNORECASE
)
FLAG_TOKEN = re.compile(r"`?(--[a-z][a-z0-9-]*)`?")


def cli_version() -> str:
    out = subprocess.run(
        ["bitbottle", "--version"], capture_output=True, text=True, check=True
    ).stdout
    m = re.search(r"(\d+\.\d+\.\d+)", out)
    if not m:
        raise SystemExit(f"could not parse bitbottle --version output: {out!r}")
    return m.group(1)


def cli_help(args: list[str]) -> str:
    return subprocess.run(
        ["bitbottle", *args, "-h"], capture_output=True, text=True
    ).stdout


def iter_md(skill_dir: Path) -> Iterable[Path]:
    for p in [skill_dir / "SKILL.md", *sorted(skill_dir.glob("references/*.md"))]:
        if p.exists():
            yield p


def check_versions(skill_dir: Path, version: str) -> list[str]:
    problems: list[str] = []
    for path in iter_md(skill_dir):
        text = path.read_text()
        for pat in VERSION_PATTERNS:
            for m in pat.finditer(text):
                found = m.group(1)
                if found != version:
                    problems.append(
                        f"{path.relative_to(skill_dir)}: claims version "
                        f"{found}, but bitbottle --version is {version}"
                    )
    return problems


def collect_real_flags() -> set[str]:
    """Union of every long flag listed by every help screen we scan."""
    flags: set[str] = set()
    for cmd in COMMANDS_TO_SCAN:
        try:
            out = cli_help(cmd)
        except FileNotFoundError:
            return flags
        for line in out.splitlines():
            for tok in FLAG_TOKEN.finditer(line):
                flags.add(tok.group(1))
    return flags


def check_phantom_flags(skill_dir: Path, real_flags: set[str]) -> list[str]:
    problems: list[str] = []
    for path in iter_md(skill_dir):
        text = path.read_text()
        for m in PHANTOM_LINE.finditer(text):
            claimed = {tok.group(1) for tok in FLAG_TOKEN.finditer(m.group(1))}
            for flag in sorted(claimed & real_flags):
                problems.append(
                    f"{path.relative_to(skill_dir)}: claims `{flag}` does "
                    f"not exist, but it appears in `bitbottle … -h`"
                )
    return problems


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[1])
    default_dir = Path(__file__).resolve().parents[1]
    ap.add_argument(
        "--skill-dir",
        type=Path,
        default=default_dir,
        help=f"Path to the skill directory (default: {default_dir})",
    )
    args = ap.parse_args()

    skill_dir: Path = args.skill_dir
    if not skill_dir.is_dir():
        print(f"skill dir not found: {skill_dir}", file=sys.stderr)
        return 2

    try:
        version = cli_version()
    except FileNotFoundError:
        print("bitbottle not on PATH; install it before running this audit",
              file=sys.stderr)
        return 2

    problems: list[str] = []
    problems += check_versions(skill_dir, version)
    problems += check_phantom_flags(skill_dir, collect_real_flags())

    if problems:
        print(f"sync_help: {len(problems)} drift issue(s) detected\n")
        for p in problems:
            print(f"  - {p}")
        print(f"\nCLI version: {version}")
        return 1

    print(f"sync_help: skill is in sync with bitbottle {version}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
