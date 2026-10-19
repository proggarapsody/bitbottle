#!/usr/bin/env python3
"""Exercise the Linux release archive, packed npm wrapper, and Docker image."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Thread


def run(*args, env=None, cwd=None):
    result = subprocess.run(args, env=env, cwd=cwd, text=True, capture_output=True, timeout=45)
    if result.returncode:
        raise RuntimeError(f"{args!r} failed: {result.stderr or result.stdout}")
    return result.stdout


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def mcp(command, env):
    requests = [
        {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
            "protocolVersion": "2025-03-26", "capabilities": {},
            "clientInfo": {"name": "candidate-smoke", "version": "1"}}},
        {"jsonrpc": "2.0", "method": "notifications/initialized"},
        {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}},
    ]
    payload = "\n".join(json.dumps(item) for item in requests) + "\n"
    result = subprocess.run(command, input=payload, text=True, capture_output=True,
                            env=env, timeout=30)
    if result.returncode:
        raise RuntimeError(f"MCP server exited {result.returncode}: {result.stderr}")
    replies = [json.loads(line) for line in result.stdout.splitlines()]
    by_id = {reply.get("id"): reply for reply in replies if "id" in reply}
    if "serverInfo" not in by_id.get(1, {}).get("result", {}):
        raise RuntimeError(f"MCP initialize failed: {replies!r}")
    tools = by_id.get(2, {}).get("result", {}).get("tools", [])
    if not tools or not any(tool.get("name") for tool in tools):
        raise RuntimeError(f"MCP tools/list returned no tools: {replies!r}")
    return len(tools)


def inspect_archive(archive, destination):
    expected = {"bitbottle", "README.md", "LICENSE"}
    with tarfile.open(archive, "r:gz") as bundle:
        members = bundle.getmembers()
        names = {member.name for member in members}
        if names != expected or any(not member.isfile() for member in members):
            raise RuntimeError(f"unexpected archive content: {sorted(names)}")
        for member in members:
            data = bundle.extractfile(member).read()
            if member.name == "bitbottle":
                binary = destination / member.name
                binary.write_bytes(data)
                binary.chmod(0o755)
    return destination / "bitbottle"


def candidate_version(version_output):
    match = re.fullmatch(r"bitbottle version (\S+) \([^)]+\) built .+", version_output)
    if not match or match.group(1) == "dev":
        raise RuntimeError(f"unexpected candidate version: {version_output!r}")
    return match.group(1)


def read_only_provider_call(command, env, config_dir):
    class Fixture(BaseHTTPRequestHandler):
        requests = []

        def do_GET(self):
            self.requests.append((self.path, self.headers.get("Authorization")))
            if (self.path != "/2.0/user/permissions/workspaces?pagelen=30"
                    or self.headers.get("Authorization") != "Bearer candidate-fixture-token"):
                self.send_error(400)
                return
            body = json.dumps({"values": [{"workspace": {
                "uuid": "{aaaaaaaa-1111-1111-1111-111111111111}",
                "slug": "candidate-fixture", "name": "Candidate Fixture",
                "links": {"html": {"href": "https://bitbucket.org/candidate-fixture/"}}
            }}]}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, *_args):
            pass

    config_dir.mkdir()
    (config_dir / "hosts.yml").write_text(
        "bitbucket.org:\n  backend_type: cloud\n  git_protocol: https\n")
    server = ThreadingHTTPServer(("127.0.0.1", 0), Fixture)
    thread = Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        fixture_env = dict(env, BB_CONFIG_DIR=str(config_dir), BB_TOKEN="candidate-fixture-token",
                           BB_CLOUD_BASE_URL=f"http://127.0.0.1:{server.server_port}/2.0")
        output = run(*command, "workspace", "list", "--hostname", "bitbucket.org", env=fixture_env)
        if "candidate-fixture" not in output or len(Fixture.requests) != 1:
            raise RuntimeError(f"read-only Cloud fixture was not exercised: {Fixture.requests!r}")
        return len(Fixture.requests)
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=5)


def main():
    parser = argparse.ArgumentParser(
        description="Smoke a GoReleaser snapshot archive and its packed npm wrapper.",
        epilog="Example: smoke-release-candidate.py --archive dist/bitbottle_linux_amd64.tar.gz "
               "--checksums dist/checksums.txt --metadata dist/metadata.json "
               "--npm-tarball wrapper.tgz --docker-image image:tag --artifacts dist/artifacts.json")
    parser.add_argument("--archive", type=Path, required=True)
    parser.add_argument("--npm-tarball", type=Path, required=True)
    parser.add_argument("--checksums", type=Path, required=True)
    parser.add_argument("--metadata", type=Path, required=True)
    parser.add_argument("--docker-image")
    parser.add_argument("--artifacts", type=Path)
    args = parser.parse_args()
    archive = args.archive.resolve()
    npm_tarball = args.npm_tarball.resolve()
    release_metadata = json.loads(args.metadata.read_text())
    expected_version = release_metadata.get("version")
    if not isinstance(expected_version, str) or not expected_version:
        raise RuntimeError("GoReleaser metadata lacks a snapshot version")
    with tarfile.open(npm_tarball, "r:gz") as bundle:
        names = {member.name for member in bundle.getmembers() if member.isfile()}
        expected = {"package/package.json", "package/install.js", "package/bin/run.js", "package/README.md"}
        if names != expected:
            raise RuntimeError(f"unexpected npm package content: {sorted(names)}")
        package = json.load(bundle.extractfile("package/package.json"))
    if package.get("version") != expected_version:
        raise RuntimeError(f"npm version {package.get('version')!r} differs from GoReleaser snapshot {expected_version!r}")
    checksums = dict(line.split(maxsplit=1)[::-1] for line in args.checksums.read_text().splitlines())
    if checksums.get(archive.name) != digest(archive):
        raise RuntimeError("archive digest differs from GoReleaser checksums.txt")
    with tempfile.TemporaryDirectory(prefix="bitbottle-candidate-") as scratch:
        root = Path(scratch)
        binary = inspect_archive(archive, root)
        env = {key: value for key, value in os.environ.items()
               if not key.startswith(("BB_", "BITBOTTLE_"))}
        env["HOME"] = str(root / "home")
        env["XDG_CONFIG_HOME"] = str(root / "config")
        (root / "home").mkdir()
        version = run(str(binary), "--version", env=env).strip()
        if candidate_version(version) != expected_version:
            raise RuntimeError(f"binary version differs from GoReleaser snapshot: {version!r}")
        if "USAGE" not in run(str(binary), "--help", env=env):
            raise RuntimeError("CLI help missing usage")
        provider_calls = read_only_provider_call([str(binary)], env, root / "native-config")
        tools = mcp([str(binary), "mcp", "serve"], env)

        # Intercept only the release download; npm executes the unchanged packed
        # postinstall script against the candidate archive, without a public release.
        shim = root / "candidate-download.cjs"
        shim.write_text("""const https = require('https');
const fs = require('fs');
const {PassThrough} = require('stream');
https.get = (url, callback) => {
  const expected = process.env.CANDIDATE_RELEASE_URL;
  if (url !== expected) throw new Error('unexpected release URL: ' + url);
  const response = fs.createReadStream(process.env.CANDIDATE_ARCHIVE);
  response.statusCode = 200;
  process.nextTick(() => callback(response));
  return new PassThrough();
};
""")
        npm_env = dict(env, NODE_OPTIONS=f"--require={shim}",
                       CANDIDATE_ARCHIVE=str(archive),
                       CANDIDATE_RELEASE_URL=f"https://github.com/proggarapsody/bitbottle/releases/download/v{expected_version}/{archive.name}",
                       BITBOTTLE_SKIP_SKILL_INSTALL="1", CI="true")
        install_dir = root / "npm-install"
        install_dir.mkdir()
        run("npm", "install", "--offline", "--no-audit", "--no-fund",
            "--prefix", str(install_dir), str(npm_tarball), env=npm_env)
        wrapper = install_dir / "node_modules/@proggarapsody/bitbottle/bin/run.js"
        installed = wrapper.with_name("bitbottle")
        if digest(installed) != digest(binary):
            raise RuntimeError("npm installed binary differs from release archive")
        run("node", str(wrapper), "--help", env=env)
        provider_calls += read_only_provider_call(["node", str(wrapper)], env, root / "npm-config")
        mcp(["node", str(wrapper), "mcp", "serve"], env)

        image_digest = None
        if args.docker_image:
            if args.artifacts:
                artifacts = json.loads(args.artifacts.read_text())
                matches = [item for item in artifacts if item.get("type") == "Docker Image"
                           and item.get("name") == args.docker_image
                           and item.get("goos") == "linux" and item.get("goarch") == "amd64"]
                if len(matches) != 1:
                    raise RuntimeError("Docker image is not the Linux amd64 GoReleaser artifact")
            image_digest = run("docker", "image", "inspect", "--format", "{{.Id}}",
                               args.docker_image).strip()
            if not image_digest.startswith("sha256:"):
                raise RuntimeError(f"invalid Docker image ID: {image_digest!r}")
            container_id = run("docker", "create", args.docker_image).strip()
            try:
                image_binary = root / "image-bitbottle"
                run("docker", "cp", f"{container_id}:/bitbottle", str(image_binary))
                if digest(image_binary) != digest(binary):
                    raise RuntimeError("GoReleaser Docker binary differs from release archive")
            finally:
                run("docker", "rm", "-f", container_id)
            mcp(["docker", "run", "--rm", "-i", "--pull=never", args.docker_image], env)
        print(json.dumps({"archive": archive.name, "archive_sha256": digest(archive),
                          "npm_sha256": digest(npm_tarball), "binary_sha256": digest(binary),
                          "version": version, "candidate_version": expected_version,
                          "mcp_tools": tools,
                          "read_only_fixture_calls": provider_calls,
                          "docker_image": args.docker_image,
                          "docker_image_id": image_digest,
                          "evidence_scope": "snapshot-only; publication rebuilds separately"}))


if __name__ == "__main__":
    main()
