#!/bin/sh
set -eu

curl --request POST http://localhost:8080/pipeline/failures \
  --header 'Content-Type: application/json' \
  --data '{"asset_id":"asset-42","job_id":"job-7","stage":"transcode","creator":"creator-9","message":"encoder exited","exception":{"type":"EncodeError","stack":"worker/transcode"}}'

curl --request GET http://localhost:8080/creator/delivery-errors
