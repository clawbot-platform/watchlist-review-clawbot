package artifacts

import (
	"context"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

type fakeStore struct {
	refs []ArtifactRef
	idx  int
}

func (f *fakeStore) WriteJSON(input WriteInput) (ArtifactRef, error) {
	ref := ArtifactRef{
		ArtifactID:   "art-test-" + string(input.Kind),
		Kind:         input.Kind,
		ContentType:  "application/json",
		RelativePath: string(input.Kind) + ".json",
		CreatedAt:    time.Unix(0, 0).UTC(),
	}
	f.refs = append(f.refs, ref)
	return ref, nil
}

type fakeManifestWriter struct{}

func (f *fakeManifestWriter) UpsertCaseManifest(input ManifestInput) (ArtifactRef, error) {
	return ArtifactRef{
		ArtifactID:   "manifest-case-1",
		Kind:         KindCaseManifest,
		ContentType:  "application/json",
		RelativePath: "cases/tenant-1/case-1/case_manifest.json",
		CreatedAt:    time.Unix(1, 0).UTC(),
	}, nil
}

type fakePublisher struct {
	events []ArtifactCreatedEvent
}

func (f *fakePublisher) PublishArtifactCreated(_ context.Context, event ArtifactCreatedEvent) error {
	f.events = append(f.events, event)
	return nil
}

func TestServicePersistWritesArtifactsManifestAndEvents(t *testing.T) {
	store := &fakeStore{}
	manifest := &fakeManifestWriter{}
	publisher := &fakePublisher{}
	service := NewService(store, manifest, publisher)

	refs, warnings := service.Persist(context.Background(), PersistInput{
		TenantID:      "tenant-1",
		CaseID:        "case-1",
		AlertID:       "alert-1",
		CorrelationID: "corr-1",
		DecisionLabel: "escalate",
		ReviewContext: map[string]any{"decision_label": "escalate"},
		AnalystNote: &notes.AnalystNote{
			Status: notes.StatusGenerated,
			Note:   "Escalate for analyst review.",
		},
	})

	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v, want none", warnings)
	}
	if len(refs) != 3 {
		t.Fatalf("refs len = %d, want 3", len(refs))
	}
	if len(publisher.events) != 3 {
		t.Fatalf("events len = %d, want 3", len(publisher.events))
	}
}
