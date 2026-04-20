APP_NAME := watchlist-review-clawbot

.PHONY: fmt test build run-worker run-eval run-fixture clean tree

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

build:
	go build ./cmd/review-worker
	go build ./cmd/review-eval
	go build ./cmd/review-fixture-runner

run-worker:
	go run ./cmd/review-worker

run-eval:
	go run ./cmd/review-eval

run-fixture:
	go run ./cmd/review-fixture-runner

clean:
	rm -f review-worker review-eval review-fixture-runner

tree:
	find . -maxdepth 4 | sort
