// Package handlers provides HTTP request handlers for the GSM emulator API endpoints.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charlesgreen/gsm/internal/models"
)

// Version represents the current version of the GSM emulator.
const Version = "1.0.0"

// HealthHandler handles health check and readiness probe endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a new health handler instance.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health responds with the current health status of the emulator.
func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	response := models.HealthResponse{
		Status:    "OK",
		Timestamp: time.Now().UTC(),
		Version:   Version,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// Ready responds with the readiness status of the emulator.
func (h *HealthHandler) Ready(w http.ResponseWriter, _ *http.Request) {
	response := models.HealthResponse{
		Status:    "READY",
		Timestamp: time.Now().UTC(),
		Version:   Version,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}