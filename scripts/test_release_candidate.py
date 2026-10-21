"""Hermetic public-CLI checks for immutable release promotion."""
import hashlib
import base64
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import tarfile
import unittest

SCRIPT = Path(__file__).with_name('release_candidate.py')
SHA = '1234567890abcdef1234567890abcdef12345678'

class CandidateTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / 'candidate'
        self.evidence = Path(self.temp.name) / 'evidence'
        for sub in ('dist', 'npm', 'docker/blobs/sha256'):
            (self.root / sub).mkdir(parents=True)
        self.evidence.mkdir()
        self.write('dist/metadata.json', {'version': '1.2.3', 'commit': SHA})
        self.write('dist/artifacts.json', [])
        self.write('sbom.spdx.json', {'spdxVersion': 'SPDX-2.3', 'SPDXID': 'SPDXRef-DOCUMENT', 'packages': []})
        (self.root / 'release-notes.md').write_text('Release 1.2.3\nDC не проверен live\n')
        # SHA256("abc") is the published NIST/FIPS worked example, not production code output.
        self.archive_digest = 'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad'
        self.targets = ('linux_amd64', 'linux_arm64', 'darwin_amd64', 'darwin_arm64', 'windows_amd64')
        names = ['bitbottle_' + t + ('.zip' if t.startswith('windows') else '.tar.gz') for t in self.targets]
        names += ['bitbottle_' + arch + '.' + ext for arch in ('amd64', 'arm64') for ext in ('deb', 'rpm', 'apk')]
        for name in names:
            (self.root / 'dist' / name).write_bytes(b'abc')
        (self.root / 'dist/checksums.txt').write_text(''.join(self.archive_digest + '  ' + name + '\n' for name in names))
        with tarfile.open(self.root / 'npm/package.tgz', 'w:gz') as tar:
            data = b'{"name":"@proggarapsody/bitbottle","version":"1.2.3"}'
            info = tarfile.TarInfo('package/package.json')
            info.size = len(data)
            tar.addfile(info, io.BytesIO(data))
        self.write('docker/oci-layout', {'imageLayoutVersion': '1.0.0'})
        children = []
        for arch in ('amd64', 'arm64'):
            config = self.blob(json.dumps({'os': 'linux', 'architecture': arch}).encode(), 'application/vnd.oci.image.config.v1+json')
            child = self.blob(json.dumps({'schemaVersion': 2, 'mediaType': 'application/vnd.oci.image.manifest.v1+json', 'config': config, 'layers': []}).encode(), 'application/vnd.oci.image.manifest.v1+json')
            child['platform'] = {'os': 'linux', 'architecture': arch}
            children.append(child)
        combined = self.blob(json.dumps({'schemaVersion': 2, 'mediaType': 'application/vnd.oci.image.index.v1+json', 'manifests': children}).encode(), 'application/vnd.oci.image.index.v1+json')
        combined['annotations'] = {'org.opencontainers.image.ref.name': 'candidate'}
        self.write('docker/index.json', {'schemaVersion': 2, 'manifests': [combined]})
        for target in self.targets:
            self.report('native-smoke-' + target, {'commit_sha': SHA, 'target': target,
                        'archive_sha256': self.archive_digest, 'mcp_tools': 281, 'fixture_calls': 1,
                        'version': 'bitbottle version 1.2.3 (' + SHA + ') built 2026-10-21'})
        npm_digest = hashlib.sha256((self.root / 'npm/package.tgz').read_bytes()).hexdigest()
        for target in self.targets[:2]:
            self.report('candidate-smoke-' + target, {'commit_sha': SHA, 'target': target,
                        'archive_sha256': self.archive_digest, 'mcp_tools': 281, 'read_only_fixture_calls': 2,
                        'candidate_version': '1.2.3', 'npm_sha256': npm_digest, 'binary_sha256': self.archive_digest,
                        'docker_image': 'candidate', 'docker_image_id': 'sha256:' + self.archive_digest,
                        'oci_manifest_sha256': children[0 if target.endswith('amd64') else 1]['digest'], 'oci_index_sha256': combined['digest']})

    def write(self, path, value):
        (self.root / path).write_text(json.dumps(value))

    def report(self, name, value):
        (self.evidence / (name + '.json')).write_text(json.dumps(value))

    def blob(self, data, media):
        sha = hashlib.sha256(data).hexdigest()
        (self.root / 'docker/blobs/sha256' / sha).write_bytes(data)
        return {'mediaType': media, 'digest': 'sha256:' + sha, 'size': len(data)}

    def run_cli(self, action, *extra, timeout=None):
        return subprocess.run([sys.executable, str(SCRIPT), action, '--candidate-dir', str(self.root),
                               '--sha', SHA, '--tag', 'v1.2.3', '--evidence-dir', str(self.evidence), *extra],
                              capture_output=True, text=True, env=getattr(self, 'execute_env', None), timeout=timeout)

    def seal(self):
        result = self.run_cli('seal')
        self.assertEqual(result.returncode, 0, result.stderr)
        manifest_digest = hashlib.sha256((self.root / 'candidate-manifest.json').read_bytes()).hexdigest()
        for path in self.evidence.glob('*.json'):
            report = json.loads(path.read_text())
            report['candidate_manifest_sha256'] = manifest_digest
            path.write_text(json.dumps(report))

    def test_verified_plan_uses_existing_bytes_and_all_release_targets(self):
        self.seal()
        result = self.run_cli('publish', '--dry-run')
        self.assertEqual(result.returncode, 0, result.stderr)
        plan = json.loads(result.stdout)
        commands = plan['commands']
        self.assertTrue(plan['dry_run'])
        self.assertEqual(len(commands), 10)
        self.assertEqual(commands[3], ['npm', 'publish', str(self.root / 'npm/package.tgz'), '--access', 'public', '--provenance'])
        self.assertEqual({c[-1] for c in commands if c[0] == 'skopeo'},
                         {'docker://proggarapsody/bitbottle:' + t for t in
                          ('v1.2.3', 'latest', 'v1.2.3-amd64', 'v1.2.3-arm64', 'latest-amd64', 'latest-arm64')})
        self.assertIn(str(self.root / 'dist/bitbottle_arm64.apk'), commands[1])
        self.assertFalse(any('build' in c or 'pack' in c for c in commands))

    def test_mutations_cannot_produce_any_publish_commands(self):
        self.seal()
        mutations = [
            ('dist/bitbottle_linux_amd64.tar.gz', b'replaced'),
            ('npm/package.tgz', b'replaced'),
            ('docker/index.json', b'{}'),
            ('release-notes.md', b'no caveat'),
            ('dist/unlisted.zip', b'extra'),
        ]
        for relative, data in mutations:
            with self.subTest(relative=relative):
                path = self.root / relative
                original = path.read_bytes() if path.exists() else None
                path.write_bytes(data)
                result = self.run_cli('publish')
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, '')
                if original is None:
                    path.unlink()
                else:
                    path.write_bytes(original)

    def test_each_platform_and_linux_container_proof_is_mandatory(self):
        self.seal()
        for path in self.evidence.glob('*.json'):
            with self.subTest(report=path.name):
                original = path.read_text()
                path.unlink()
                result = self.run_cli('publish')
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, '')
                path.write_text(original)

    def test_wrong_evidence_identity_and_incomplete_checks_block_plan(self):
        self.seal()
        path = self.evidence / 'candidate-smoke-linux_amd64.json'
        original = path.read_text()
        for field, bad in [('commit_sha', '0' * 40), ('candidate_version', '1.2.4'),
                           ('npm_sha256', '0' * 64), ('oci_index_sha256', 'sha256:' + '0' * 64),
                           ('read_only_fixture_calls', 0), ('mcp_tools', 0), ('target', 'linux_ppc64')]:
            with self.subTest(field=field):
                report = json.loads(original)
                report[field] = bad
                path.write_text(json.dumps(report))
                result = self.run_cli('publish')
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, '')
        path.write_text(original)

    def test_manifest_path_traversal_and_changed_tag_are_rejected(self):
        self.seal()
        path = self.root / 'candidate-manifest.json'
        manifest = json.loads(path.read_text())
        manifest['files']['../outside'] = self.archive_digest
        path.write_text(json.dumps(manifest))
        self.assertNotEqual(self.run_cli('publish').returncode, 0)
        del manifest['files']['../outside']
        manifest['tag'] = 'v1.2.4'
        path.write_text(json.dumps(manifest))
        self.assertNotEqual(self.run_cli('publish').returncode, 0)

    def test_missing_caveat_and_unsupported_candidate_cannot_be_sealed(self):
        (self.root / 'release-notes.md').write_text('Data Center unverified')
        self.assertNotEqual(self.run_cli('seal').returncode, 0)
        (self.root / 'release-notes.md').write_text('DC не проверен live')
        (self.root / 'outside-link').symlink_to('/tmp')
        self.assertNotEqual(self.run_cli('seal').returncode, 0)

    def test_resealing_changed_notes_cannot_reuse_old_verification(self):
        self.seal()
        notes = self.root / 'release-notes.md'
        notes.write_text('Changed release description\nDC не проверен live\n')
        path = self.root / 'candidate-manifest.json'
        manifest = json.loads(path.read_text())
        manifest['files']['release-notes.md'] = hashlib.sha256(notes.read_bytes()).hexdigest()
        path.write_text(json.dumps(manifest, sort_keys=True, indent=2) + '\n')
        result = self.run_cli('publish')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, '')

    def test_wrong_build_identity_checksum_or_missing_package_cannot_be_sealed(self):
        for metadata in ({'version': '1.2.4', 'commit': SHA},
                         {'version': '1.2.3', 'commit': SHA[:7]}):
            self.write('dist/metadata.json', metadata)
            self.assertNotEqual(self.run_cli('seal').returncode, 0)
        self.write('dist/metadata.json', {'version': '1.2.3', 'commit': SHA})
        package = self.root / 'dist/bitbottle_arm64.rpm'
        package.unlink()
        self.assertNotEqual(self.run_cli('seal').returncode, 0)
        package.write_bytes(b'abc')
        sums = self.root / 'dist/checksums.txt'
        sums.write_text(sums.read_text().replace(self.archive_digest, '0' * 64, 1))
        self.assertNotEqual(self.run_cli('seal').returncode, 0)

    def test_sealed_bundle_rejects_replaced_artifact_symlink(self):
        self.seal()
        archive = self.root / 'dist/bitbottle_linux_amd64.tar.gz'
        archive.unlink()
        outside = Path(self.temp.name) / 'outside'
        outside.write_bytes(b'abc')
        archive.symlink_to(outside)
        result = self.run_cli('publish', '--execute')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, '')

    def fake_publishers(self, corrupt=''):
        directory = Path(self.temp.name) / 'tools'
        directory.mkdir()
        program = directory / 'publisher'
        program.write_text('#!' + sys.executable + """
import base64, hashlib, json, os, pathlib, sys
root = pathlib.Path(os.environ['PUBLISH_FIXTURE'])
state_path = root / 'published.json'
state = json.loads(state_path.read_text()) if state_path.exists() else {'assets': {}, 'npm': os.environ.get('INITIAL_NPM'), 'calls': [], 'npm_views': 0}
tool, args = pathlib.Path(sys.argv[0]).name, sys.argv[1:]
state['calls'].append([tool] + args)
def save():
    state_path.write_text(json.dumps(state))
if tool == 'gh' and args[:1] == ['api']:
    print(json.dumps({'assets': [{'name': name, 'digest': digest} for name, digest in state['assets'].items()]}))
elif tool == 'gh' and args[:2] == ['release', 'upload']:
    for value in args[5:]:
        path = pathlib.Path(value)
        assert path.is_file(), value
        state['assets'][path.name] = 'sha256:' + hashlib.sha256(path.read_bytes()).hexdigest()
    if os.environ.get('CORRUPT_RECEIPT') == 'github-first':
        state['assets'][pathlib.Path(args[5]).name] = 'sha256:' + '0' * 64
    if os.environ.get('CORRUPT_RECEIPT') == 'github':
        state['assets']['checksums.txt'] = 'sha256:' + '0' * 64
elif tool == 'npm' and args[:1] == ['view']:
    state['npm_views'] += 1
    mode = os.environ.get('NPM_VIEW_MODE')
    if mode in ('E401', 'E403', 'registry-error') or (isinstance(mode, str) and mode.startswith('after-') and state['npm'] is not None):
        code = mode.removeprefix('after-') if mode.startswith('after-') else mode
        if code.startswith('E'):
            print(json.dumps({'error': {'code': code, 'summary': 'credential-bearing detail should not escape'}}))
        else:
            print('ECONNREFUSED registry', file=sys.stderr)
        sys.exit(1)
    if mode == 'hang' or (mode == 'hang-after-publish' and state['npm'] is not None):
        import time
        time.sleep(10)
    if mode in ('empty', 'null') and state['npm'] is not None and state['npm_views'] == 2:
        save()
        print('' if mode == 'empty' else 'null')
        sys.exit(0)
    if mode == 'persistent-404' and state['npm'] is not None:
        save()
        print('E404', file=sys.stderr)
        sys.exit(1)
    if state['npm'] is not None and state['npm_views'] <= 2 and os.environ.get('DELAY_NPM_VISIBILITY') == '1':
        save()
        print('E404', file=sys.stderr)
        sys.exit(1)
    if state['npm'] is None:
        save()
        print('E404', file=sys.stderr)
        sys.exit(1)
    print(json.dumps(state['npm']))
elif tool == 'npm' and args[:1] == ['publish']:
    assert '--provenance' in args
    state['npm'] = 'sha512-' + base64.b64encode(hashlib.sha512(pathlib.Path(args[1]).read_bytes()).digest()).decode()
    if os.environ.get('CORRUPT_RECEIPT') == 'npm':
        state['npm'] = 'sha512-invalid'
elif tool == 'skopeo':
    manifest = json.loads((root / 'candidate/candidate-manifest.json').read_text())
    digest = manifest['docker_digest']
    if '--override-arch' in args:
        digest = manifest['docker_manifests']['linux/' + args[args.index('--override-arch') + 1]]
    pathlib.Path(args[args.index('--digestfile') + 1]).write_text(digest)
elif not (tool == 'gh' and args[:2] in (['release', 'view'], ['release', 'edit'])):
    raise SystemExit('unexpected command: ' + repr([tool] + args))
save()
""")
        program.chmod(0o700)
        for tool in ('gh', 'npm', 'skopeo'):
            (directory / tool).symlink_to(program)
        self.execute_env = dict(os.environ, PATH=str(directory) + os.pathsep + os.defpath,
                                PUBLISH_FIXTURE=self.temp.name, CORRUPT_RECEIPT=corrupt)

    def test_execute_waits_for_npm_registry_visibility_after_publish(self):
        self.seal()
        self.fake_publishers()
        self.execute_env['DELAY_NPM_VISIBILITY'] = '1'
        result = self.run_cli('publish', '--execute', '--npm-wait-timeout', '0.25',
                              '--npm-poll-interval', '0.01')
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertEqual(len([call for call in calls if call[:2] == ['npm', 'publish']]), 1)

    def test_execute_reuses_matching_npm_package_without_republishing(self):
        self.seal()
        self.fake_publishers()
        npm_digest = hashlib.sha512((self.root / 'npm/package.tgz').read_bytes()).digest()
        self.execute_env['INITIAL_NPM'] = 'sha512-' + base64.b64encode(npm_digest).decode()
        result = self.run_cli('publish', '--execute')
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[:2] == ['npm', 'publish'] for call in calls))

    def test_execute_rejects_existing_npm_mismatch_before_publishing(self):
        self.seal()
        self.fake_publishers()
        self.execute_env['INITIAL_NPM'] = 'sha512-wrong'
        result = self.run_cli('publish', '--execute')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('existing npm package integrity mismatch', result.stderr)
        self.assertEqual(result.stdout, '')
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[:2] == ['npm', 'publish'] for call in calls))

    def test_execute_waits_through_empty_and_null_npm_results(self):
        for mode in ('empty', 'null'):
            with self.subTest(mode=mode):
                self.setUp()
                self.addCleanup(self.temp.cleanup)
                self.seal()
                self.fake_publishers()
                self.execute_env['NPM_VIEW_MODE'] = mode
                result = self.run_cli('publish', '--execute', '--npm-wait-timeout', '0.25',
                                      '--npm-poll-interval', '0.01')
                self.assertEqual(result.returncode, 0, result.stderr)
                calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
                self.assertEqual(len([call for call in calls if call[:2] == ['npm', 'publish']]), 1)

    def test_execute_times_out_waiting_for_persistent_npm_404_without_docker(self):
        self.seal()
        self.fake_publishers()
        self.execute_env['NPM_VIEW_MODE'] = 'persistent-404'
        result = self.run_cli('publish', '--execute', '--npm-wait-timeout', '0.05',
                              '--npm-poll-interval', '0.01')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('timed out waiting for npm package', result.stderr)
        self.assertEqual(result.stdout, '')
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[0] == 'skopeo' for call in calls))
        self.assertEqual(len([call for call in calls if call[:2] == ['npm', 'publish']]), 1)

    def test_execute_fails_closed_on_npm_auth_and_registry_errors(self):
        for mode in ('E401', 'E403', 'registry-error'):
            with self.subTest(mode=mode):
                self.setUp()
                self.addCleanup(self.temp.cleanup)
                self.seal()
                self.fake_publishers()
                self.execute_env['NPM_VIEW_MODE'] = mode
                result = self.run_cli('publish', '--execute')
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('npm registry lookup failed', result.stderr)
                self.assertEqual(result.stdout, '')
                calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
                self.assertFalse(any(call[0] == 'skopeo' for call in calls))
                self.assertFalse(any(call[:2] == ['npm', 'publish'] for call in calls))

    def test_execute_fails_closed_on_errors_after_publish(self):
        for mode, diagnostic in (('after-E401', 'E401'), ('after-E403', 'E403'),
                                 ('after-E500', 'E500'), ('after-registry-error', 'lookup failed')):
            with self.subTest(mode=mode):
                self.setUp()
                self.addCleanup(self.temp.cleanup)
                self.seal()
                self.fake_publishers()
                self.execute_env['NPM_VIEW_MODE'] = mode
                result = self.run_cli('publish', '--execute', '--npm-wait-timeout', '0.2',
                                      '--npm-poll-interval', '0.01')
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(diagnostic, result.stderr)
                self.assertNotIn('credential-bearing detail', result.stderr)
                calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
                self.assertEqual(len([call for call in calls if call[:2] == ['npm', 'publish']]), 1)
                self.assertFalse(any(call[0] == 'skopeo' for call in calls))

    def test_execute_bounds_hung_npm_registry_process(self):
        self.seal()
        self.fake_publishers()
        self.execute_env['NPM_VIEW_MODE'] = 'hang-after-publish'
        result = self.run_cli('publish', '--execute', '--npm-wait-timeout', '0.2', timeout=5)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('timed out waiting for npm package', result.stderr)
        self.assertEqual(result.stdout, '')

    def test_npm_wait_options_must_be_finite_and_positive(self):
        for option, value in (('--npm-wait-timeout', '0'), ('--npm-wait-timeout', 'nan'),
                              ('--npm-poll-interval', '-1'), ('--npm-poll-interval', 'inf')):
            with self.subTest(option=option, value=value):
                result = self.run_cli('publish', '--execute', option, value)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(option + ' must be a finite positive number', result.stderr)

    def test_execute_promotes_exact_files_verifies_receipts_and_reuses_existing_uploads(self):
        self.seal()
        self.fake_publishers()
        result = self.run_cli('publish', '--execute')
        self.assertEqual(result.returncode, 0, result.stderr)
        state_path = Path(self.temp.name) / 'published.json'
        state = json.loads(state_path.read_text())
        expected_assets = {'bitbottle_' + target + ('.zip' if target.startswith('windows') else '.tar.gz')
                           for target in self.targets}
        expected_assets |= {'bitbottle_' + arch + '.' + suffix for arch in ('amd64', 'arm64')
                            for suffix in ('deb', 'rpm', 'apk')}
        expected_assets |= {'checksums.txt', 'candidate-manifest.json', 'sbom.spdx.json'}
        self.assertEqual(set(state['assets']), expected_assets)
        self.assertEqual(state['assets']['bitbottle_linux_amd64.tar.gz'], 'sha256:' + self.archive_digest)
        self.assertEqual(len([call for call in state['calls'] if call[0] == 'skopeo']), 6)
        result = self.run_cli('publish', '--execute')
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = json.loads(state_path.read_text())['calls']
        self.assertEqual(len([call for call in calls if call[:3] == ['gh', 'release', 'upload']]), 1)
        self.assertEqual(len([call for call in calls if call[:2] == ['npm', 'publish']]), 1)
        self.assertFalse(any('build' in call or 'pack' in call for call in calls))

    def test_execute_rejects_wrong_github_upload_receipt_before_npm(self):
        self.seal()
        self.fake_publishers('github')
        result = self.run_cli('publish', '--execute')
        self.assertNotEqual(result.returncode, 0)
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[:2] == ['npm', 'publish'] for call in calls))
        self.assertFalse(any(call[0] == 'skopeo' for call in calls))

    def test_execute_verifies_the_first_uploaded_asset_too(self):
        self.seal()
        self.fake_publishers('github-first')
        result = self.run_cli('publish', '--execute')
        self.assertNotEqual(result.returncode, 0)
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[:2] == ['npm', 'publish'] for call in calls))

    def test_execute_rejects_wrong_npm_registry_integrity_before_docker(self):
        self.seal()
        self.fake_publishers('npm')
        result = self.run_cli('publish', '--execute')
        self.assertNotEqual(result.returncode, 0)
        calls = json.loads((Path(self.temp.name) / 'published.json').read_text())['calls']
        self.assertFalse(any(call[0] == 'skopeo' for call in calls))

    def test_missing_bundle_never_produces_a_publish_plan(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = subprocess.run([sys.executable, str(SCRIPT), 'publish', '--candidate-dir', tmp,
                                     '--sha', SHA, '--tag', 'v1.2.3', '--evidence-dir', tmp],
                                    capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn('candidate:', result.stderr)
            self.assertEqual(result.stdout, '')

if __name__ == '__main__':
    unittest.main()
