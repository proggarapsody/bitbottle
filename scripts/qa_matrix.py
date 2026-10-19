#!/usr/bin/env python3
"""Discover CLI/MCP operations and assess executed Go test evidence.

`--run` performs one catalog build and one selected go test invocation. `--check`
rechecks saved evidence against the current source and requires full coverage.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile


ROOT = Path(__file__).resolve().parents[1]
PREFIX = "QA_MATRIX "
OUTCOMES = {"success", "error", "unsupported"}
ASSERTIONS = {"output", "effect", "input", "error", "pagination"}
PROVIDERS = {"server", "cloud", "local"}
STATUSES = {"supported", "unsupported", "not_applicable"}
CRITICAL_PROFILE = ROOT / "test/qamatrix/release-critical.json"
# Approved release baseline: policy edits must update this digest in reviewed code.
APPROVED_CRITICAL_PROFILE_SHA256 = "97faa92afa5ff00f145e79c0fc6266d4257d6cfec5c0ef158d2c37dfdd06fd8b"
ALLOWED_CRITICAL_DEFERRALS = {
    ("cli:bitbottle pr merge", "cloud"): {"--auto", "--auto-off", "--rebase"},
    ("cli:bitbottle pr merge", "server"): {"--rebase"},
}
SUPPORTED_MINIMUM = {"success": {"output", "input", "effect"},
                     "error": {"error", "input", "effect"}}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def load_critical_profile(data):
    if sha(data) != APPROVED_CRITICAL_PROFILE_SHA256:
        raise ValueError("critical release baseline differs from approved profile")
    return json.loads(data)


def is_behavior_input(name):
    """Track product, test, fixture, generation, and gate inputs, not dated reports."""
    roots = ("acceptance/", "api/", "git/", "internal/", "pkg/", "cmd/", "test/", "scripts/",
             "skills/", "packages/", "auto-iter/scripts/", ".github/workflows/")
    return (name.startswith(roots) or name in {
        "go.mod", "go.sum", "go.work", "go.work.sum", "Makefile",
        ".golangci.yml", ".golangci.yaml", "Dockerfile",
    } or ("/" not in name and name.endswith((".sh", ".yaml", ".yml", ".json", ".toml"))))


def source_identity():
    paths = subprocess.check_output(
        ["git", "ls-files", "-co", "--exclude-standard", "-z"], cwd=ROOT
    ).split(b"\0")
    selected = sorted({p.decode() for p in paths if p and is_behavior_input(p.decode())})
    h = hashlib.sha256()
    for name in selected:
        path = ROOT / name
        h.update(name.encode() + b"\0" + sha(path.read_bytes()).encode() + b"\n")
    status = subprocess.check_output(
        ["git", "status", "--porcelain", "--untracked-files=normal"], cwd=ROOT,
        text=True,
    )
    head = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
    return {"source_digest": h.hexdigest(), "source_files": len(selected),
            "head": head, "dirty": bool(status.strip())}


def hermetic_go_env(home):
    """Preserve Go build settings while hiding caller Bitbucket credentials."""
    env = {key: value for key, value in os.environ.items()
           if not key.upper().startswith(("BB_", "BITBOTTLE_", "BITBUCKET_", "ATLASSIAN_"))
           and key not in {"SSH_AUTH_SOCK", "GIT_ASKPASS", "SSH_ASKPASS"}}
    go_paths = json.loads(subprocess.check_output(
        ["go", "env", "-json", "GOCACHE", "GOPATH", "GOMODCACHE"], cwd=ROOT,
    ))
    env.update({key: value for key, value in go_paths.items() if value})
    env.update({"HOME": str(home), "XDG_CONFIG_HOME": str(home),
                "XDG_DATA_HOME": str(home)})
    return env


def parse_events(lines):
    """A record is accepted only after its exact Go package/test pass."""
    records, passed_tests, passed_packages = [], set(), set()
    failed, skipped, malformed = [], [], []
    raw_skips, package_output, test_output = [], {}, {}
    for line in lines:
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            malformed.append("invalid go test JSON")
            continue
        package, test, action = event.get("Package"), event.get("Test"), event.get("Action")
        if action == "pass":
            (passed_tests if test else passed_packages).add((package, test) if test else package)
        elif action == "fail":
            failed.append(f"{package}/{test or '<package>'}")
        elif action == "skip":
            raw_skips.append((package, test))
        output = event.get("Output", "")
        if action == "output":
            if test:
                test_output.setdefault((package, test), []).append(output)
            else:
                package_output.setdefault(package, []).append(output)
        if PREFIX not in output:
            continue
        if action != "output" or not package or not test:
            malformed.append("record without Go output/package/test identity")
            continue
        payload = output.split(PREFIX, 1)[1].strip()
        try:
            record = json.loads(payload)
            if not isinstance(record, dict):
                raise ValueError("not object")
        except (json.JSONDecodeError, ValueError) as exc:
            malformed.append(f"{package}/{test}: invalid record: {exc}")
            continue
        records.append({"package": package, "test": test, "scenario": record})
    for record in records:
        record["passed"] = ((record["package"], record["test"]) in passed_tests
                            and record["package"] in passed_packages)
    no_test_packages, live_acceptance_skipped = [], []
    for package, test in raw_skips:
        if not test and any("[no test files]" in out for out in package_output.get(package, [])):
            no_test_packages.append(package)
        elif (test == "TestAcceptance" and package.endswith("/acceptance") and any(
            "acceptance tests skipped: set BITBOTTLE_E2E=1 to run" in out
            for out in test_output.get((package, test), [])
        )):
            live_acceptance_skipped.append(f"{package}/{test}")
        else:
            skipped.append(f"{package}/{test or '<package>'}")
    return {"records": records, "failed": failed, "skipped": skipped,
            "passed_tests": sorted({f"{package}/{test}" for package, test in passed_tests
                                    if package in passed_packages}),
            "no_test_packages": no_test_packages,
            "live_acceptance_skipped": live_acceptance_skipped,
            "live_acceptance_passed": any(test == "TestAcceptance" and package.endswith("/acceptance")
                                          for package, test in passed_tests),
            "malformed": malformed, "passed_packages": sorted(passed_packages)}


def validate_record(record, operations):
    s = record["scenario"]
    problems = []
    if not isinstance(s, dict):
        return ["scenario payload must be an object"]
    op, backend, outcome = (s.get(k) for k in ("operation", "backend", "outcome"))
    if op not in operations:
        problems.append(f"unregistered operation {op!r}")
    else:
        known = known_inputs(operations[op])
        if isinstance(s.get("flags", []), list):
            problems.extend(f"input {name!r} absent from catalog" for name in s.get("flags", [])
                            if isinstance(name, str) and name not in known)
        if isinstance(s.get("not_applicable", {}), dict):
            problems.extend(f"not_applicable input {name!r} absent from catalog"
                            for name in s.get("not_applicable", {}) if name not in known)
    if backend not in PROVIDERS:
        problems.append(f"invalid provider {backend!r}")
    if outcome not in OUTCOMES:
        problems.append(f"invalid outcome {outcome!r}")
    assertions, flags, na = (s.get(k, default) for k, default in (
        ("assertions", []), ("flags", []), ("not_applicable", {})))
    if not isinstance(assertions, list) or not assertions or any(
        not isinstance(a, str) or a not in ASSERTIONS for a in assertions
    ):
        problems.append("invalid assertions")
    if not isinstance(flags, list) or any(not isinstance(f, str) or not f for f in flags):
        problems.append("invalid flags")
    if not isinstance(na, dict) or any(
        not isinstance(k, str) or not isinstance(v, str) or not v.strip()
        for k, v in na.items()
    ):
        problems.append("invalid not_applicable reasons")
    if not record["passed"]:
        problems.append("declared without matching passed Go test and package")
    return problems


def known_inputs(entry):
    values = set(entry.get("parameters", []))
    values |= {x.strip("[]<>") for x in values}
    for flag in entry.get("flags", []):
        values.add("--" + flag["name"])
        if flag.get("shorthand"):
            values.add("-" + flag["shorthand"])
    return values


def assess(catalog, contracts, events):
    cli = [x for x in catalog["cli"] if x["runnable"]]
    mcp = catalog["mcp"]
    operations = {x["id"]: x for x in cli + mcp}
    issues = list(events["malformed"])
    if catalog.get("schema") != 1 or contracts.get("schema") != 1:
        issues.append("unsupported catalog or contract schema")
    if not catalog.get("global_version", {}).get("observed"):
        issues.append("global --version was not observed from app.Run")
    if len(operations) != len(cli) + len(mcp):
        issues.append("duplicate operation IDs in catalog")
    reviewed = contracts.get("operations", {})
    if not isinstance(reviewed, dict):
        reviewed = {}
        issues.append("contracts.operations must be an object")
    for op in reviewed:
        if op not in operations:
            issues.append(f"contract for unregistered operation {op}")
    grouped = {}
    valid_records = []
    for r in events["records"]:
        errors = validate_record(r, operations)
        if errors:
            issues.extend(f"{r['package']}/{r['test']}: {e}" for e in errors)
            continue
        valid_records.append(r)
        s = r["scenario"]
        grouped.setdefault((s["operation"], s["backend"], s["outcome"]), []).append(s)
    if not valid_records:
        issues.append("no passed QA scenario record")
    rows = []
    for op, entry in sorted(operations.items()):
        contract = reviewed.get(op, {})
        if not isinstance(contract, dict):
            issues.append(f"{op}: contract must be an object")
            contract = {}
        providers = contract.get("providers", {})
        scope = contract.get("scope", "remote")
        if scope not in ("remote", "local"):
            issues.append(f"{op}: invalid scope {scope!r}")
        required = contract.get("required_assertions", {})
        important_flags = contract.get("important_flags", [])
        important_parameters = contract.get("important_parameters", [])
        if not isinstance(important_flags, list) or not isinstance(important_parameters, list):
            issues.append(f"{op}: important inputs must be arrays")
            important_flags, important_parameters = [], []
        important = important_flags + important_parameters
        if any(not isinstance(x, str) or not x for x in important):
            issues.append(f"{op}: invalid important input")
            important = []
        unknown_inputs = set(important) - known_inputs(entry)
        if unknown_inputs:
            issues.append(f"{op}: important inputs absent from catalog: {sorted(unknown_inputs)}")
        if not isinstance(required, dict):
            issues.append(f"{op}: required_assertions must be an object")
            required = {}
        if not isinstance(providers, dict):
            issues.append(f"{op}: providers must be an object")
            providers = {}
        if scope == "local" and any(
            not isinstance(providers.get(backend), dict) or
            providers[backend].get("status") != expected
            for backend, expected in (("server", "not_applicable"),
                                      ("cloud", "not_applicable"),
                                      ("local", "supported"))
        ):
            issues.append(f"{op}: local scope requires server/cloud not_applicable and local supported")
        for backend in ("server", "cloud", "local"):
            if backend == "local" and backend not in providers:
                continue
            p = providers.get(backend, {})
            status = p.get("status", "unknown") if isinstance(p, dict) else "unknown"
            reasons = []
            if status not in STATUSES:
                reasons.append("provider support unreviewed")
            if status in ("unsupported", "not_applicable"):
                reason = p.get("reason") if isinstance(p, dict) else None
                if not isinstance(reason, str) or not reason.strip():
                    reasons.append("provider classification lacks reason")
            present_outcomes = {key[2] for key in grouped if key[:2] == (op, backend)}
            if status == "supported":
                outcomes = ("success", "error")
            elif status == "unsupported":
                outcomes = ("unsupported",)
            else:
                outcomes = ()
            if present_outcomes - set(outcomes):
                reasons.append(f"scenario contradicts provider status: {sorted(present_outcomes - set(outcomes))}")
            for outcome in outcomes:
                if outcome == "unsupported":
                    needed = ["error", "effect"]
                else:
                    needed = required.get(outcome)
                    if not isinstance(needed, list) or not needed or any(
                        not isinstance(a, str) or a not in ASSERTIONS for a in needed
                    ):
                        reasons.append(f"{outcome}: missing/invalid required assertions")
                        continue
                    if status == "supported" and not SUPPORTED_MINIMUM[outcome].issubset(needed):
                        reasons.append(f"{outcome}: required assertions omit supported minimum")
                candidates = grouped.get((op, backend, outcome), [])
                if not candidates:
                    reasons.append(f"{outcome}: no passed scenario")
                    continue
                if not any(set(needed).issubset(s["assertions"]) for s in candidates):
                    reasons.append(f"{outcome}: no single passed scenario asserts {needed}")
            if status == "supported":
                supported_scenarios = [s for outcome in ("success", "error")
                                       for s in grouped.get((op, backend, outcome), [])]
                for input_name in important:
                    if not any(input_name in s.get("flags", []) for s in supported_scenarios):
                        reasons.append(f"important input {input_name} untested")
            if status == "not_applicable":
                reasons.append("applicability lacks public outcome proof")
            rows.append({"operation": op, "backend": backend, "status": status,
                         "verified": status in ("supported", "unsupported") and not reasons,
                         "reasons": sorted(set(reasons))})
        operation_rows = [row for row in rows if row["operation"] == op]
        by_backend = {row["backend"]: row for row in operation_rows}
        if (scope == "local" and all(
            by_backend.get(remote, {}).get("status") == "not_applicable" and
            by_backend[remote]["reasons"] == ["applicability lacks public outcome proof"] and
            isinstance(providers.get(remote), dict) and
            isinstance(providers[remote].get("reason"), str) and
            providers[remote]["reason"].strip()
            for remote in ("server", "cloud")
        ) and by_backend.get("local", {}).get("status") == "supported" and
                by_backend["local"]["verified"]):
            for remote in ("server", "cloud"):
                by_backend[remote]["reasons"] = []
                by_backend[remote]["verified"] = True
        if operation_rows and all(row["status"] == "not_applicable" for row in operation_rows):
            issues.append(f"{op}: every provider is not_applicable; no operation was exercised")
    if not events["passed_packages"]:
        issues.append("no Go test package passed")
    if events["failed"]:
        issues.append("Go test failure")
    if events["skipped"]:
        issues.append("Go test skip")
    if events["live_acceptance_passed"]:
        issues.append("live acceptance unexpectedly passed in hermetic lane")
    missing = [r for r in rows if not r["verified"]]
    reviewed_rows = [r for r in rows if r["status"] in STATUSES]
    verified_status = lambda status: sum(r["verified"] and r["status"] == status for r in rows)
    declaration_unknown = sum("declared_meta_backends" not in t for t in mcp)
    return {"schema": 1, "complete": not missing and not issues,
            "counts": {"cli_nodes": len(catalog["cli"]), "cli_runnable": len(cli),
                       "mcp_tools": len(mcp), "operations": len(operations),
                       "provider_pairs": len(rows), "reviewed_pairs": len(reviewed_rows),
                       "verified_pairs": sum(r["verified"] for r in rows),
                       "verified_supported_pairs": verified_status("supported"),
                       "verified_unsupported_pairs": verified_status("unsupported"),
                       "verified_not_applicable_pairs": verified_status("not_applicable"),
                       "verified_local_pairs": sum(r["verified"] and r["backend"] == "local" for r in rows),
                       "unverified_pairs": len(missing),
                       "mcp_declaration_unknown": declaration_unknown,
                       "passed_scenario_records": len(valid_records),
                       "failed_tests": len(events["failed"]),
                       "skipped_tests": len(events["skipped"]),
                       "no_test_packages": len(events["no_test_packages"]),
                       "live_acceptance_skipped": len(events["live_acceptance_skipped"])},
            "issues": issues, "pairs": rows,
            "live_acceptance": {"status": "not_run" if events["live_acceptance_skipped"] else
                                ("unexpected_pass" if events["live_acceptance_passed"] else "not_observed"),
                                "skipped_tests": events["live_acceptance_skipped"]},
            "reviewed_pilot": {"reviewed_pairs": len(reviewed_rows),
                               "verified_pairs": sum(r["verified"] for r in reviewed_rows)}}


def assess_critical(report, catalog, contracts, profile, events=None):
    """Require a frozen, reviewed release subset without changing full DoD rows."""
    issues, selected, deferred = [], [], []
    operations = {x["id"]: x for x in catalog["cli"] if x["runnable"]}
    operations.update({x["id"]: x for x in catalog["mcp"]})
    reviewed = contracts.get("operations", {})
    chosen = profile.get("operations", {}) if isinstance(profile, dict) else {}
    if not isinstance(profile, dict) or profile.get("schema") != 1 or not isinstance(chosen, dict) or not chosen:
        issues.append("critical profile schema or operation selection invalid")
        chosen = {}
    if set(chosen) != set(reviewed):
        issues.append("critical operation selection changed from reviewed contracts: "
                      f"missing={sorted(set(reviewed) - set(chosen))}, "
                      f"unknown={sorted(set(chosen) - set(reviewed))}")
    required_tests = profile.get("required_tests", []) if isinstance(profile, dict) else []
    if (not isinstance(required_tests, list) or not required_tests or
            any(not isinstance(x, str) or "/Test" not in x for x in required_tests) or
            len(required_tests) != len(set(required_tests))):
        issues.append("critical required_tests invalid")
        required_tests = []
    passed_tests = set(events.get("passed_tests", [])) if events is not None else set()
    if events is not None:
        issues.extend(f"required public test did not pass: {name}" for name in required_tests
                      if name not in passed_tests)
    rows = {(r["operation"], r["backend"]): r for r in report["pairs"]}
    for op, selection in sorted(chosen.items()):
        contract, entry = reviewed.get(op), operations.get(op)
        if not isinstance(selection, dict) or not isinstance(contract, dict) or entry is None:
            issues.append(f"{op}: critical operation absent or invalid in catalog/contracts")
            continue
        for key in ("important_flags", "important_parameters"):
            values = selection.get(key)
            if (not isinstance(values, list) or
                    any(not isinstance(x, str) or not x for x in values) or
                    len(values) != len(set(values)) or values != contract.get(key, [])):
                issues.append(f"{op}: important inputs changed ({key})")
        provider_selection = selection.get("providers")
        providers = contract.get("providers", {})
        if not isinstance(provider_selection, dict) or set(provider_selection) != set(providers):
            issues.append(f"{op}: provider selection changed")
            continue
        for backend, chosen_provider in sorted(provider_selection.items()):
            contract_provider = providers[backend]
            if (not isinstance(chosen_provider, dict) or
                    not isinstance(contract_provider, dict) or
                    chosen_provider.get("status") != contract_provider.get("status") or
                    chosen_provider.get("status") not in STATUSES):
                issues.append(f"{op}/{backend}: provider status changed or unreviewed")
                continue
            exemptions = chosen_provider.get("deferred_inputs")
            if not isinstance(exemptions, dict):
                issues.append(f"{op}/{backend}: deferred_inputs must be an object")
                continue
            allowed = ALLOWED_CRITICAL_DEFERRALS.get((op, backend), set())
            important_flags = selection.get("important_flags", [])
            if not isinstance(important_flags, list) or any(not isinstance(x, str) for x in important_flags):
                important_flags = []
            if (set(exemptions) - allowed or
                    set(exemptions) - set(important_flags) or
                    (exemptions and chosen_provider["status"] != "supported") or
                    any(not isinstance(reason, str) or not reason.strip()
                        for reason in exemptions.values())):
                issues.append(f"{op}/{backend}: unapproved or unexplained deferred input")
            for name, reason in sorted(exemptions.items()):
                deferred.append({"operation": op, "backend": backend,
                                 "input": name, "reason": reason})
            row = rows.get((op, backend))
            if row is None:
                issues.append(f"{op}/{backend}: selected provider pair missing")
                continue
            remaining = [reason for reason in row["reasons"]
                         if reason not in {f"important input {name} untested" for name in exemptions}]
            verified = row["status"] == chosen_provider["status"] and not remaining
            selected.append({"operation": op, "backend": backend, "status": row["status"],
                             "verified": verified, "reasons": remaining,
                             "deferred_inputs": sorted(exemptions)})
            if not verified:
                issues.extend(f"{op}/{backend}: {reason}" for reason in remaining or
                              ["provider status mismatch"])
    if report["issues"]:
        issues.extend(report["issues"])
    if not report.get("scope_complete"):
        issues.append("full package scope ./... was not tested")
    return {"passed": not issues, "selected_pairs": len(selected),
            "selected_verified_pairs": sum(row["verified"] for row in selected),
            "full_incomplete_pairs": report["counts"]["unverified_pairs"],
            "required_tests": required_tests,
            "passed_required_tests": sorted(set(required_tests) & passed_tests),
            "deferred_inputs": deferred, "pairs": selected, "issues": issues}


def markdown(report):
    c = report["counts"]
    lines = ["# QA matrix", "", f"Complete: **{'yes' if report['complete'] else 'NO'}**", "",
             f"Full Go package scope executed: {'yes' if report.get('scope_complete') else 'no'}.", "",
             f"Catalog: {c['cli_nodes']} CLI nodes ({c['cli_runnable']} runnable), "
             f"{c['mcp_tools']} MCP tools; {c['operations']} runnable operations.", "",
             f"Full DoD: {c['verified_pairs']}/{c['provider_pairs']} provider pairs verified; "
             f"{c['unverified_pairs']} incomplete.",
             f"Verified outcomes: {c['verified_supported_pairs']} supported, "
             f"{c['verified_unsupported_pairs']} unsupported, "
             f"{c['verified_not_applicable_pairs']} locally proven not applicable "
             f"({c['verified_local_pairs']} verified local pairs).",
             f"Reviewed pilot: {report['reviewed_pilot']['verified_pairs']}/"
             f"{report['reviewed_pilot']['reviewed_pairs']} reviewed pairs verified.",
             f"MCP provider declaration absent: {c['mcp_declaration_unknown']}. "
             "Declaration absence is unknown; an explicit reviewed contract may classify support.", "",
             f"Live acceptance: {report['live_acceptance']['status']} "
             "(separate from hermetic QA coverage).", ""]
    critical = report.get("release_critical")
    if critical:
        lines += ["## Release critical profile", "",
                  f"Passed: **{'yes' if critical['passed'] else 'NO'}**; "
                  f"{critical['selected_verified_pairs']}/{critical['selected_pairs']} "
                  "selected provider pairs verified with stated deferrals.",
                  f"Full matrix still has {critical['full_incomplete_pairs']} incomplete pairs.",
                  "DC не проверен live (hermetic tests only).", "",
                  "Deferred important inputs:"]
        lines += [f"- {x['operation']} / {x['backend']} / {x['input']}: {x['reason']}"
                  for x in critical["deferred_inputs"]] or ["- None"]
        lines += ["", "Critical issues:"]
        lines += [f"- {x}" for x in critical["issues"][:30]] or ["- None"]
        lines += [""]
    lines += ["## Blocking issues", ""]
    lines += [f"- {x}" for x in report["issues"][:30]] or ["- None"]
    if len(report["issues"]) > 30:
        lines.append(f"- … {len(report['issues']) - 30} more in report.json")
    lines += ["", "## Incomplete provider pairs", ""]
    lines += [f"- {r['operation']} / {r['backend']}: {', '.join(r['reasons'])}"
              for r in report["pairs"] if not r["verified"]][:50] or ["- None"]
    if c["unverified_pairs"] > 50:
        lines.append(f"- … {c['unverified_pairs'] - 50} more in report.json")
    return "\n".join(lines) + "\n"


def save_json(path, data):
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")


def run(output, packages, coverprofile=None):
    output.mkdir(parents=True, exist_ok=True)
    before = source_identity()
    catalog_bytes = subprocess.check_output(["go", "run", "./cmd/qa-catalog"], cwd=ROOT)
    catalog = json.loads(catalog_bytes)
    contract_bytes = (ROOT / "test/qamatrix/contracts.json").read_bytes()
    contracts = json.loads(contract_bytes)
    command = ["go", "test", "-json", "-race", "-shuffle=on", "-count=1"]
    if coverprofile is not None:
        command.append("-coverprofile=" + str(coverprofile.resolve()))
    command.extend(packages)
    with tempfile.TemporaryDirectory(prefix="bitbottle-qa-home-") as home:
        with (output / "test-events.jsonl").open("wb") as log:
            result = subprocess.run(command, cwd=ROOT, env=hermetic_go_env(home),
                                    stdout=log, stderr=subprocess.PIPE)
    event_bytes = (output / "test-events.jsonl").read_bytes()
    events = parse_events(event_bytes.decode("utf-8", "replace").splitlines())
    after = source_identity()
    report = assess(catalog, contracts, events)
    report["scope_complete"] = "./..." in packages
    if not report["scope_complete"]:
        report["complete"] = False
    if before["source_digest"] != after["source_digest"]:
        report["issues"].append("source changed during catalog/test run")
    if result.returncode != 0:
        report["issues"].append(f"go test exited {result.returncode}: {result.stderr.decode('utf-8', 'replace')[:1000]}")
    if report["issues"]:
        report["complete"] = False
    profile_bytes = CRITICAL_PROFILE.read_bytes()
    report["release_critical"] = assess_critical(
        report, catalog, contracts, load_critical_profile(profile_bytes), events)
    save_json(output / "catalog.json", catalog)
    save_json(output / "scenarios.json", events)
    save_json(output / "report.json", report)
    (output / "report.md").write_text(markdown(report))
    save_json(output / "run.json", {"schema": 1, "command": command,
              "lane": "hermetic", "private_home": True,
              "source": after, "source_before": before,
              "catalog_sha256": sha((output / "catalog.json").read_bytes()),
              "scenarios_sha256": sha((output / "scenarios.json").read_bytes()),
              "events_sha256": sha(event_bytes), "contracts_sha256": sha(contract_bytes),
              "critical_profile_sha256": sha(profile_bytes),
              "report_sha256": sha((output / "report.json").read_bytes()),
              "test_exit_code": result.returncode})
    (output / "run.sha256").write_text(sha((output / "run.json").read_bytes()) + "\n")
    print(f"QA matrix: {report['counts']['verified_pairs']}/{report['counts']['provider_pairs']} "
          f"pairs verified; full DoD {'complete' if report['complete'] else 'incomplete'}; "
          f"report {output / 'report.md'}")
    return report


def check(output, critical=False):
    if sha((output / "run.json").read_bytes()) != (output / "run.sha256").read_text().strip():
        raise ValueError("run manifest SHA mismatch")
    manifest = json.loads((output / "run.json").read_text())
    for name, key in (("catalog.json", "catalog_sha256"),
                      ("scenarios.json", "scenarios_sha256"),
                      ("test-events.jsonl", "events_sha256"),
                      ("report.json", "report_sha256")):
        if sha((output / name).read_bytes()) != manifest[key]:
            raise ValueError(f"saved {name} differs from run manifest")
    if sha((ROOT / "test/qamatrix/contracts.json").read_bytes()) != manifest["contracts_sha256"]:
        raise ValueError("contracts changed since run")
    profile_bytes = CRITICAL_PROFILE.read_bytes()
    if sha(profile_bytes) != manifest["critical_profile_sha256"]:
        raise ValueError("critical profile changed since run")
    if source_identity()["source_digest"] != manifest["source"]["source_digest"]:
        raise ValueError("source changed since run")
    if manifest["source_before"]["source_digest"] != manifest["source"]["source_digest"]:
        raise ValueError("source changed during run")
    if manifest["test_exit_code"] != 0:
        raise ValueError("go test did not exit successfully")
    if manifest.get("lane") != "hermetic" or not manifest.get("private_home"):
        raise ValueError("test run was not in the hermetic lane")
    if "./..." not in manifest["command"]:
        raise ValueError("full package scope ./... was not tested")
    catalog = json.loads((output / "catalog.json").read_text())
    events = parse_events((output / "test-events.jsonl").read_text().splitlines())
    if events != json.loads((output / "scenarios.json").read_text()):
        raise ValueError("saved scenarios do not match Go test event stream")
    contracts = json.loads((ROOT / "test/qamatrix/contracts.json").read_text())
    report = assess(catalog, contracts, events)
    report["scope_complete"] = True
    report["release_critical"] = assess_critical(
        report, catalog, contracts, load_critical_profile(profile_bytes), events)
    if report != json.loads((output / "report.json").read_text()):
        raise ValueError("saved report does not match execution evidence")
    if critical:
        if not report["release_critical"]["passed"]:
            raise ValueError("critical profile incomplete: " +
                             "; ".join(report["release_critical"]["issues"][:5]))
        print("Release critical profile complete and evidence current; full matrix "
              f"has {report['counts']['unverified_pairs']} incomplete pairs")
        return
    if not report["complete"]:
        raise ValueError(f"matrix incomplete: {report['counts']['unverified_pairs']} provider pairs")
    print("QA matrix complete and evidence current")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="store_true")
    parser.add_argument("--check", "--require-complete", action="store_true")
    parser.add_argument("--check-critical", "--require-critical", action="store_true")
    parser.add_argument("--packages", nargs="+", default=["./..."])
    parser.add_argument("--coverprofile", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if not args.run and not args.check and not args.check_critical:
        parser.error("select --run, --check, or --check-critical")
    try:
        if args.run:
            report = run(args.output, args.packages, args.coverprofile)
            if report["issues"]:
                raise ValueError("run contains invalid, failed, skipped, or empty evidence; see report.md")
        if args.check:
            check(args.output)
        if args.check_critical:
            check(args.output, critical=True)
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
        print(f"QA matrix: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
