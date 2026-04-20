package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/parsers/screeningjson"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/runtime"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: review-worker <alert.json>")
	}

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("read alert file: %v", err)
	}

	alert, err := screeningjson.New().Parse(context.Background(), raw)
	if err != nil {
		log.Fatalf("parse alert: %v", err)
	}

	identityClient := identity.New(
		os.Getenv("CLAWBOT_IDENTITY_BASE_URL"),
		10*time.Second,
		os.Getenv("WATCHLIST_REVIEW_DEFAULT_TENANT"),
	)

	flow := runtime.NewFlow(identityClient, nil, nil)
	result, err := flow.BuildReviewContext(context.Background(), runtime.ReviewInput{
		TenantID:      os.Getenv("WATCHLIST_REVIEW_DEFAULT_TENANT"),
		CaseID:        alert.Metadata.CaseID,
		CorrelationID: "review-worker-local",
		Alert:         alert,
	})
	if err != nil {
		log.Fatalf("build review context: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		log.Fatalf("encode result: %v", err)
	}
}
