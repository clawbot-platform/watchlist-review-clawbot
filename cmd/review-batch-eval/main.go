package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
)

func main() {
	specPath := envOr("REVIEW_BATCH_SPEC_PATH", "eval/specs/review-batch-example.json")
	baseURL := envOr("REVIEW_BATCH_EVAL_BASE_URL", "http://127.0.0.1:8090")
	timeout, err := time.ParseDuration(envOr("REVIEW_BATCH_EVAL_TIMEOUT", "60s"))
	if err != nil {
		log.Fatalf("parse REVIEW_BATCH_EVAL_TIMEOUT: %v", err)
	}

	spec, err := eval.LoadBatchSpec(specPath)
	if err != nil {
		log.Fatalf("load batch spec: %v", err)
	}
	client := eval.NewClient(baseURL, timeout)
	report, err := client.Run(context.Background(), spec)
	if err != nil {
		log.Fatalf("run batch eval: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		log.Fatalf("encode report: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
