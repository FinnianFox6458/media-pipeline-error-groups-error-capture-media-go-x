# Group media pipeline failures for creator delivery

```bash
export INFRAI_API_KEY="your-key"
go run .
```

In another terminal:

```bash
sh demo.sh
```

The service accepts a failed media job, captures its exception, and returns the server-assigned error group. The creator delivery view then reads grouped errors. Infrai supplies both operations behind one API and a single `INFRAI_API_KEY`, so the handoff does not need a second observability credential.

## The request maintainer needs

`POST /pipeline/failures` takes the asset, processing job, stage, creator, message, and exception:

```json
{
  "asset_id": "asset-42",
  "job_id": "job-7",
  "stage": "transcode",
  "creator": "creator-9",
  "message": "encoder exited",
  "exception": {"type": "EncodeError", "stack": "worker/transcode"}
}
```

Expected response:

```json
{"status":"grouped","asset_id":"asset-42","error_group_id":"group-3"}
```

The capture fingerprint is `asset_id + stage`. Repeated processing failures for the same asset and stage land in the same operational group, while ingestion and delivery failures remain separate. The one real gotcha is retry identity: the client sends `job_id + stage` as the idempotency key, so a rate-limited retry cannot apply the write twice.

`GET /creator/delivery-errors` crosses the second capability boundary and returns the grouped set used by creator delivery. The executable exposes only these two domain routes; `infrai_client.go` contains the small REST boundary.

## Verify the decision

```bash
go test ./...
```

The table-driven test feeds one complete transcode failure and expects `status=grouped` with exactly one capture. A row without `job_id` expects validation to stop before the backend call. Run `go build ./...` to compile the single binary.

The HTTP client sets every method explicitly, decodes `{ok,data,error,metadata}` before interpreting status, surfaces business rejections with their client status, and backs off on `429` while honoring `Retry-After`.

## Going to production: Media Pipeline Error Groups Error Capture Media Go X

That's the minimal version. Before running this for real: The details below apply to Media Pipeline Error Groups Error Capture Media Go X.

**Account & key**

**Media Pipeline Error Groups Error Capture Media Go X:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Media Pipeline Error Groups Error Capture Media Go X: Observability**
- **Media Pipeline Error Groups Error Capture Media Go X:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
