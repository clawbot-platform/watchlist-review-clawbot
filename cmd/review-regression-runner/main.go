package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/promotion"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/reports"
)

func main() {
	specPath := envOr("REVIEW_BATCH_SPEC_PATH", "eval/goldset/manifest.json")
	baseURL := envOr("REVIEW_BATCH_EVAL_BASE_URL", "http://127.0.0.1:8090")
	timeout, err := time.ParseDuration(envOr("REVIEW_BATCH_EVAL_TIMEOUT", "60s"))
	if err != nil {
		log.Fatalf("parse REVIEW_BATCH_EVAL_TIMEOUT: %v", err)
	}
	reportDir := envOr("REVIEW_REPORTS_DIR", "eval/reports")

	spec, err := eval.LoadBatchSpec(specPath)
	if err != nil {
		log.Fatalf("load batch spec: %v", err)
	}

	client := eval.NewClient(baseURL, timeout)
	report, err := client.Run(context.Background(), spec)
	if err != nil {
		log.Fatalf("run batch eval: %v", err)
	}

	decision := promotion.Evaluate(spec, report, promotion.DefaultConfig())
	paths, err := reports.WriteAll(reportDir, report, decision)
	if err != nil {
		log.Fatalf("write reports: %v", err)
	}

	output := map[string]any{
		"report":    report,
		"promotion": decision,
		"artifacts": paths,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		log.Fatalf("encode output: %v", err)
	}

	if !decision.Promoted {
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
