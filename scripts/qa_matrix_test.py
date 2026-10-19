import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import qa_matrix as qa


PKG = "example/test"
TEST = "TestPublicBehavior"
OP = "cli:bitbottle pr list"


def event(action, test=TEST, output=""):
    value = {"Action": action, "Package": PKG}
    if test is not None:
        value["Test"] = test
    if output:
        value["Output"] = output
    return json.dumps(value)


def scenario(outcome, operation=OP, backend="server"):
    return {"operation": operation, "backend": backend, "outcome": outcome,
            "assertions": ["output", "input", "effect"] if outcome == "success" else
            (["error", "effect"] if outcome == "unsupported" else ["error", "input", "effect"]),
            "flags": ["--limit"]}


def observed(records, test_action="pass", package_action="pass"):
    lines = [event("output", output="    test.go:12: " + qa.PREFIX + json.dumps(r) + "\n")
             for r in records]
    lines += [event(test_action), event(package_action, test=None)]
    return qa.parse_events(lines)


def catalog():
    return {"schema": 1, "cli": [{"id": OP, "runnable": True,
                                  "parameters": [], "flags": [{"name": "limit"}]}], "mcp": [],
            "global_version": {"id": "cli:bitbottle --version", "observed": "v1"}}


def contracts():
    return {"schema": 1, "operations": {OP: {
        "providers": {"server": {"status": "supported"},
                      "cloud": {"status": "unsupported", "reason": "fixture has no Cloud behavior"}},
        "required_assertions": {"success": ["output", "input", "effect"],
                                "error": ["error", "input", "effect"]},
        "important_flags": ["--limit"], "important_parameters": []}}}


def local_case():
    op = "cli:bitbottle alias set"
    public = {"schema": 1, "cli": [{"id": op, "runnable": True,
              "parameters": ["name", "expansion"], "flags": []}], "mcp": [],
              "global_version": {"observed": "v1"}}
    manifest = {"schema": 1, "operations": {op: {
        "scope": "local",
        "providers": {
            "server": {"status": "not_applicable", "reason": "local config only"},
            "cloud": {"status": "not_applicable", "reason": "local config only"},
            "local": {"status": "supported"}},
        "required_assertions": {"success": ["output", "effect", "input"],
                                "error": ["error", "effect", "input"]},
        "important_flags": [], "important_parameters": ["name", "expansion"]}}}
    success = {"operation": op, "backend": "local", "outcome": "success",
               "assertions": ["output", "effect", "input"], "flags": ["name", "expansion"]}
    error = {"operation": op, "backend": "local", "outcome": "error",
             "assertions": ["error", "effect", "input"], "flags": ["name", "expansion"]}
    return public, manifest, success, error


class EvidenceAssessment(unittest.TestCase):
    def test_tested_local_operation_derives_both_remote_not_applicable_pairs(self):
        public, manifest, success, error = local_case()
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertTrue(report["complete"], report["issues"])
        self.assertEqual(report["counts"]["provider_pairs"], 3)
        self.assertEqual(report["counts"]["verified_local_pairs"], 1)
        self.assertEqual(report["counts"]["verified_not_applicable_pairs"], 2)

    def test_untested_local_operation_does_not_derive_remote_applicability(self):
        public, manifest, _, _ = local_case()
        report = qa.assess(public, manifest, observed([]))
        self.assertFalse(report["complete"])
        self.assertTrue(all(not row["verified"] for row in report["pairs"]))

    def test_one_sided_remote_not_applicable_does_not_inherit_local_proof(self):
        public, manifest, success, error = local_case()
        manifest["operations"][success["operation"]]["providers"]["cloud"] = {"status": "supported"}
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertFalse(report["complete"])
        self.assertFalse(next(r for r in report["pairs"] if r["backend"] == "server")["verified"])

    def test_remote_scope_cannot_use_local_proof_to_claim_not_applicable(self):
        public, manifest, success, error = local_case()
        manifest["operations"][success["operation"]]["scope"] = "remote"
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertFalse(report["complete"])
        self.assertFalse(next(r for r in report["pairs"] if r["backend"] == "server")["verified"])

    def test_local_proof_needs_required_assertions_and_important_inputs(self):
        public, manifest, success, error = local_case()
        success["assertions"] = ["output", "input"]
        success["flags"] = ["name"]
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertFalse(report["complete"])
        self.assertTrue(all(not r["verified"] for r in report["pairs"]))

    def test_local_contract_cannot_weaken_minimum_assertions(self):
        public, manifest, success, error = local_case()
        contract = manifest["operations"][success["operation"]]
        contract["required_assertions"] = {"success": ["output"], "error": ["error"]}
        contract["important_parameters"] = []
        success["assertions"] = ["output"]
        error["assertions"] = ["error"]
        success["flags"] = []
        error["flags"] = []
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertFalse(report["complete"])
        self.assertTrue(all(not row["verified"] for row in report["pairs"]))

    def test_remote_supported_contract_cannot_weaken_minimum_assertions(self):
        manifest = contracts()
        manifest["operations"][OP]["required_assertions"] = {"success": ["output"], "error": ["error"]}
        manifest["operations"][OP]["important_flags"] = []
        report = qa.assess(catalog(), manifest, observed([
            {**scenario("success"), "flags": []},
            {**scenario("error"), "flags": []},
            scenario("unsupported", backend="cloud")]))
        self.assertFalse(report["complete"])
        server = next(row for row in report["pairs"] if row["backend"] == "server")
        self.assertFalse(server["verified"])

    def test_non_string_provider_reasons_are_rejected(self):
        for bad_reason in (None, 17, {"detail": "claimed gap"}):
            with self.subTest(reason=bad_reason):
                public, manifest, success, error = local_case()
                contract = manifest["operations"][success["operation"]]
                contract["providers"]["server"]["reason"] = bad_reason
                report = qa.assess(public, manifest, observed([success, error]))
                self.assertFalse(report["complete"])
                server = next(row for row in report["pairs"] if row["backend"] == "server")
                self.assertIn("provider classification lacks reason", server["reasons"])
                unsupported = contracts()
                unsupported["operations"][OP]["providers"]["cloud"]["reason"] = bad_reason
                remote = qa.assess(catalog(), unsupported, observed([
                    scenario("success"), scenario("error"),
                    scenario("unsupported", backend="cloud")]))
                cloud = next(row for row in remote["pairs"] if row["backend"] == "cloud")
                self.assertFalse(cloud["verified"])

    def test_local_scope_with_supported_remote_provider_is_invalid(self):
        public, manifest, success, error = local_case()
        manifest["operations"][success["operation"]]["providers"]["server"] = {"status": "supported"}
        report = qa.assess(public, manifest, observed([success, error]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("local scope" in issue for issue in report["issues"]))

    def test_local_proof_does_not_erase_remote_scenario_contradiction(self):
        public, manifest, success, error = local_case()
        remote = dict(success, backend="server")
        report = qa.assess(public, manifest, observed([success, error, remote]))
        self.assertFalse(report["complete"])
        server = next(r for r in report["pairs"] if r["backend"] == "server")
        self.assertFalse(server["verified"])
        self.assertTrue(any("contradicts" in reason for reason in server["reasons"]))

    def test_fingerprint_includes_fixture_inputs_but_excludes_dated_reports(self):
        self.assertTrue(qa.is_behavior_input("test/script/testdata/pr_view.txtar"))
        self.assertTrue(qa.is_behavior_input("api/cloud/openapi.yaml"))
        self.assertTrue(qa.is_behavior_input(".github/workflows/test.yml"))
        self.assertFalse(qa.is_behavior_input("docs/quality/2026-10-19-report.md"))

    def test_hermetic_runner_strips_live_opt_in_and_secrets(self):
        with patch.dict(os.environ, {"PATH": "/bin", "BB_E2E_BACKEND": "cloud",
                                     "BITBOTTLE_E2E": "1", "BB_TOKEN": "secret",
                                     "BITBUCKET_TOKEN": "secret"}, clear=True), patch.object(
                                         qa.subprocess, "check_output",
                                         return_value=b'{"GOCACHE":"/cache","GOPATH":"/go","GOMODCACHE":"/modules"}'):
            env = qa.hermetic_go_env("/private")
        self.assertEqual(env["HOME"], "/private")
        self.assertEqual(env["PATH"], "/bin")
        self.assertEqual(env["GOCACHE"], "/cache")
        self.assertFalse(any(key.startswith(("BB_", "BITBOTTLE_", "BITBUCKET_")) for key in env))

    def test_genuine_passed_outcome_pair_is_verified(self):
        report = qa.assess(catalog(), contracts(), observed([
            scenario("success"), scenario("error"), scenario("unsupported", backend="cloud")]))
        self.assertTrue(report["complete"], report["issues"])
        self.assertEqual(report["counts"]["verified_pairs"], 2)

    def test_important_input_can_be_covered_in_success_while_error_checks_error(self):
        failure = scenario("error")
        failure["flags"] = []
        report = qa.assess(catalog(), contracts(), observed([
            scenario("success"), failure, scenario("unsupported", backend="cloud")]))
        self.assertTrue(report["complete"])

    def test_unsupported_provider_needs_error_and_effect_but_no_method_flags(self):
        manifest = contracts()
        manifest["operations"][OP]["providers"]["server"] = {
            "status": "unsupported", "reason": "Bitbucket Cloud only"}
        manifest["operations"][OP]["required_assertions"]["unsupported"] = ["error", "effect", "input"]
        unsupported = scenario("unsupported")
        unsupported["flags"] = []
        unsupported["assertions"] = ["error", "effect"]
        report = qa.assess(catalog(), manifest, observed([
            unsupported, scenario("unsupported", backend="cloud")]))
        self.assertTrue(report["complete"], report["issues"])

    def test_split_unsupported_error_and_effect_cannot_combine_into_proof(self):
        manifest = contracts()
        manifest["operations"][OP]["providers"]["server"] = {
            "status": "unsupported", "reason": "unsupported here"}
        only_error = scenario("unsupported")
        only_error["assertions"] = ["error"]
        only_effect = scenario("unsupported")
        only_effect["assertions"] = ["effect"]
        report = qa.assess(catalog(), manifest, observed([
            only_error, only_effect, scenario("unsupported", backend="cloud")]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("no single passed scenario" in reason
                            for row in report["pairs"] for reason in row["reasons"]))

    def test_split_success_assertions_cannot_combine_into_proof(self):
        manifest = contracts()
        output_only = scenario("success")
        output_only["assertions"] = ["output"]
        input_only = scenario("success")
        input_only["assertions"] = ["input"]
        report = qa.assess(catalog(), manifest, observed([
            output_only, input_only, scenario("error"), scenario("unsupported", backend="cloud")]))
        self.assertFalse(report["complete"])

    def test_not_applicable_claim_does_not_cover_important_input(self):
        success = scenario("success")
        success["flags"] = []
        success["not_applicable"] = {"--limit": "this scenario does not paginate"}
        failure = scenario("error")
        failure["flags"] = []
        report = qa.assess(catalog(), contracts(), observed([success, failure]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("important input --limit untested" in r["reasons"]
                            for r in report["pairs"]))

    def test_one_not_applicable_provider_cannot_be_verified_by_reason(self):
        manifest = contracts()
        manifest["operations"][OP]["providers"]["cloud"] = {
            "status": "not_applicable", "reason": "claimed Cloud gap"}
        report = qa.assess(catalog(), manifest, observed([scenario("success"), scenario("error")]))
        self.assertFalse(report["complete"])
        cloud = next(row for row in report["pairs"] if row["backend"] == "cloud")
        self.assertFalse(cloud["verified"])
        self.assertIn("applicability lacks public outcome proof", cloud["reasons"])

    def test_all_not_applicable_contract_cannot_complete_operation(self):
        manifest = contracts()
        manifest["operations"][OP]["providers"] = {
            "server": {"status": "not_applicable", "reason": "declared local"},
            "cloud": {"status": "not_applicable", "reason": "declared local"}}
        report = qa.assess(catalog(), manifest, observed([]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("every provider is not_applicable" in x for x in report["issues"]))

    def test_record_before_failed_test_cannot_count(self):
        report = qa.assess(catalog(), contracts(), observed(
            [scenario("success"), scenario("error")], test_action="fail", package_action="fail"))
        self.assertFalse(report["complete"])
        self.assertEqual(report["counts"]["passed_scenario_records"], 0)

    def test_record_before_skipped_test_cannot_count(self):
        report = qa.assess(catalog(), contracts(), observed(
            [scenario("success"), scenario("error")], test_action="skip"))
        self.assertFalse(report["complete"])
        self.assertEqual(report["counts"]["passed_scenario_records"], 0)

    def test_go_no_test_files_package_is_reported_separately(self):
        lines = [event("output", test=None, output="? example/test [no test files]\n"),
                 event("skip", test=None)]
        parsed = qa.parse_events(lines)
        self.assertEqual(parsed["no_test_packages"], [PKG])
        self.assertEqual(parsed["skipped"], [])

    def test_live_acceptance_opt_in_skip_is_not_hermetic_coverage_or_failure(self):
        acceptance = "github.com/proggarapsody/bitbottle/acceptance"
        lines = [json.dumps({"Action": "output", "Package": acceptance,
                             "Test": "TestAcceptance", "Output":
                             "acceptance tests skipped: set BITBOTTLE_E2E=1 to run\n"}),
                 json.dumps({"Action": "skip", "Package": acceptance,
                             "Test": "TestAcceptance"})]
        parsed = qa.parse_events(lines)
        self.assertEqual(parsed["skipped"], [])
        self.assertEqual(parsed["live_acceptance_skipped"], [acceptance + "/TestAcceptance"])
        self.assertFalse(parsed["live_acceptance_passed"])

    def test_other_skipped_test_still_blocks(self):
        parsed = qa.parse_events([event("skip")])
        self.assertEqual(parsed["skipped"], [PKG + "/" + TEST])

    def test_one_missing_outcome_keeps_pair_incomplete(self):
        report = qa.assess(catalog(), contracts(), observed([scenario("success")]))
        self.assertFalse(report["complete"])
        server = next(r for r in report["pairs"] if r["backend"] == "server")
        self.assertIn("error: no passed scenario", server["reasons"])

    def test_unregistered_operation_is_rejected(self):
        report = qa.assess(catalog(), contracts(), observed([
            scenario("success", operation="mcp:missing"), scenario("error")]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("unregistered operation" in x for x in report["issues"]))

    def test_claimed_flag_not_in_public_catalog_is_rejected(self):
        wrong = scenario("success")
        wrong["flags"] = ["--invented"]
        report = qa.assess(catalog(), contracts(), observed([wrong, scenario("error")]))
        self.assertFalse(report["complete"])
        self.assertTrue(any("absent from catalog" in x for x in report["issues"]))

    def test_unknown_provider_review_is_not_green(self):
        manifest = contracts()
        del manifest["operations"][OP]["providers"]["server"]
        report = qa.assess(catalog(), manifest, observed([scenario("success"), scenario("error")]))
        self.assertFalse(report["complete"])
        self.assertTrue(any(r["status"] == "unknown" for r in report["pairs"]))

    def test_declaration_without_actual_package_pass_is_not_evidence(self):
        lines = [event("output", output=qa.PREFIX + json.dumps(scenario("success"))), event("pass")]
        report = qa.assess(catalog(), contracts(), qa.parse_events(lines))
        self.assertEqual(report["counts"]["passed_scenario_records"], 0)
        self.assertFalse(report["complete"])

    def test_stale_source_identity_fails_check(self):
        with tempfile.TemporaryDirectory() as folder:
            output = Path(folder)
            for name in ("catalog.json", "scenarios.json", "test-events.jsonl", "report.json"):
                (output / name).write_text("{}")
            manifest = {"source": {"source_digest": "old"},
                        "source_before": {"source_digest": "old"},
                        "contracts_sha256": qa.sha((qa.ROOT / "test/qamatrix/contracts.json").read_bytes()),
                        "test_exit_code": 0}
            for name, key in (("catalog.json", "catalog_sha256"),
                              ("scenarios.json", "scenarios_sha256"),
                              ("test-events.jsonl", "events_sha256"),
                              ("report.json", "report_sha256")):
                manifest[key] = qa.sha((output / name).read_bytes())
            qa.save_json(output / "run.json", manifest)
            (output / "run.sha256").write_text(qa.sha((output / "run.json").read_bytes()))
            with patch.object(qa, "source_identity", return_value={"source_digest": "new"}):
                with self.assertRaisesRegex(ValueError, "source changed since run"):
                    qa.check(output)


if __name__ == "__main__":
    unittest.main()
