package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	service := PipelineService{errors: InfraiErrorBackend{client: NewInfraiClient(key)}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pipeline/failures", func(w http.ResponseWriter, r *http.Request) {
		var failure PipelineFailure
		if err := json.NewDecoder(r.Body).Decode(&failure); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		result, err := service.RecordFailure(r.Context(), failure)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
	mux.HandleFunc("GET /creator/delivery-errors", func(w http.ResponseWriter, r *http.Request) {
		groups, err := service.CreatorDelivery(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, groups)
	})
	addr := ":8080"
	log.Printf("media error service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": validationErr.Error()})
		return
	}
	var apiErr *InfraiError
	if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
		writeJSON(w, apiErr.Status, map[string]string{"error": apiErr.Code})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "error backend request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
