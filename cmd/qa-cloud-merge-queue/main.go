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
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()
	report, err := acceptancecloud.RunMergeQueue(ctx, cfg, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "merge queue acceptance failed; inspect report.json:", report.Error)
		os.Exit(1)
	}
	fmt.Println("merge queue acceptance:", report.Status)
}
