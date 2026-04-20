package artifacts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Store struct {
	client *minio.Client
	bucket string
	prefix string
	now    func() time.Time
}

func NewS3Store(endpoint, accessKey, secretKey, bucket, prefix string, useSSL bool) (*S3Store, error) {
	client, err := minio.New(strings.TrimSpace(endpoint), &minio.Options{
		Creds:  credentials.NewStaticV4(strings.TrimSpace(accessKey), strings.TrimSpace(secretKey), ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("build minio client: %w", err)
	}
	return &S3Store{
		client: client,
		bucket: strings.TrimSpace(bucket),
		prefix: strings.Trim(strings.TrimSpace(prefix), "/"),
		now:    time.Now,
	}, nil
}

func (s *S3Store) WriteJSON(input WriteInput) (ArtifactRef, error) {
	if s == nil || s.client == nil {
		return ArtifactRef{}, fmt.Errorf("s3 store is not configured")
	}
	if s.bucket == "" {
		return ArtifactRef{}, fmt.Errorf("artifact bucket is required")
	}
	if input.Kind == "" {
		return ArtifactRef{}, fmt.Errorf("artifact kind is required")
	}
	if input.Payload == nil {
		return ArtifactRef{}, fmt.Errorf("artifact payload is required")
	}

	now := s.now().UTC()
	artifactID := fmt.Sprintf("art_%d", now.UnixNano())
	relativePath := path.Join(
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		safePathPart(defaultString(input.CaseID, "case-unknown")),
		safePathPart(defaultString(input.AlertID, "alert-unknown")),
		safePathPart(string(input.Kind))+"-"+artifactID+".json",
	)
	objectKey := s.objectKey(relativePath)

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

	_, err = s.client.PutObject(
		context.Background(),
		s.bucket,
		objectKey,
		bytes.NewReader(encoded),
		int64(len(encoded)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("put object: %w", err)
	}

	return ArtifactRef{
		ArtifactID:   artifactID,
		Kind:         input.Kind,
		ContentType:  "application/json",
		RelativePath: relativePath,
		CreatedAt:    now,
	}, nil
}

func (s *S3Store) UpsertCaseManifest(input ManifestInput) (ArtifactRef, error) {
	if s == nil || s.client == nil {
		return ArtifactRef{}, fmt.Errorf("s3 manifest store is not configured")
	}
	now := s.now().UTC()
	relativePath := path.Join(
		"cases",
		safePathPart(defaultString(input.TenantID, "tenant-unknown")),
		safePathPart(defaultString(input.CaseID, "case-unknown")),
		"case_manifest.json",
	)
	objectKey := s.objectKey(relativePath)

	manifest := CaseManifest{
		ManifestID: fmt.Sprintf("manifest_%s", safePathPart(defaultString(input.CaseID, "case-unknown"))),
		TenantID:   input.TenantID,
		CaseID:     input.CaseID,
		UpdatedAt:  now,
	}
	if existing, err := s.readManifest(objectKey); err == nil {
		manifest = existing
		manifest.UpdatedAt = now
	}

	seen := map[string]struct{}{}
	for _, item := range manifest.Artifacts {
		seen[item.ArtifactID] = struct{}{}
	}
	for _, ref := range input.NewArtifacts {
		if _, ok := seen[ref.ArtifactID]; ok {
			continue
		}
		seen[ref.ArtifactID] = struct{}{}
		manifest.Artifacts = append(manifest.Artifacts, ManifestItem{
			ArtifactID:   ref.ArtifactID,
			Kind:         ref.Kind,
			ContentType:  ref.ContentType,
			RelativePath: ref.RelativePath,
			CreatedAt:    ref.CreatedAt,
		})
	}

	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("marshal case manifest: %w", err)
	}
	_, err = s.client.PutObject(
		context.Background(),
		s.bucket,
		objectKey,
		bytes.NewReader(encoded),
		int64(len(encoded)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("put manifest object: %w", err)
	}

	return ArtifactRef{
		ArtifactID:   manifest.ManifestID,
		Kind:         KindCaseManifest,
		ContentType:  "application/json",
		RelativePath: relativePath,
		CreatedAt:    now,
	}, nil
}

func (s *S3Store) readManifest(objectKey string) (CaseManifest, error) {
	_, err := s.client.StatObject(context.Background(), s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return CaseManifest{}, err
	}

	reader, err := s.client.GetObject(context.Background(), s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return CaseManifest{}, err
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		return CaseManifest{}, err
	}

	var manifest CaseManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return CaseManifest{}, err
	}
	return manifest, nil
}

func (s *S3Store) objectKey(relativePath string) string {
	if s.prefix == "" {
		return strings.Trim(relativePath, "/")
	}
	return path.Join(s.prefix, strings.Trim(relativePath, "/"))
}
