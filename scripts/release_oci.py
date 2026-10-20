#!/usr/bin/env python3
"""Seal local Docker images into an OCI layout, or load a sealed platform locally.

Skopeo owns image transport/conversion. This helper never contacts a registry.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

INDEX = 'application/vnd.oci.image.index.v1+json'
MANIFEST = 'application/vnd.oci.image.manifest.v1+json'
REPOSITORY = 'proggarapsody/bitbottle'


def read_blob(layout, descriptor):
    digest = descriptor.get('digest', '')
    if not re.fullmatch(r'sha256:[a-f0-9]{64}', digest):
        raise RuntimeError('invalid OCI digest')
    path = layout / 'blobs' / 'sha256' / digest[7:]
    if path.is_symlink() or not path.is_file() or layout.resolve() not in path.resolve().parents:
        raise RuntimeError('OCI blob must be a regular file inside the layout')
    data = path.read_bytes()
    if len(data) != descriptor.get('size') or hashlib.sha256(data).hexdigest() != digest[7:]:
        raise RuntimeError('OCI blob digest or size mismatch')
    return data


def inspect_layout(layout):
    """Validate all reachable bytes and return stable index/platform identities."""
    layout = Path(layout)
    if json.loads((layout / 'oci-layout').read_text()) != {'imageLayoutVersion': '1.0.0'}:
        raise RuntimeError('unsupported OCI layout version')
    root = json.loads((layout / 'index.json').read_text())
    refs = root.get('manifests', [])
    if root.get('schemaVersion') != 2 or len(refs) != 1 or refs[0].get('mediaType') != INDEX:
        raise RuntimeError('OCI root must reference exactly one combined index')
    ref = refs[0]
    if ref.get('annotations', {}).get('org.opencontainers.image.ref.name') != 'candidate':
        raise RuntimeError('OCI root must name the candidate reference')
    index = json.loads(read_blob(layout, ref))
    if index.get('schemaVersion') != 2 or index.get('mediaType') != INDEX:
        raise RuntimeError('invalid OCI combined index')
    results = {}
    for descriptor in index.get('manifests', []):
        platform = descriptor.get('platform', {})
        arch = platform.get('architecture')
        if platform.get('os') != 'linux' or arch not in ('amd64', 'arm64') or arch in results:
            raise RuntimeError('OCI index requires unique linux amd64 and arm64 platforms')
        if descriptor.get('mediaType') != MANIFEST:
            raise RuntimeError('OCI platform descriptor must be an image manifest')
        manifest = json.loads(read_blob(layout, descriptor))
        if manifest.get('schemaVersion') != 2 or manifest.get('mediaType') != MANIFEST:
            raise RuntimeError('invalid OCI image manifest')
        config = json.loads(read_blob(layout, manifest['config']))
        if config.get('os') != 'linux' or config.get('architecture') != arch:
            raise RuntimeError('OCI config platform differs from index platform')
        for layer in manifest['layers']:
            read_blob(layout, layer)
        results[arch] = {'manifest_digest': descriptor['digest'],
                         'config_digest': manifest['config']['digest']}
    if set(results) != {'amd64', 'arm64'}:
        raise RuntimeError('OCI index requires both linux amd64 and arm64')
    return {'index_digest': ref['digest'], 'manifests': results}


def store_blob(layout, value):
    data = json.dumps(value, separators=(',', ':'), sort_keys=True).encode()
    checksum = hashlib.sha256(data).hexdigest()
    (layout / 'blobs/sha256' / checksum).write_bytes(data)
    return {'mediaType': INDEX, 'digest': 'sha256:' + checksum, 'size': len(data)}


def export(candidate_dir, tag):
    destination = candidate_dir / 'docker'
    if destination.exists():
        raise RuntimeError('refusing to overwrite existing sealed OCI layout')
    candidate_dir.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='.oci-', dir=candidate_dir) as scratch:
        scratch = Path(scratch)
        combined = scratch / 'combined'
        (combined / 'blobs/sha256').mkdir(parents=True)
        manifests = []
        for arch in ('amd64', 'arm64'):
            source = scratch / arch
            subprocess.run(['skopeo', 'copy', '--format', 'oci',
                            f'docker-daemon:{REPOSITORY}:{tag}-{arch}', f'oci:{source}:candidate'], check=True)
            root = json.loads((source / 'index.json').read_text())
            if len(root.get('manifests', [])) != 1:
                raise RuntimeError('exported Docker image must contain one manifest')
            descriptor = root['manifests'][0]
            descriptor.pop('annotations', None)
            descriptor['platform'] = {'os': 'linux', 'architecture': arch}
            manifests.append(descriptor)
            for blob in (source / 'blobs/sha256').iterdir():
                if not blob.is_file() or blob.is_symlink():
                    raise RuntimeError('export produced a non-regular blob')
                shutil.copyfile(blob, combined / 'blobs/sha256' / blob.name)
        ref = store_blob(combined, {'schemaVersion': 2, 'mediaType': INDEX, 'manifests': manifests})
        ref['annotations'] = {'org.opencontainers.image.ref.name': 'candidate'}
        (combined / 'index.json').write_text(json.dumps({'schemaVersion': 2, 'manifests': [ref]}, separators=(',', ':')))
        (combined / 'oci-layout').write_text('{"imageLayoutVersion":"1.0.0"}')
        info = inspect_layout(combined)
        combined.rename(destination)
    return info


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('operation', choices=('export', 'import'))
    parser.add_argument('--candidate-dir', required=True, type=Path)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--arch', choices=('amd64', 'arm64'))
    args = parser.parse_args()
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', args.tag):
        parser.error('--tag must be a release version tag')
    if args.operation == 'export':
        info = export(args.candidate_dir.resolve(), args.tag)
    else:
        if not args.arch:
            parser.error('import requires --arch')
        layout = args.candidate_dir.resolve() / 'docker'
        info = inspect_layout(layout)
        subprocess.run(['skopeo', '--override-os', 'linux', '--override-arch', args.arch,
                        'copy', f'oci:{layout}:candidate',
                        f'docker-daemon:{REPOSITORY}:{args.tag}-{args.arch}'], check=True)
    print(json.dumps(info))


if __name__ == '__main__':
    main()
