"""
Behavior tests for sync_help.py.

The audit script's job is to surface drift between the skill content
and the actual CLI. We test through its public CLI: feed it fixture
skill directories plus a fake bitbottle binary, and assert the exit
code + report text describe the drift accurately.

We intentionally do NOT mock the script's internals — what matters is
what a CI run of the script reports, not how it walks the files.
"""

from __future__ import annotations

import os
import shutil
import stat
import subprocess
import sys
import textwrap
import unittest
from pathlib import Path
from tempfile import TemporaryDirectory

SCRIPT = Path(__file__).resolve().parent / "sync_help.py"


def _write_fake_bitbottle(bin_dir: Path, version: str, list_flags: str = "") -> None:
    """Write a shell stub that mimics `bitbottle --version` and `-h`.

    No textwrap.dedent here — the `list_flags` param often has its own
    indentation that breaks dedent's common-prefix calculation, which
    would leave whitespace before `#!/bin/sh` and silently brick the
    shebang. We build the stub as a plain string instead.
    """
    bin_dir.mkdir(parents=True, exist_ok=True)
    flags = list_flags or (
        "      --jq string         Filter JSON output\n"
        "      --json string       Output JSON with specified fields\n"
        "      --limit int         Maximum number of pull requests (default 30)\n"
        "      --state string      State filter: open, closed, merged\n"
    )
    stub = bin_dir / "bitbottle"
    stub.write_text(
        "#!/bin/sh\n"
        "if [ \"$1\" = \"--version\" ]; then\n"
        f"  echo 'bitbottle version {version} (deadbeef) built 2026-01-01T00:00:00Z'\n"
        "  exit 0\n"
        "fi\n"
        "cat <<'EOF'\n"
        "FLAGS\n"
        f"{flags}"
        "      -h, --help              help for list\n"
        "EOF\n"
    )
    st = stub.stat()
    stub.chmod(st.st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)


def _run(skill_dir: Path, fake_bin_dir: Path) -> subprocess.CompletedProcess:
    env = os.environ.copy()
    # Prepend our fake bin so `bitbottle` resolves to the stub.
    env["PATH"] = f"{fake_bin_dir}{os.pathsep}{env.get('PATH', '')}"
    return subprocess.run(
        [sys.executable, str(SCRIPT), "--skill-dir", str(skill_dir)],
        env=env,
        capture_output=True,
        text=True,
    )


class SyncHelpAuditTest(unittest.TestCase):
    def test_passes_when_skill_matches_cli(self) -> None:
        with TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill = tmp_path / "skills"
            skill.mkdir()
            (skill / "SKILL.md").write_text(
                "Verified against bitbottle 9.9.9.\n"
                "last verified against **9.9.9**.\n"
            )
            _write_fake_bitbottle(tmp_path / "bin", version="9.9.9")
            r = _run(skill, tmp_path / "bin")
            self.assertEqual(r.returncode, 0,
                             f"clean skill should pass; got: {r.stdout}{r.stderr}")

    def test_flags_stale_version_label(self) -> None:
        with TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill = tmp_path / "skills"
            skill.mkdir()
            (skill / "SKILL.md").write_text(
                "Verified against bitbottle 1.13.1.\n"
            )
            _write_fake_bitbottle(tmp_path / "bin", version="9.9.9")
            r = _run(skill, tmp_path / "bin")
            self.assertNotEqual(r.returncode, 0,
                                "stale version must fail the audit")
            self.assertIn("1.13.1", r.stdout + r.stderr)
            self.assertIn("9.9.9", r.stdout + r.stderr)

    def test_flags_phantom_flag_that_now_exists(self) -> None:
        # SKILL.md claims --limit doesn't exist, but the CLI lists it.
        # The audit must catch this contradiction.
        with TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill = tmp_path / "skills"
            skill.mkdir()
            (skill / "SKILL.md").write_text(
                "Verified against bitbottle 9.9.9.\n"
                "Flags that DO NOT exist: `--limit`, `--mine`.\n"
            )
            _write_fake_bitbottle(tmp_path / "bin", version="9.9.9")
            r = _run(skill, tmp_path / "bin")
            self.assertNotEqual(r.returncode, 0,
                                "phantom-flag claim that's actually real must fail")
            self.assertIn("--limit", r.stdout + r.stderr)


class RealSkillTest(unittest.TestCase):
    """Runs the audit against the REAL skill dir + real `bitbottle` on
    PATH. This is the test that catches drift in CI / dev. Skipped if
    `bitbottle` isn't installed locally."""

    def test_real_skill_is_in_sync(self) -> None:
        if shutil.which("bitbottle") is None:
            self.skipTest("bitbottle binary not on PATH")
        repo_root = Path(__file__).resolve().parents[2]
        skill = repo_root / "skills"
        r = subprocess.run(
            [sys.executable, str(SCRIPT), "--skill-dir", str(skill)],
            capture_output=True, text=True,
        )
        self.assertEqual(r.returncode, 0,
                         f"real skill out of sync:\n{r.stdout}\n{r.stderr}")


if __name__ == "__main__":
    unittest.main()
