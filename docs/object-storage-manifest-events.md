# Object Storage, Case Manifest, and NATS Events

This layer upgrades artifact persistence from local filesystem JSON files to S3-compatible object storage such as MinIO.

MinIO's Go SDK is designed for Amazon S3-compatible object storage, and MinIO documents S3 API compatibility for existing S3 applications. NATS uses subject-based publish-subscribe messaging, which fits artifact-created fan-out well. Amazon S3 `PutObject` overwrites an existing object key unless versioning or related controls are enabled, so the case manifest in this phase is best treated as a single-writer object or backed by bucket versioning. citeturn534991search0turn534991search6turn534991search10turn534991search2

## What is added

- S3/MinIO-backed JSON artifact persistence
- `case_manifest.json` object per case
- NATS event on every created artifact ref
- optional backend selection via env

## Suggested env for MinIO

```bash
export ENABLE_REVIEW_ARTIFACTS='true'
export REVIEW_ARTIFACTS_BACKEND='minio'
export REVIEW_ARTIFACTS_S3_ENDPOINT='minio.local:9000'
export REVIEW_ARTIFACTS_S3_ACCESS_KEY='minioadmin'
export REVIEW_ARTIFACTS_S3_SECRET_KEY='minioadmin'
export REVIEW_ARTIFACTS_S3_BUCKET='watchlist-review'
export REVIEW_ARTIFACTS_S3_PREFIX='watchlist-review'
export REVIEW_ARTIFACTS_S3_USE_SSL='false'
```

## Suggested env for NATS artifact events

```bash
export ENABLE_REVIEW_ARTIFACT_EVENTS='true'
export REVIEW_ARTIFACT_EVENTS_NATS_URL='nats://100.67.85.91:4222'
export REVIEW_ARTIFACT_EVENTS_SUBJECT='clawbot.watchlist.review.artifact.created.v1'
```

## Event behavior

Each created artifact publishes an event with:
- tenant id
- case id
- alert id
- correlation id
- decision label
- artifact ref

That lets downstream services subscribe without polling the object store.
