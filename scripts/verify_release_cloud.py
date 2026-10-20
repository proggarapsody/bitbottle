#!/usr/bin/env python3
"""Require trusted, cleaned-up Cloud acceptance for the exact release commit."""
import argparse
import hashlib
import io
import json
import re
import subprocess
import sys
import zipfile

WORKFLOW = '.github/workflows/nightly-e2e.yml'
LIMIT = 1024 * 1024
INSTRUCTION = 'Run Cloud acceptance on main for this SHA, then rerun release verification'


def require(condition, message):
    if not condition:
        raise ValueError(message)


def api(gh, endpoint, raw=False):
    result = subprocess.run([gh, 'api', '--method', 'GET', endpoint],
                            capture_output=True, check=False, timeout=60)
    require(result.returncode == 0, 'GitHub evidence API request failed')
    require(len(result.stdout) <= LIMIT, 'GitHub evidence response exceeds size limit')
    return result.stdout if raw else json.loads(result.stdout)


def trusted(run, sha, repo):
    return (run.get('head_sha') == sha and run.get('head_branch') == 'main'
            and run.get('head_repository', {}).get('full_name') == repo
            and run.get('path') == WORKFLOW and run.get('event') in ('schedule', 'workflow_dispatch')
            and run.get('status') == 'completed' and run.get('conclusion') == 'success'
            and type(run.get('id')) is int and type(run.get('run_attempt')) is int
            and run['id'] > 0 and run['run_attempt'] > 0)


def check_jobs(jobs):
    for name in ('Validate trusted workflow ref', 'Cloud acceptance'):
        matching = [job for job in jobs if job.get('name') == name]
        require(len(matching) == 1 and matching[0].get('status') == 'completed'
                and matching[0].get('conclusion') == 'success', 'trusted Cloud job did not pass')
        if name == 'Cloud acceptance':
            for step_name in ('Run guarded Cloud acceptance', 'Always clean Cloud fixtures',
                              'Check Cloud evidence files', 'Upload sanitized Cloud evidence'):
                steps = [step for step in matching[0].get('steps', []) if step.get('name') == step_name]
                require(len(steps) == 1 and steps[0].get('status') == 'completed'
                        and steps[0].get('conclusion') == 'success', 'Cloud live/cleanup/evidence step did not pass')


def check_bundle(data, sha):
    with zipfile.ZipFile(io.BytesIO(data)) as archive:
        entries = archive.infolist()
        require(len(entries) == 2 and {entry.filename for entry in entries} == {'report.json', 'manifest.json'},
                'Cloud evidence must contain exactly report.json and manifest.json')
        require(sum(entry.file_size for entry in entries) <= LIMIT,
                'uncompressed Cloud evidence exceeds size limit')
        require(all(not entry.flag_bits & 1 and (entry.external_attr >> 16) & 0o170000 != 0o120000
                    for entry in entries), 'unsupported Cloud evidence file')
        report = json.loads(archive.read('report.json'))
        manifest = json.loads(archive.read('manifest.json'))
    require(report.get('version') == 1 and report.get('commit_sha') == sha
            and report.get('backend') == 'cloud' and report.get('status') == 'pass',
            'Cloud report does not prove this release commit passed')
    require(type(report.get('run_count')) is int and report['run_count'] >= 1
            and type(report.get('pass_count')) is int and report['pass_count'] == report['run_count']
            and type(report.get('fail_count')) is int and report['fail_count'] == 0
            and type(report.get('skip_count')) is int and report['skip_count'] == 0,
            'Cloud report contains missing, failed or skipped scenarios')
    require(report.get('cleanup', {}).get('status') == 'pass', 'Cloud cleanup was not verified')
    checks = report.get('checks', {})
    require(checks and all(check.get('status') == 'pass' for check in checks.values()),
            'Cloud report contains incomplete checks')
    resource, run_id = manifest.get('resource', {}), manifest.get('run_id', '')
    require(manifest.get('version') == 1 and re.fullmatch(r'[0-9a-f]{16}', run_id)
            and resource.get('workspace') == 'bitbottle-qa-lab'
            and resource.get('marker') == 'bitbottle-qa-run:' + run_id
            and re.fullmatch(r'bbqa-[a-z0-9-]{1,58}', resource.get('slug', ''))
            and resource['slug'].endswith('-' + run_id), 'Cloud manifest lacks valid run ownership')


def verify(gh, repo, sha):
    endpoint = f'repos/{repo}/actions/workflows/nightly-e2e.yml/runs?branch=main&head_sha={sha}&status=completed&per_page=100'
    runs = api(gh, endpoint).get('workflow_runs', [])
    for run in runs:
        if not trusted(run, sha, repo):
            continue
        run_id, attempt = run['id'], run['run_attempt']
        jobs = api(gh, f'repos/{repo}/actions/runs/{run_id}/attempts/{attempt}/jobs?per_page=100')
        try:
            check_jobs(jobs.get('jobs', []))
        except ValueError:
            continue
        artifacts = api(gh, f'repos/{repo}/actions/runs/{run_id}/artifacts?per_page=100')
        name = f'cloud-acceptance-{run_id}-{attempt}'
        matches = [artifact for artifact in artifacts.get('artifacts', []) if artifact.get('name') == name]
        require(len(matches) == 1, 'Cloud current-attempt evidence artifact is missing or ambiguous')
        artifact = matches[0]
        require(artifact.get('expired') is False and type(artifact.get('id')) is int
                and 0 < artifact.get('size_in_bytes', 0) <= LIMIT, 'Cloud artifact is expired or oversized')
        source = artifact.get('workflow_run', {})
        require(source.get('id') == run_id and source.get('head_sha') == sha
                and source.get('head_branch') == 'main', 'Cloud artifact source differs from trusted run')
        data = api(gh, f'repos/{repo}/actions/artifacts/{artifact["id"]}/zip', raw=True)
        checksum = 'sha256:' + hashlib.sha256(data).hexdigest()
        require(artifact.get('digest') == checksum, 'Cloud artifact digest mismatch')
        check_bundle(data, sha)
        return {'commit_sha': sha, 'run_url': f'https://github.com/{repo}/actions/runs/{run_id}',
                'run_attempt': attempt, 'artifact_digest': checksum, 'cleanup': True}
    raise ValueError('no trusted successful Cloud acceptance exists for this commit')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--sha', required=True)
    parser.add_argument('--repo', required=True)
    parser.add_argument('--gh', default='gh')
    args = parser.parse_args()
    require(re.fullmatch(r'[0-9a-f]{40}', args.sha), 'expected full 40-character commit SHA')
    require(re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', args.repo), 'invalid GitHub repository')
    print(json.dumps(verify(args.gh, args.repo, args.sha), sort_keys=True))


if __name__ == '__main__':
    try:
        main()
    except (ValueError, TypeError, KeyError, AttributeError, OSError, subprocess.TimeoutExpired,
            zipfile.BadZipFile, RuntimeError) as error:
        print(f'Cloud release proof: {error}. {INSTRUCTION}', file=sys.stderr)
        sys.exit(1)
