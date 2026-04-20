package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	artifactevents "github.com/clawbot-platform/watchlist-review-clawbot/internal/events/artifacts"
	feedbackevents "github.com/clawbot-platform/watchlist-review-clawbot/internal/events/feedback"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

func main() {
	addr := envOr("WATCHLIST_REVIEW_HTTP_ADDR", ":8090")
	identityBaseURL := os.Getenv("CLAWBOT_IDENTITY_BASE_URL")
	tenant := os.Getenv("WATCHLIST_REVIEW_DEFAULT_TENANT")

	identityClient := identity.New(identityBaseURL, 10*time.Second, tenant)
	noteService := buildNoteService()
	artifactService := buildArtifactService()
	retrievalService := buildRetrievalService()
	feedbackService := buildFeedbackService()

	server, err := api.NewServer(identityClient, noteService, artifactService, retrievalService, feedbackService)
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
	generator := notes.NewOllamaGraniteGenerator(baseURL, model, 90*time.Second)
	return notes.NewService(generator)
}

func buildArtifactBackend() (artifacts.Store, artifacts.ManifestWriter) {
	if !parseBool(envOr("ENABLE_REVIEW_ARTIFACTS", "false")) {
		return nil, nil
	}

	backend := strings.TrimSpace(strings.ToLower(envOr("REVIEW_ARTIFACTS_BACKEND", "filesystem")))
	switch backend {
	case "s3", "minio":
		s3Store, err := artifacts.NewS3Store(
			envOr("REVIEW_ARTIFACTS_S3_ENDPOINT", ""),
			envOr("REVIEW_ARTIFACTS_S3_ACCESS_KEY", ""),
			envOr("REVIEW_ARTIFACTS_S3_SECRET_KEY", ""),
			envOr("REVIEW_ARTIFACTS_S3_BUCKET", ""),
			envOr("REVIEW_ARTIFACTS_S3_PREFIX", "watchlist-review"),
			parseBool(envOr("REVIEW_ARTIFACTS_S3_USE_SSL", "false")),
		)
		if err != nil {
			log.Printf("artifacts: failed to build s3 store: %v", err)
			return nil, nil
		}
		return s3Store, s3Store
	default:
		fsStore := artifacts.NewFileSystemStore(envOr("REVIEW_ARTIFACTS_DIR", "./var/artifacts"))
		return fsStore, nil
	}
}

func buildArtifactService() *artifacts.Service {
	store, manifest := buildArtifactBackend()
	if store == nil {
		return nil
	}

	var extras []any
	if manifest != nil {
		extras = append(extras, manifest)
	}
	if parseBool(envOr("ENABLE_REVIEW_ARTIFACT_EVENTS", "false")) {
		publisher, err := artifactevents.NewNATSPublisher(
			envOr("REVIEW_ARTIFACT_EVENTS_NATS_URL", envOr("CLAWBOT_NATS_URL", "")),
			envOr("REVIEW_ARTIFACT_EVENTS_SUBJECT", artifactevents.DefaultSubject),
		)
		if err != nil {
			log.Printf("artifacts: failed to build nats publisher: %v", err)
		} else {
			extras = append(extras, publisher)
		}
	}

	return artifacts.NewService(store, extras...)
}

func buildFeedbackService() *feedback.Service {
	if !parseBool(envOr("ENABLE_REVIEW_FEEDBACK_CAPTURE", "false")) {
		return nil
	}
	store, manifest := buildArtifactBackend()
	if store == nil {
		return nil
	}

	var extras []any
	if manifest != nil {
		extras = append(extras, manifest)
	}
	if parseBool(envOr("ENABLE_REVIEW_FEEDBACK_EVENTS", "false")) {
		publisher, err := feedbackevents.NewNATSPublisher(
			envOr("REVIEW_FEEDBACK_EVENTS_NATS_URL", envOr("CLAWBOT_NATS_URL", "")),
			envOr("REVIEW_FEEDBACK_EVENTS_SUBJECT", feedbackevents.DefaultSubject),
		)
		if err != nil {
			log.Printf("feedback: failed to build nats publisher: %v", err)
		} else {
			extras = append(extras, publisher)
		}
	}

	return feedback.NewService(store, extras...)
}

func buildRetrievalService() *retrieval.Service {
	if !parseBool(envOr("ENABLE_REVIEW_RETRIEVAL", "false")) {
		return nil
	}
	baseURL := envOr("REVIEW_RETRIEVAL_BASE_URL", "")
	timeout, err := time.ParseDuration(envOr("REVIEW_RETRIEVAL_TIMEOUT", "5s"))
	if err != nil {
		log.Printf("retrieval: invalid timeout: %v", err)
		timeout = 5 * time.Second
	}
	return retrieval.NewService(retrieval.NewClient(baseURL, timeout))
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
