#!/usr/bin/env python3
"""Seal, verify and promote the same release bytes. Promotion defaults to a plan."""
import argparse
import base64
import hashlib
import json
import math
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys
import tarfile
import tempfile
import time

from release_oci import inspect_layout

TARGETS = ('linux/amd64', 'linux/arm64', 'darwin/amd64', 'darwin/arm64', 'windows/amd64')
MANIFEST = 'candidate-manifest.json'
DC_NOTE = 'DC не проверен live'
REPOSITORY = 'proggarapsody/bitbottle'


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def read_json(path):
    return json.loads(path.read_text())


def files(root):
    result = {}
    require(root.is_dir() and not root.is_symlink(), 'candidate directory missing or unsafe')
    for path in sorted(root.rglob('*')):
        require(not path.is_symlink(), f'symlink is unsupported: {path}')
        if path.is_dir():
            continue
        require(path.is_file(), f'unsupported file: {path}')
        name = path.relative_to(root).as_posix()
        if name != MANIFEST:
            result[name] = digest(path)
    return result


def safe_path(root, name):
    require(isinstance(name, str) and '\\' not in name, 'invalid candidate path')
    path = PurePosixPath(name)
    require(not path.is_absolute() and '..' not in path.parts and name == path.as_posix(),
            f'unsafe candidate path: {name}')
    require(name and name != '.', 'empty candidate path')
    return root / name


def describe(root, sha, tag):
    require(re.fullmatch(r'[0-9a-f]{40}', sha), 'expected full 40-character commit SHA')
    require(re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag), 'unsupported release tag')
    version = tag[1:]
    metadata = read_json(root / 'dist/metadata.json')
    require(metadata.get('version') == version and metadata.get('commit') == sha,
            'GoReleaser version or full commit differs from candidate')
    require(isinstance(read_json(root / 'dist/artifacts.json'), list), 'invalid GoReleaser artifact inventory')
    require(DC_NOTE in (root / 'release-notes.md').read_text(), 'release notes lack literal DC caveat')
    sbom = read_json(root / 'sbom.spdx.json')
    require(sbom.get('spdxVersion') == 'SPDX-2.3' and sbom.get('SPDXID') == 'SPDXRef-DOCUMENT'
            and isinstance(sbom.get('packages'), list), 'candidate lacks valid SPDX SBOM')
    checksums = {}
    for line in (root / 'dist/checksums.txt').read_text().splitlines():
        checksum, name = line.split(maxsplit=1)
        name = name.removeprefix('*')
        require(name not in checksums and re.fullmatch(r'[0-9a-f]{64}', checksum), 'invalid checksum inventory')
        path = safe_path(root / 'dist', name)
        require(path.is_file() and digest(path) == checksum, f'checksum mismatch: {name}')
        checksums[name] = checksum
    archives = {}
    for target in TARGETS:
        name = 'bitbottle_' + target.replace('/', '_') + ('.zip' if target.startswith('windows/') else '.tar.gz')
        require(name in checksums, f'missing native archive: {name}')
        archives[target] = {'path': 'dist/' + name, 'sha256': checksums[name]}
    for arch in ('amd64', 'arm64'):
        for suffix in ('deb', 'rpm', 'apk'):
            # nfpm chooses format-specific version/architecture filenames.
            matches = [name for name in checksums if name.endswith('.' + suffix)
                       and (arch in name or (arch == 'amd64' and 'x86_64' in name)
                            or (arch == 'arm64' and 'aarch64' in name))]
            require(len(matches) == 1, f'missing or ambiguous {arch} {suffix} package')
    npm = list((root / 'npm').glob('*.tgz'))
    require(len(npm) == 1, 'expected one already packed npm tarball')
    with tarfile.open(npm[0], 'r:gz') as archive:
        package = json.load(archive.extractfile('package/package.json'))
    require(package.get('name') == '@proggarapsody/bitbottle' and package.get('version') == version,
            'npm package identity or version differs from release')
    docker = inspect_layout(root / 'docker')
    docker_manifests = {'linux/' + arch: image['manifest_digest'] for arch, image in docker['manifests'].items()}
    return {'schema': 1, 'commit': sha, 'tag': tag, 'version': version,
            'archives': archives, 'npm': npm[0].relative_to(root).as_posix(),
            'docker_digest': docker['index_digest'], 'docker_manifests': docker_manifests, 'files': files(root)}


def verify(root, sha, tag, evidence):
    manifest_path = root / MANIFEST
    require(manifest_path.is_file() and not manifest_path.is_symlink(), 'candidate manifest missing or unsafe')
    manifest = read_json(manifest_path)
    for name in manifest.get('files', {}):
        safe_path(root, name)
    require(manifest == describe(root, sha, tag), 'candidate manifest or sealed bytes changed')
    for platform in TARGETS:
        target = platform.replace('/', '_')
        report = read_json(evidence / ('native-smoke-' + target + '.json'))
        check_report(report, manifest, platform)
        require(report.get('candidate_manifest_sha256') == digest(manifest_path), 'verification manifest mismatch')
        require(report.get('fixture_calls', 0) >= 1, 'native evidence lacks fixture checks')
        version = report.get('version', '')
        require(re.fullmatch(r'bitbottle version ' + re.escape(tag[1:]) + r' \(' + sha + r'\) built .+', version),
                'native binary version or commit mismatch')
    for platform in TARGETS[:2]:
        target = platform.replace('/', '_')
        report = read_json(evidence / ('candidate-smoke-' + target + '.json'))
        check_report(report, manifest, platform)
        require(report.get('candidate_manifest_sha256') == digest(manifest_path), 'verification manifest mismatch')
        require(report.get('candidate_version') == tag[1:], 'Linux evidence version mismatch')
        require(report.get('read_only_fixture_calls', 0) >= 2, 'Linux evidence lacks fixture checks')
        require(report.get('npm_sha256') == manifest['files'][manifest['npm']], 'npm evidence mismatch')
        require(report.get('oci_index_sha256') == manifest['docker_digest'], 'Docker index evidence mismatch')
        require(report.get('oci_manifest_sha256') == manifest['docker_manifests'][platform],
                'Docker platform evidence mismatch')
        require(bool(report.get('docker_image_id')) and bool(report.get('docker_image')),
                'Docker execution evidence missing')
        require(re.fullmatch(r'[0-9a-f]{64}', report.get('binary_sha256', '')),
                'Linux evidence lacks binary digest')
    return manifest


def check_report(report, manifest, platform):
    require(report.get('target') == platform.replace('/', '_'), 'evidence target mismatch')
    require(report.get('commit_sha') == manifest['commit'], 'evidence commit mismatch')
    require(report.get('archive_sha256') == manifest['archives'][platform]['sha256'],
            'evidence archive mismatch')
    require(type(report.get('mcp_tools')) is int and report['mcp_tools'] > 0,
            'evidence lacks MCP tools')


def publish_plan(root, manifest):
    tag = manifest['tag']
    assets = [str(root / name) for name in manifest['files']
              if name.startswith('dist/') and name not in ('dist/artifacts.json', 'dist/metadata.json')
              and (name.endswith(('.tar.gz', '.zip', '.deb', '.rpm', '.apk')) or name == 'dist/checksums.txt')]
    assets += [str(root / MANIFEST), str(root / 'sbom.spdx.json')]
    commands = [['gh', 'release', 'view', tag, '--repo', REPOSITORY],
                ['gh', 'release', 'upload', tag, '--repo', REPOSITORY, *assets],
                ['gh', 'release', 'edit', tag, '--repo', REPOSITORY, '--notes-file', str(root / 'release-notes.md')],
                ['npm', 'publish', str(root / manifest['npm']), '--access', 'public', '--provenance']]
    for image_tag in (tag, 'latest'):
        commands.append(['skopeo', 'copy', '--all', '--preserve-digests',
                         'oci:' + str(root / 'docker') + ':candidate',
                         'docker://' + REPOSITORY + ':' + image_tag])
    for platform, expected in manifest['docker_manifests'].items():
        arch = platform.split('/')[1]
        for image_tag in (tag + '-' + arch, 'latest-' + arch):
            commands.append(['skopeo', '--override-os', 'linux', '--override-arch', arch,
                             'copy', '--preserve-digests',
                             'oci:' + str(root / 'docker') + ':candidate',
                             'docker://' + REPOSITORY + ':' + image_tag])
    return commands


def gh_assets(tag):
    response = subprocess.run(['gh', 'api', '--method', 'GET',
                               f'repos/{REPOSITORY}/releases/tags/{tag}'],
                              check=True, capture_output=True, text=True)
    assets = json.loads(response.stdout)['assets']
    require(len({item['name'] for item in assets}) == len(assets), 'duplicate GitHub release assets')
    return {item['name']: item.get('digest') for item in assets}


def _structured_npm_error_code(value):
    try:
        payload = json.loads(value)
    except (ValueError, TypeError):
        return None
    if not isinstance(payload, dict):
        return None
    nested = payload.get('error')
    return payload.get('code') or (nested.get('code') if isinstance(nested, dict) else None)


def npm_error_code(stdout, stderr):
    output = '\n'.join((stdout, stderr)).strip()
    for value in (stdout, stderr):
        code = _structured_npm_error_code(value)
        if code:
            return code
    match = re.search(r'(?:^|\n)npm (?:ERR!|error) code (E\d+)(?:\s|$)', output)
    return match.group(1) if match else None


def _normalize_npm_integrity(output):
    if not output.strip():
        return None
    try:
        value = json.loads(output)
    except (ValueError, TypeError) as error:
        raise ValueError('npm registry returned invalid integrity data') from error
    if value is None or value == '':
        return None
    if isinstance(value, dict):
        nested = value.get('dist')
        value = nested.get('integrity') if isinstance(nested, dict) else value.get('integrity')
    if value is None or value == '':
        return None
    require(isinstance(value, str), 'npm registry returned invalid integrity data')
    return value


def npm_integrity(version, timeout=30):
    response = subprocess.run(['npm', 'view', '@proggarapsody/bitbottle@' + version,
                               'dist.integrity', '--json'], capture_output=True, text=True, timeout=timeout)
    if response.returncode:
        output = '\n'.join((response.stdout, response.stderr)).strip()
        code = npm_error_code(response.stdout, response.stderr)
        if code == 'E404' or output == 'E404':
            return None
        detail = f' ({code})' if isinstance(code, str) and re.fullmatch(r'E\d+', code) else ''
        raise ValueError(f'npm registry lookup failed{detail}')
    return _normalize_npm_integrity(response.stdout)


def publish_npm(version, integrity, command, wait_timeout, poll_interval, existing_npm):
    if existing_npm is not None:
        require(existing_npm == integrity, 'existing npm package integrity mismatch')
    else:
        subprocess.run(command, check=True)
    deadline = time.monotonic() + wait_timeout
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise ValueError('timed out waiting for npm package to become visible after publish')
        try:
            visible = npm_integrity(version, timeout=min(30, remaining))
        except subprocess.TimeoutExpired as error:
            if remaining <= 30:
                raise ValueError('timed out waiting for npm package to become visible after publish') from error
            raise ValueError('npm registry lookup timed out') from error
        if visible is not None:
            require(visible == integrity, 'published npm package integrity mismatch')
            return
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise ValueError('timed out waiting for npm package to become visible after publish')
        time.sleep(min(poll_interval, remaining))


def execute_plan(root, manifest, evidence, commands, npm_wait_timeout, npm_poll_interval):
    expected = {Path(name).name: 'sha256:' + checksum for name, checksum in manifest['files'].items()}
    expected[MANIFEST] = 'sha256:' + digest(root / MANIFEST)
    published = gh_assets(manifest['tag'])
    assets = commands[1][6:]
    # A matching previous upload can be reused; an existing mismatch is never overwritten.
    for asset in assets:
        name = Path(asset).name
        require(name not in published or published[name] == expected[name], 'existing GitHub asset digest mismatch')
    npm_path = root / manifest['npm']
    integrity = 'sha512-' + base64.b64encode(hashlib.sha512(npm_path.read_bytes()).digest()).decode()
    try:
        existing_npm = npm_integrity(manifest['version'], timeout=30)
    except subprocess.TimeoutExpired as error:
        raise ValueError('npm registry lookup timed out') from error
    require(existing_npm in (None, integrity), 'existing npm package integrity mismatch')
    for command in commands:
        verify(root, manifest['commit'], manifest['tag'], evidence)
        if command[:3] == ['gh', 'release', 'upload']:
            missing = [asset for asset in assets if Path(asset).name not in published]
            if missing:
                subprocess.run(command[:6] + missing, check=True)
            uploaded = gh_assets(manifest['tag'])
            require(all(uploaded.get(Path(asset).name) == expected[Path(asset).name] for asset in assets),
                    'uploaded GitHub asset digest mismatch')
        elif command[:2] == ['npm', 'publish']:
            publish_npm(manifest['version'], integrity, command, npm_wait_timeout,
                        npm_poll_interval, existing_npm)
        elif command[0] == 'skopeo':
            expected_image = manifest['docker_digest']
            if '--override-arch' in command:
                arch = command[command.index('--override-arch') + 1]
                expected_image = manifest['docker_manifests']['linux/' + arch]
            with tempfile.TemporaryDirectory(prefix='release-publish-') as temporary:
                receipt = Path(temporary) / 'digest'
                operation = command.copy()
                offset = operation.index('copy') + 1
                operation[offset:offset] = ['--digestfile', str(receipt)]
                subprocess.run(operation, check=True)
                require(receipt.read_text().strip() == expected_image, 'published Docker digest mismatch')
        else:
            subprocess.run(command, check=True)
        verify(root, manifest['commit'], manifest['tag'], evidence)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=('seal', 'verify', 'publish'))
    parser.add_argument('--candidate-dir', type=Path, required=True)
    parser.add_argument('--sha', required=True)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--evidence-dir', type=Path)
    parser.add_argument('--npm-wait-timeout', type=float, default=300,
                        help='seconds to wait for the published npm package to appear (default: 300)')
    parser.add_argument('--npm-poll-interval', type=float, default=5,
                        help='seconds between npm registry checks (default: 5)')
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument('--execute', action='store_true', help='perform publication (default: print plan only)')
    mode.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()
    for name, value in (('--npm-wait-timeout', args.npm_wait_timeout),
                        ('--npm-poll-interval', args.npm_poll_interval)):
        require(math.isfinite(value) and value > 0, f'{name} must be a finite positive number')
    root = args.candidate_dir.absolute()
    if args.action == 'seal':
        require(not args.execute, '--execute is only supported for publish')
        manifest = describe(root, args.sha, args.tag)
        destination = root / MANIFEST
        require(not destination.exists(), 'candidate already sealed')
        destination.write_text(json.dumps(manifest, sort_keys=True, indent=2) + '\n')
        print(json.dumps(manifest, sort_keys=True))
        return
    require(args.evidence_dir is not None, '--evidence-dir is required')
    manifest = verify(root, args.sha, args.tag, args.evidence_dir)
    commands = publish_plan(root, manifest) if args.action == 'publish' else []
    if args.execute:
        require(args.action == 'publish', '--execute is only supported for publish')
        execute_plan(root, manifest, args.evidence_dir, commands, args.npm_wait_timeout, args.npm_poll_interval)
    print(json.dumps({'commit': args.sha, 'tag': args.tag, 'docker_digest': manifest['docker_digest'],
                      'dry_run': not args.execute, 'commands': commands}, sort_keys=True))


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, ValueError, OSError, KeyError, TypeError, tarfile.TarError, subprocess.CalledProcessError) as error:
        print(f'candidate: {error}', file=sys.stderr)
        sys.exit(1)
