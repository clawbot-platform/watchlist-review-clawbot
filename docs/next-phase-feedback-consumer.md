# Feedback Consumer

This service subscribes to `clawbot.watchlist.review.feedback.created.v1` and writes normalized feedback rows into a Postgres tuning/eval store.

## Required env

```bash
export TUNING_EVAL_STORE_DSN='postgres://postgres:postgres@localhost:5432/watchlist_review_eval?sslmode=disable'
export REVIEW_FEEDBACK_EVENTS_NATS_URL='nats://127.0.0.1:4222'
export REVIEW_FEEDBACK_EVENTS_SUBJECT='clawbot.watchlist.review.feedback.created.v1'
```

## Run

```bash
go run ./cmd/feedback-consumer
```
