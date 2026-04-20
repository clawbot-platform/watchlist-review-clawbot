package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

func main() {
	addr := envOr("WATCHLIST_REVIEW_HTTP_ADDR", ":8090")
	identityBaseURL := os.Getenv("CLAWBOT_IDENTITY_BASE_URL")
	tenant := os.Getenv("WATCHLIST_REVIEW_DEFAULT_TENANT")

	identityClient := identity.New(identityBaseURL, 10*time.Second, tenant)
	noteService := buildNoteService()

	server, err := api.NewServer(identityClient, noteService)
	if err != nil {
		log.Fatalf("build server: %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("review-worker-api listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}

func buildNoteService() *notes.Service {
	if !parseBool(envOr("ENABLE_GRANITE_ANALYST_NOTES", "false")) {
		return nil
	}
	provider := strings.TrimSpace(strings.ToLower(envOr("MODEL_PROVIDER", "ollama")))
	if provider != "ollama" && provider != "local_ollama" {
		return nil
	}

	baseURL := envOr("INFERENCE_BASE_URL", "http://127.0.0.1:11434")
	model := envOr("PRIMARY_MODEL", "granite3.3:8b")
	generator := notes.NewOllamaGraniteGenerator(baseURL, model, 30*time.Second)
	return notes.NewService(generator)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(v string) bool {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
