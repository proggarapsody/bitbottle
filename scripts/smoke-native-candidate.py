#!/usr/bin/env python3
"""Smoke a native snapshot archive; never contact a real Bitbucket instance."""

import argparse
import importlib.util
import json
import os
from pathlib import Path
import platform
import tarfile
import tempfile
import zipfile


spec = importlib.util.spec_from_file_location(
    "candidate", Path(__file__).with_name("smoke-release-candidate.py"))
candidate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(candidate)


def validate(archive, checksums, metadata, expected_sha):
    if len(expected_sha) != 40 or metadata.get("commit") != expected_sha:
        raise RuntimeError("snapshot metadata commit differs from full workflow SHA")
    entries = dict(line.split(maxsplit=1)[::-1] for line in checksums.read_text().splitlines())
    if entries.get(archive.name) != candidate.digest(archive):
        raise RuntimeError("archive digest differs from checksums.txt")


def unpack(archive, root):
    binary_name = "bitbottle.exe" if archive.suffix == ".zip" else "bitbottle"
    expected = {binary_name, "README.md", "LICENSE"}
    if archive.suffix == ".zip":
        with zipfile.ZipFile(archive) as bundle:
            if set(bundle.namelist()) != expected:
                raise RuntimeError("unexpected ZIP content")
            data = bundle.read(binary_name)
    else:
        with tarfile.open(archive, "r:gz") as bundle:
            members = bundle.getmembers()
            if {item.name for item in members} != expected or any(not item.isfile() for item in members):
                raise RuntimeError("unexpected tar content")
            data = bundle.extractfile(binary_name).read()
    binary = root / binary_name
    binary.write_bytes(data)
    binary.chmod(0o755)
    return binary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("archive", "checksums", "metadata", "output"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--expected-sha", required=True)
    parser.add_argument("--target", required=True)
    parser.add_argument("--candidate-manifest", type=Path)
    args = parser.parse_args()
    native_os = {"Linux": "linux", "Darwin": "darwin", "Windows": "windows"}[platform.system()]
    native_arch = {"x86_64": "amd64", "AMD64": "amd64", "arm64": "arm64", "aarch64": "arm64"}[platform.machine()]
    if args.target != f"{native_os}_{native_arch}":
        raise RuntimeError("runner architecture differs from declared native target")
    metadata = json.loads(args.metadata.read_text())
    validate(args.archive, args.checksums, metadata, args.expected_sha)
    manifest_digest = None
    if args.candidate_manifest:
        manifest_digest = candidate.validate_sealed_manifest(args.candidate_manifest, args.archive,
                                                              metadata, args.target)
    with tempfile.TemporaryDirectory(prefix="bitbottle-native-") as scratch:
        root = Path(scratch)
        binary = unpack(args.archive, root)
        env = {key: value for key, value in os.environ.items()
               if not key.startswith(("BB_", "BITBOTTLE_"))}
        home = root / "home"
        home.mkdir()
        env.update(HOME=str(home), USERPROFILE=str(home), XDG_CONFIG_HOME=str(home),
                   APPDATA=str(home), LOCALAPPDATA=str(home), BB_CONFIG_DIR=str(home))
        version = candidate.run(str(binary), "--version", env=env).strip()
        if candidate.candidate_version(version) != metadata.get("version"):
            raise RuntimeError("binary version differs from snapshot metadata")
        if "USAGE" not in candidate.run(str(binary), "--help", env=env):
            raise RuntimeError("CLI help missing usage")
        calls = candidate.read_only_provider_call([str(binary)], env, root / "fixture-config")
        tools = candidate.mcp([str(binary), "mcp", "serve"], env)
    args.output.write_text(json.dumps({
        "commit_sha": args.expected_sha, "target": args.target,
        "archive_sha256": candidate.digest(args.archive), "version": version,
        "candidate_version": metadata.get("version"),
        "candidate_manifest_sha256": manifest_digest,
        "fixture_calls": calls, "mcp_tools": tools,
        "evidence_scope": "sealed-candidate; publish consumes same files" if args.candidate_manifest else "snapshot-only; publication rebuilds separately",
    }, indent=2) + "\n")


if __name__ == "__main__":
    main()
