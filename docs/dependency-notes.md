# Dependency Notes

This pack adds code that depends on:

- `github.com/minio/minio-go/v7`
- `github.com/minio/minio-go/v7/pkg/credentials`
- `github.com/nats-io/nats.go`

If your repo does not already have them, run:

```bash
go get github.com/minio/minio-go/v7
go get github.com/nats-io/nats.go
go mod tidy
```
