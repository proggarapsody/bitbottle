"""Read-only Cloud gate checks through its public CLI and a fake gh boundary."""
import base64
import hashlib
import io
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
import warnings
import zipfile

SCRIPT = Path(__file__).with_name('verify_release_cloud.py')
SHA = '1234567890abcdef1234567890abcdef12345678'
REPO = 'proggarapsody/bitbottle'

class CloudProofTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.fake = self.root / 'gh'
        self.fake.write_text("""#!/usr/bin/env python3
import base64, json, pathlib, sys
root = pathlib.Path(__file__).parent
args = sys.argv[1:]
if args[:3] != ['api', '--method', 'GET']:
    raise SystemExit('only reads permitted')
with (root / 'calls').open('a') as stream:
    stream.write(args[3] + '\\n')
fixture = json.loads((root / 'fixture.json').read_text())
if args[3] not in fixture:
    raise SystemExit('unexpected endpoint')
sys.stdout.buffer.write(base64.b64decode(fixture[args[3]]))
""")
        self.fake.chmod(0o700)
        self.run = {'id': 42, 'run_attempt': 2, 'head_sha': SHA, 'head_branch': 'main',
                    'head_repository': {'full_name': REPO}, 'path': '.github/workflows/nightly-e2e.yml',
                    'event': 'workflow_dispatch', 'status': 'completed', 'conclusion': 'success'}
        self.jobs = [{'name': 'Validate trusted workflow ref', 'status': 'completed', 'conclusion': 'success'},
                     {'name': 'Cloud acceptance', 'status': 'completed', 'conclusion': 'success',
                      'steps': [{'name': name, 'status': 'completed', 'conclusion': 'success'} for name in
                                ('Run guarded Cloud acceptance', 'Always clean Cloud fixtures',
                                 'Check Cloud evidence files', 'Upload sanitized Cloud evidence')]}]
        self.report = {'version': 1, 'commit_sha': SHA, 'backend': 'cloud', 'status': 'pass',
                       'run_count': 1, 'pass_count': 1, 'fail_count': 0, 'skip_count': 0,
                       'cleanup': {'status': 'pass'}, 'checks': {'read_only_verify': {'status': 'pass'}}}
        self.manifest = {'version': 1, 'run_id': '0123456789abcdef', 'resource': {
                         'workspace': 'bitbottle-qa-lab', 'slug': 'bbqa-proof-0123456789abcdef',
                         'marker': 'bitbottle-qa-run:0123456789abcdef'}}
        self.artifact = {'id': 99, 'name': 'cloud-acceptance-42-2', 'expired': False,
                         'workflow_run': {'id': 42, 'head_sha': SHA, 'head_branch': 'main'}}
        self.extra = None
        self.corrupt = False
        self.duplicate = False

    def call(self):
        output = io.BytesIO()
        with zipfile.ZipFile(output, 'w') as archive:
            archive.writestr('report.json', json.dumps(self.report))
            archive.writestr('manifest.json', json.dumps(self.manifest))
            if self.extra:
                with warnings.catch_warnings():
                    warnings.simplefilter('ignore', UserWarning)
                    archive.writestr(self.extra, '{}')
        payload = output.getvalue()
        artifact = dict(self.artifact, size_in_bytes=len(payload), digest='sha256:' + hashlib.sha256(payload).hexdigest())
        if self.corrupt:
            payload = b'not the published archive'
        endpoints = {
            f'repos/{REPO}/actions/workflows/nightly-e2e.yml/runs?branch=main&head_sha={SHA}&status=completed&per_page=100':
                json.dumps({'workflow_runs': [self.run]}).encode(),
            f'repos/{REPO}/actions/runs/42/attempts/2/jobs?per_page=100': json.dumps({'jobs': self.jobs}).encode(),
            f'repos/{REPO}/actions/runs/42/artifacts?per_page=100': json.dumps({'artifacts': [artifact, artifact] if self.duplicate else [artifact]}).encode(),
            f'repos/{REPO}/actions/artifacts/99/zip': payload}
        (self.root / 'fixture.json').write_text(json.dumps({key: base64.b64encode(value).decode() for key, value in endpoints.items()}))
        return subprocess.run([sys.executable, str(SCRIPT), '--sha', SHA, '--repo', REPO,
                               '--gh', str(self.fake)], capture_output=True, text=True)

    def test_trusted_current_attempt_returns_only_a_sanitized_receipt(self):
        result = self.call()
        self.assertEqual(result.returncode, 0, result.stderr)
        receipt = json.loads(result.stdout)
        self.assertEqual(receipt['run_url'], 'https://github.com/proggarapsody/bitbottle/actions/runs/42')
        self.assertEqual(receipt['commit_sha'], SHA)
        self.assertEqual(receipt['run_attempt'], 2)
        self.assertIs(receipt['cleanup'], True)
        self.assertNotIn('bbqa-', result.stdout)
        self.assertEqual(len((self.root / 'calls').read_text().splitlines()), 4)

    def assert_blocked(self):
        result = self.call()
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, '')
        self.assertIn('Run Cloud acceptance on main for this SHA', result.stderr)

    def test_untrusted_run_identity_never_downloads_artifacts(self):
        original = dict(self.run)
        for field, bad in [('head_sha', '0' * 40), ('head_branch', 'feature/proof'),
                           ('head_repository', {'full_name': 'attacker/bitbottle'}),
                           ('event', 'pull_request'), ('path', '.github/workflows/other.yml')]:
            with self.subTest(field=field):
                self.run = dict(original)
                self.run[field] = bad
                self.assert_blocked()
        calls = (self.root / 'calls').read_text()
        self.assertNotIn('/artifacts', calls)

    def test_cloud_skip_or_missing_cleanup_step_blocks_proof(self):
        original = json.loads(json.dumps(self.jobs))
        self.jobs[1]['conclusion'] = 'skipped'
        self.assert_blocked()
        self.jobs = json.loads(json.dumps(original))
        self.jobs[1]['steps'][1]['conclusion'] = 'failure'
        self.assert_blocked()
        self.jobs = json.loads(json.dumps(original))
        del self.jobs[1]['steps'][1]
        self.assert_blocked()
        self.jobs = json.loads(json.dumps(original))
        self.jobs[0]['conclusion'] = 'failure'
        self.assert_blocked()

    def test_report_identity_skips_and_cleanup_failure_block_proof(self):
        original = json.loads(json.dumps(self.report))
        for field, bad in [('commit_sha', '0' * 40), ('backend', 'dc'), ('status', 'skip'),
                           ('skip_count', 1), ('fail_count', 1), ('pass_count', 0),
                           ('cleanup', {'status': 'fail'})]:
            with self.subTest(field=field):
                self.report = dict(original)
                self.report[field] = bad
                self.assert_blocked()

    def test_artifact_attempt_expiration_and_source_are_bound(self):
        original = json.loads(json.dumps(self.artifact))
        for field, bad in [('name', 'cloud-acceptance-42-1'), ('expired', True),
                           ('workflow_run', {'id': 41, 'head_sha': SHA, 'head_branch': 'main'})]:
            with self.subTest(field=field):
                self.artifact = dict(original)
                self.artifact[field] = bad
                self.assert_blocked()

    def test_zip_integrity_and_exact_safe_members_are_required(self):
        self.corrupt = True
        self.assert_blocked()
        self.corrupt = False
        for name in ('../report.json', 'extra.json', 'report.json'):
            with self.subTest(name=name):
                self.extra = name
                self.assert_blocked()

    def test_cleanup_ownership_must_match_allowed_workspace_and_marker(self):
        self.manifest['resource']['workspace'] = 'production'
        self.assert_blocked()
        self.manifest['resource']['workspace'] = 'bitbottle-qa-lab'
        self.manifest['resource']['marker'] = 'unrelated'
        self.assert_blocked()

    def test_suspicious_duplicate_artifact_and_malformed_counts_are_rejected(self):
        self.duplicate = True
        self.assert_blocked()
        self.duplicate = False
        self.report['fail_count'] = False
        self.assert_blocked()

    def test_missing_proof_fails_with_actionable_instruction(self):
        with tempfile.TemporaryDirectory() as temporary:
            fake = Path(temporary) / 'gh'
            fake.write_text('#!/usr/bin/env python3\nprint(\'{"workflow_runs": []}\')\n')
            fake.chmod(0o700)
            result = subprocess.run([sys.executable, str(SCRIPT), '--sha', SHA, '--repo', REPO,
                                     '--gh', str(fake)], capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(result.stdout, '')
            self.assertIn('Run Cloud acceptance on main for this SHA', result.stderr)

if __name__ == '__main__':
    unittest.main()
