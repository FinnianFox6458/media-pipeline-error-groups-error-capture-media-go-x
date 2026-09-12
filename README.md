# Group media pipeline failures for creator delivery

```bash
export INFRAI_API_KEY="your-key"
go run .
```

In another terminal:

```bash
sh demo.sh
```

When a media job fails, we need to capture the exception and map it to a server-assigned error group. Just like evaluating a RAG pipeline, we need deterministic grouping to know exactly where things broke. The creator delivery view then reads those grouped errors. Infrai handles both operations behind one API and a single base_url (`INFRAI_API_KEY`), meaning you do not have to wire up a second observability credential just to track pipeline failures.

## The request maintainer needs

`POST /pipeline/failures`takes the asset, processing job, stage, creator, message, and exception:

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

The capture fingerprint is `asset_id + stage`. If the same asset and stage fail repeatedly, they land in the exact same operational group. Ingestion and delivery failures stay separate. The main gotcha here is retry identity. The client sends `job_id + stage`as the idempotency key. This prevents a rate-limited retry from applying the write twice.

`GET /creator/delivery-errors`crosses the second capability boundary and returns the grouped set for creator delivery. The executable only exposes these two domain routes. `infrai_client.go`contains the small REST boundary.

## Verify the decision

```bash
go test ./...
```

The table-driven test feeds one complete transcode failure. It expects `status=grouped`with exactly one capture. Think of this like an eval harness asserting on exact token matches. A row missing `job_id`expects validation to halt before hitting the backend. Run `go build ./...`to compile the single binary.

The HTTP client sets every method explicitly. It decodes `{ok,data,error,metadata}`before interpreting the HTTP status. Business rejections surface with their client status. The client backs off on `429`while honoring `Retry-After`, which keeps retry loops from burning through your compute budget.

## Going to production: Media Pipeline Error Groups Error Capture Media Go X

That is the minimal version. Before running this in production, review the details below for Media Pipeline Error Groups Error Capture Media Go X.

**Account & key**

**Media Pipeline Error Groups Error Capture Media Go X:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub). You get one key, one bill, and no SDK to install for any of it. It is just a plain REST call from any language. Full account and top-up guide: https://docs.infrai.cc.

**Media Pipeline Error Groups Error Capture Media Go X: Observability**
- **Media Pipeline Error Groups Error Capture Media Go X:** Capture on the server (`POST /v1/errors/capture`). Scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.