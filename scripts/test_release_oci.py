"""Hermetic tests at the OCI inspection and smoke identity boundaries."""
import hashlib
import importlib.util
import json
import os
import subprocess
import sys
from pathlib import Path
import tempfile
import unittest


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def fixture(root):
    (root / 'blobs/sha256').mkdir(parents=True)
    def blob(value, media):
        data = json.dumps(value).encode() if isinstance(value, dict) else value
        checksum = hashlib.sha256(data).hexdigest()
        (root / 'blobs/sha256' / checksum).write_bytes(data)
        return {'mediaType': media, 'size': len(data), 'digest': 'sha256:' + checksum}
    manifests = []
    for arch in ('amd64', 'arm64'):
        config = blob({'os': 'linux', 'architecture': arch}, 'application/vnd.oci.image.config.v1+json')
        layer = blob(b'layer', 'application/vnd.oci.image.layer.v1.tar')
        manifest = blob({'schemaVersion': 2, 'mediaType': 'application/vnd.oci.image.manifest.v1+json',
                         'config': config, 'layers': [layer]}, 'application/vnd.oci.image.manifest.v1+json')
        manifest['platform'] = {'os': 'linux', 'architecture': arch}
        manifests.append(manifest)
    index = blob({'schemaVersion': 2, 'mediaType': 'application/vnd.oci.image.index.v1+json',
                  'manifests': manifests}, 'application/vnd.oci.image.index.v1+json')
    index['annotations'] = {'org.opencontainers.image.ref.name': 'candidate'}
    (root / 'index.json').write_text(json.dumps({'schemaVersion': 2, 'manifests': [index]}))
    (root / 'oci-layout').write_text('{"imageLayoutVersion":"1.0.0"}')
    return manifests


class OCIContract(unittest.TestCase):
    def test_inspection_accepts_complete_two_platform_layout_and_rejects_tampered_blob(self):
        oci = load('release_oci', 'release_oci.py')
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            manifests = fixture(root)
            info = oci.inspect_layout(root)
            self.assertEqual(set(info['manifests']), {'amd64', 'arm64'})
            self.assertEqual(info['manifests']['amd64']['manifest_digest'], manifests[0]['digest'])
            (root / 'blobs/sha256' / manifests[0]['digest'][7:]).write_bytes(b'corrupt')
            with self.assertRaisesRegex(RuntimeError, 'digest|size'):
                oci.inspect_layout(root)


    def test_smoke_identity_rejects_docker_config_and_extra_root_reference(self):
        smoke = load('smoke', 'smoke-release-candidate.py')
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            fixture(root)
            with self.assertRaisesRegex(RuntimeError, 'Docker image ID'):
                smoke.verify_oci_identity(root, 'linux_amd64', 'sha256:' + '0' * 64)
            index = json.loads((root / 'index.json').read_text())
            index['manifests'].append(index['manifests'][0])
            (root / 'index.json').write_text(json.dumps(index))
            with self.assertRaisesRegex(RuntimeError, 'exactly one'):
                smoke.verify_oci_identity(root, 'linux_amd64', 'sha256:' + '0' * 64)

    def test_oci_digest_cannot_escape_blob_directory(self):
        oci = load('oci', 'release_oci.py')
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            fixture(root)
            index = json.loads((root / 'index.json').read_text())
            index['manifests'][0]['digest'] = 'sha256:../../outside'
            (root / 'index.json').write_text(json.dumps(index))
            with self.assertRaisesRegex(RuntimeError, 'invalid OCI digest'):
                oci.inspect_layout(root)


    def test_export_and_import_use_only_local_transports(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            source = root / 'source'
            manifests = fixture(source)
            for arch, descriptor in zip(('amd64', 'arm64'), manifests):
                target = root / arch
                target.mkdir()
                import shutil
                shutil.copytree(source / 'blobs', target / 'blobs')
                (target / 'index.json').write_text(json.dumps({'schemaVersion': 2, 'manifests': [descriptor]}))
            executable = root / 'skopeo'
            executable.write_text("#!/usr/bin/env python3\nimport os,sys,shutil,json\n"
                "from pathlib import Path\n"
                "args=sys.argv[1:]\n"
                "with open(os.environ['OCI_CALLS'],'a') as f: f.write(json.dumps(args)+'\\n')\n"
                "src,dst=args[-2:]\n"
                "if src.startswith('docker-daemon:'):\n"
                " arch=src.rsplit('-',1)[1]\n"
                " shutil.copytree(Path(os.environ['OCI_FIXTURES'])/arch,Path(dst[4:].rsplit(':',1)[0]))\n"
                "elif not(src.startswith('oci:') and dst.startswith('docker-daemon:')): sys.exit(5)\n")
            executable.chmod(0o755)
            calls = root / 'calls'
            env = dict(os.environ, PATH=str(root) + os.pathsep + os.environ['PATH'],
                       OCI_CALLS=str(calls), OCI_FIXTURES=str(root))
            command = [sys.executable, str(Path(__file__).with_name('release_oci.py'))]
            candidate = root / 'candidate'
            exported = subprocess.run(command + ['export', '--candidate-dir', str(candidate), '--tag', 'v1.2.3'],
                                      env=env, capture_output=True, text=True)
            self.assertEqual(exported.returncode, 0, exported.stderr)
            self.assertEqual(set(json.loads(exported.stdout)['manifests']), {'amd64', 'arm64'})
            imported = subprocess.run(command + ['import', '--candidate-dir', str(candidate), '--tag', 'v1.2.3', '--arch', 'arm64'],
                                      env=env, capture_output=True, text=True)
            self.assertEqual(imported.returncode, 0, imported.stderr)
            invocations = [json.loads(line) for line in calls.read_text().splitlines()]
            self.assertEqual(len(invocations), 3)
            self.assertEqual(invocations[-1][:5], ['--override-os', 'linux', '--override-arch', 'arm64', 'copy'])
            self.assertTrue(all(not argument.startswith('docker://') for call in invocations for argument in call))


if __name__ == '__main__':
    unittest.main()
