package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrievalgateway"
)

func main() {
	addr := envOr("RETRIEVAL_GATEWAY_ADDR", ":8088")
	searcher, err := buildSearcher()
	if err != nil {
		log.Fatalf("build retrieval searcher: %v", err)
	}
	handler := retrievalgateway.NewHandler(searcher)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("retrieval-gateway listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}

func buildSearcher() (retrievalgateway.Searcher, error) {
	backend := strings.TrimSpace(strings.ToLower(envOr("RETRIEVAL_GATEWAY_BACKEND", "clawmem_http")))
	switch backend {
	case "clawmem_http":
		return retrievalgateway.NewClawmemHTTPSearcher(envOr("CLAWMEM_BASE_URL", "")), nil
	case "pgvector":
		embedder := retrievalgateway.NewOllamaEmbedder(
			envOr("RETRIEVAL_EMBED_BASE_URL", "http://127.0.0.1:11434"),
			envOr("RETRIEVAL_EMBED_MODEL", "embeddinggemma"),
			30*time.Second,
		)
		return retrievalgateway.NewPgvectorSearcher(
			envOr("PGVECTOR_DSN", ""),
			envOr("PGVECTOR_TABLE", "rag_documents"),
			embedder,
		)
	case "qdrant":
		embedder := retrievalgateway.NewOllamaEmbedder(
			envOr("RETRIEVAL_EMBED_BASE_URL", "http://127.0.0.1:11434"),
			envOr("RETRIEVAL_EMBED_MODEL", "embeddinggemma"),
			30*time.Second,
		)
		return retrievalgateway.NewQdrantSearcher(
			envOr("QDRANT_BASE_URL", ""),
			envOr("QDRANT_API_KEY", ""),
			envOr("QDRANT_COLLECTION", ""),
			embedder,
			10*time.Second,
		), nil
	default:
		return nil, fmt.Errorf("unsupported retrieval backend %q", backend)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
