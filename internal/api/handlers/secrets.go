package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/charlesgreen/gsm/internal/models"
	"github.com/charlesgreen/gsm/internal/storage"
)

// SecretsHandler handles HTTP requests for secret operations.
type SecretsHandler struct {
	storage storage.Storage
}

// NewSecretsHandler creates a new SecretsHandler with the provided storage backend.
func NewSecretsHandler(storage storage.Storage) *SecretsHandler {
	return &SecretsHandler{
		storage: storage,
	}
}

// CreateSecret handles POST requests to create a new secret.
func (h *SecretsHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid project path", "INVALID_ARGUMENT")
		return
	}

	var req models.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	if req.SecretID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "secretId is required", "INVALID_ARGUMENT")
		return
	}

	secret := models.NewSecret(projectID, req.SecretID, req.Secret.Labels)
	if req.Secret.Replication != nil {
		secret.Replication = *req.Secret.Replication
	}

	if err := h.storage.CreateSecret(r.Context(), projectID, req.SecretID, secret); err != nil {
		if err == storage.ErrSecretExists {
			message := models.FormatResourceExistsError("secret", projectID, req.SecretID)
			writeErrorResponse(w, http.StatusConflict, message, "ALREADY_EXISTS")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create secret", "INTERNAL")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(secret)
}

// GetSecret handles GET requests to retrieve a secret by ID.
func (h *SecretsHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	projectID, secretID := extractProjectAndSecretID(r.URL.Path)
	if projectID == "" || secretID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid secret path", "INVALID_ARGUMENT")
		return
	}

	secret, err := h.storage.GetSecret(r.Context(), projectID, secretID)
	if err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get secret", "INTERNAL")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(secret)
}

// ListSecrets handles GET requests to list all secrets in a project.
func (h *SecretsHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid project path", "INVALID_ARGUMENT")
		return
	}

	pageSize := 100
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	pageToken := r.URL.Query().Get("pageToken")

	secrets, nextPageToken, err := h.storage.ListSecrets(r.Context(), projectID, pageSize, pageToken)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to list secrets", "INTERNAL")
		return
	}

	response := models.ListSecretsResponse{
		Secrets:       secrets,
		NextPageToken: nextPageToken,
		TotalSize:     len(secrets),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// DeleteSecret handles DELETE requests to remove a secret.
func (h *SecretsHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	projectID, secretID := extractProjectAndSecretID(r.URL.Path)
	if projectID == "" || secretID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid secret path", "INVALID_ARGUMENT")
		return
	}

	if err := h.storage.DeleteSecret(r.Context(), projectID, secretID); err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete secret", "INTERNAL")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractProjectID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractProjectAndSecretID(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var projectID, secretID string
	
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			projectID = parts[i+1]
		}
		if part == "secrets" && i+1 < len(parts) {
			secretID = parts[i+1]
		}
	}
	
	return projectID, secretID
}

func extractProjectSecretAndVersionID(path string) (string, string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var projectID, secretID, versionID string
	
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			projectID = parts[i+1]
		}
		if part == "secrets" && i+1 < len(parts) {
			secretID = parts[i+1]
		}
		if part == "versions" && i+1 < len(parts) {
			versionID = parts[i+1]
		}
	}
	
	return projectID, secretID, versionID
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := models.NewErrorResponse(statusCode, message, status)
	_ = json.NewEncoder(w).Encode(errorResp)
}

