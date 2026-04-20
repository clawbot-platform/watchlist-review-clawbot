package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileSystemStore struct {
	baseDir string
	now     func() time.Time
}

func NewFileSystemStore(baseDir string) *FileSystemStore {
	return &FileSystemStore{
		baseDir: strings.TrimSpace(baseDir),
		now:     time.Now,
	}
}

func (s *FileSystemStore) WriteJSON(input WriteInput) (ArtifactRef, error) {
	if s == nil || strings.TrimSpace(s.baseDir) == "" {
		return ArtifactRef{}, fmt.Errorf("artifact base dir is required")
	}
	if input.Kind == "" {
		return ArtifactRef{}, fmt.Errorf("artifact kind is required")
	}
	if input.Payload == nil {
		return ArtifactRef{}, fmt.Errorf("artifact payload is required")
	}

	now := s.now().UTC()
	artifactID := fmt.Sprintf("art_%d", now.UnixNano())
	caseID := safePathPart(defaultString(input.CaseID, "case-unknown"))
	alertID := safePathPart(defaultString(input.AlertID, "alert-unknown"))
	kind := safePathPart(string(input.Kind))

	relativePath := filepath.Join(
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		caseID,
		alertID,
		kind+"-"+artifactID+".json",
	)
	absPath := filepath.Join(s.baseDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return ArtifactRef{}, fmt.Errorf("create artifact dir: %w", err)
	}

	payload := map[string]any{
		"artifact_id":    artifactID,
		"kind":           input.Kind,
		"tenant_id":      input.TenantID,
		"case_id":        input.CaseID,
		"alert_id":       input.AlertID,
		"correlation_id": input.CorrelationID,
		"created_at":     now,
		"content_type":   "application/json",
		"payload":        input.Payload,
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("marshal artifact payload: %w", err)
	}

	if err := os.WriteFile(absPath, encoded, 0o644); err != nil {
		return ArtifactRef{}, fmt.Errorf("write artifact file: %w", err)
	}

	return ArtifactRef{
		ArtifactID:   artifactID,
		Kind:         input.Kind,
		ContentType:  "application/json",
		RelativePath: relativePath,
		CreatedAt:    now,
	}, nil
}

func safePathPart(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		" ", "-",
		":", "-",
	)
	return replacer.Replace(v)
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}
