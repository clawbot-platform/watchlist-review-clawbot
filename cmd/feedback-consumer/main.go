package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/evalstore"
	feedbackevents "github.com/clawbot-platform/watchlist-review-clawbot/internal/events/feedback"
)

func main() {
	dsn := os.Getenv("TUNING_EVAL_STORE_DSN")
	if dsn == "" {
		log.Fatal("TUNING_EVAL_STORE_DSN is required")
	}
	natsURL := envOr("REVIEW_FEEDBACK_EVENTS_NATS_URL", envOr("CLAWBOT_NATS_URL", ""))
	subject := envOr("REVIEW_FEEDBACK_EVENTS_SUBJECT", feedbackevents.DefaultSubject)
	if natsURL == "" {
		log.Fatal("REVIEW_FEEDBACK_EVENTS_NATS_URL or CLAWBOT_NATS_URL is required")
	}

	store, err := evalstore.NewPostgresStore(dsn)
	if err != nil {
		log.Fatalf("build eval store: %v", err)
	}
	defer store.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := store.Ping(ctx); err != nil {
		log.Fatalf("ping eval store: %v", err)
	}
	if err := store.EnsureSchema(ctx); err != nil {
		log.Fatalf("ensure eval schema: %v", err)
	}

	consumer, err := feedbackevents.NewConsumer(natsURL, subject, slog.Default(), store)
	if err != nil {
		log.Fatalf("build feedback consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Run(ctx); err != nil {
		log.Fatalf("run feedback consumer: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
