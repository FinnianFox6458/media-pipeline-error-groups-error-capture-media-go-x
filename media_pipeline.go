package main

import (
	"context"
	"net/http"
)

type PipelineFailure struct {
	AssetID   string         `json:"asset_id"`
	JobID     string         `json:"job_id"`
	Stage     string         `json:"stage"`
	Creator   string         `json:"creator"`
	Message   string         `json:"message"`
	Exception map[string]any `json:"exception"`
}

type CaptureReceipt struct {
	EventID      string `json:"event_id"`
	ErrorGroupID string `json:"error_group_id"`
}

type CaptureResult struct {
	Status       string `json:"status"`
	AssetID      string `json:"asset_id"`
	ErrorGroupID string `json:"error_group_id,omitempty"`
}

type GroupList struct {
	Groups []map[string]any `json:"groups"`
}

type ErrorBackend interface {
	Capture(context.Context, PipelineFailure) (CaptureReceipt, error)
	Groups(context.Context) (GroupList, error)
}

type InfraiErrorBackend struct{ client *InfraiClient }

func (b InfraiErrorBackend) Capture(ctx context.Context, failure PipelineFailure) (CaptureReceipt, error) {
	payload := map[string]any{
		"title":       failure.Stage + " failed",
		"message":     failure.Message,
		"level":       "error",
		"fingerprint": []string{failure.AssetID, failure.Stage},
		"exception":   failure.Exception,
		"context": map[string]string{
			"asset_id": failure.AssetID,
			"job_id":   failure.JobID,
			"stage":    failure.Stage,
			"creator":  failure.Creator,
		},
	}
	var receipt CaptureReceipt
	key := "media-failure/" + failure.JobID + "/" + failure.Stage
	err := b.client.call(ctx, http.MethodPost, "/v1/errors/capture", payload, key, &receipt)
	return receipt, err
}

func (b InfraiErrorBackend) Groups(ctx context.Context) (GroupList, error) {
	var groups GroupList
	err := b.client.call(ctx, http.MethodGet, "/v1/errors/groups", nil, "", &groups)
	return groups, err
}

type PipelineService struct{ errors ErrorBackend }

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func (s PipelineService) RecordFailure(ctx context.Context, failure PipelineFailure) (CaptureResult, error) {
	if failure.AssetID == "" || failure.JobID == "" || failure.Stage == "" || failure.Message == "" || failure.Exception == nil {
		return CaptureResult{}, &ValidationError{Message: "asset_id, job_id, stage, message, and exception are required"}
	}
	receipt, err := s.errors.Capture(ctx, failure)
	if err != nil {
		return CaptureResult{}, err
	}
	return CaptureResult{Status: "grouped", AssetID: failure.AssetID, ErrorGroupID: receipt.ErrorGroupID}, nil
}

func (s PipelineService) CreatorDelivery(ctx context.Context) (GroupList, error) {
	return s.errors.Groups(ctx)
}
