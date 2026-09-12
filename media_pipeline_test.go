package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordingBackend struct {
	receipt CaptureReceipt
	calls   []PipelineFailure
}

func TestWriteServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{name: "validation error", err: &ValidationError{Message: "missing required field"}, wantStatus: http.StatusBadRequest, wantError: "missing required field"},
		{name: "backend error", err: errors.New("connection failed"), wantStatus: http.StatusBadGateway, wantError: "error backend request failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			writeServiceError(response, tt.err)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			var body map[string]string
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["error"] != tt.wantError {
				t.Fatalf("error = %q, want %q", body["error"], tt.wantError)
			}
		})
	}
}

func (b *recordingBackend) Capture(_ context.Context, failure PipelineFailure) (CaptureReceipt, error) {
	b.calls = append(b.calls, failure)
	return b.receipt, nil
}

func (b *recordingBackend) Groups(context.Context) (GroupList, error) { return GroupList{}, nil }

func TestRecordFailureDecision(t *testing.T) {
	tests := []struct {
		name       string
		failure    PipelineFailure
		wantErr    bool
		wantCalls  int
		wantStatus string
	}{
		{
			name:      "complete processing failure is captured and grouped",
			failure:   PipelineFailure{AssetID: "asset-42", JobID: "job-7", Stage: "transcode", Creator: "creator-9", Message: "encoder exited", Exception: map[string]any{"type": "EncodeError"}},
			wantCalls: 1, wantStatus: "grouped",
		},
		{
			name:    "missing job identity is rejected before capture",
			failure: PipelineFailure{AssetID: "asset-42", Stage: "transcode", Message: "encoder exited", Exception: map[string]any{"type": "EncodeError"}},
			wantErr: true, wantCalls: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &recordingBackend{receipt: CaptureReceipt{ErrorGroupID: "group-3"}}
			result, err := (PipelineService{errors: backend}).RecordFailure(context.Background(), tt.failure)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RecordFailure() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(backend.calls) != tt.wantCalls {
				t.Fatalf("capture calls = %d, want %d", len(backend.calls), tt.wantCalls)
			}
			if result.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", result.Status, tt.wantStatus)
			}
		})
	}
}
