// Command qa-cloud-auto-merge-probe runs a manual, isolated Cloud diagnostic.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/proggarapsody/bitbottle/internal/acceptancecloud"
)

func main() {
	cfg := acceptancecloud.Config{BootstrapToken: os.Getenv("BB_CLOUD_BOOTSTRAP_TOKEN"), ActorToken: os.Getenv("BB_TOKEN"), Email: os.Getenv("BB_CLOUD_EMAIL"), Workspace: os.Getenv("BB_CLOUD_WORKSPACE"), Project: os.Getenv("BB_CLOUD_PROJECT"), ReportDir: os.Getenv("BB_ACCEPTANCE_REPORT_DIR"), Binary: os.Getenv("BITBOTTLE_BIN")}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	report, err := acceptancecloud.RunAutoMergeProbe(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auto-merge diagnostic failed; inspect probe.json and manifest.json:", report.Error)
		os.Exit(1)
	}
	fmt.Println("auto-merge diagnostic complete:", report.Outcome)
}
