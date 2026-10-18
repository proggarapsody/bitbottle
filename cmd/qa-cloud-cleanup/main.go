// Command qa-cloud-cleanup safely removes a manifest-owned Cloud QA repository.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/proggarapsody/bitbottle/internal/acceptancecloud"
)

func main() {
	manifest := flag.String("manifest", "", "path to the persisted Cloud QA manifest")
	report := flag.String("report", "", "optional report.json path to update")
	flag.Parse()
	if *manifest == "" {
		fmt.Fprintln(os.Stderr, "--manifest is required")
		os.Exit(2)
	}
	token := os.Getenv("BB_CLOUD_BOOTSTRAP_TOKEN")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	err := acceptancecloud.Cleanup(ctx, *manifest, token, "")
	if *report != "" {
		if reportErr := updateReport(*report, *manifest, err); reportErr != nil {
			err = errors.Join(err, reportErr)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func updateReport(path, manifestPath string, cleanupErr error) error {
	b, e := os.ReadFile(path)
	var r acceptancecloud.Report
	if os.IsNotExist(e) {
		r = acceptancecloud.Report{Version: 1, CommitSHA: firstNonEmpty(os.Getenv("BB_ACCEPTANCE_SHA"), os.Getenv("GITHUB_SHA")), Backend: "cloud", BackendVersion: "Bitbucket Cloud REST API v2; provider build unavailable", Scenario: acceptancecloud.ScenarioID, ScenarioIDs: []string{acceptancecloud.ScenarioID}, RunCount: 1, FailCount: 1, Status: "infrastructure_failure", IdentitySeparation: "unverified; bootstrap and actor credentials may belong to the same account", Manifest: manifestPath, Duration: "0s"}
	} else if e != nil {
		return e
	} else if e = json.Unmarshal(b, &r); e != nil {
		return e
	}
	if cleanupErr != nil {
		r.Cleanup = acceptancecloud.Check{Status: "fail", Detail: cleanupErr.Error()}
		r.Status = "cleanup_failure"
	} else {
		r.Cleanup = acceptancecloud.Check{Status: "pass"}
	}
	r.PassCount = 0
	r.FailCount = 0
	if r.Status == "pass" {
		r.PassCount = 1
	} else {
		r.FailCount = 1
	}
	b, e = json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
