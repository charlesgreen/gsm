package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/charlesgreen/gsm/internal/models"
	"github.com/charlesgreen/gsm/internal/storage"
)

// VersionsHandler handles HTTP requests for secret version operations.
type VersionsHandler struct {
	storage storage.Storage
}

// NewVersionsHandler creates a new VersionsHandler with the provided storage backend.
func NewVersionsHandler(storage storage.Storage) *VersionsHandler {
	return &VersionsHandler{
		storage: storage,
	}
}

// AddSecretVersion handles POST requests to add a new version to an existing secret.
func (h *VersionsHandler) AddSecretVersion(w http.ResponseWriter, r *http.Request) {
	projectID, secretID := extractProjectAndSecretFromAddVersionPath(r.URL.Path)
	if projectID == "" || secretID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid secret path", "INVALID_ARGUMENT")
		return
	}

	var req models.AddSecretVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	if req.Payload == nil || len(req.Payload.Data) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Payload data is required", "INVALID_ARGUMENT")
		return
	}

	version, err := h.storage.AddSecretVersion(r.Context(), projectID, secretID, req.Payload.Data)
	if err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to add secret version", "INTERNAL")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(version)
}

// AccessSecretVersion handles POST requests to access the data of a specific secret version.
func (h *VersionsHandler) AccessSecretVersion(w http.ResponseWriter, r *http.Request) {
	projectID, secretID, versionID := extractProjectSecretAndVersionFromAccessPath(r.URL.Path)
	if projectID == "" || secretID == "" || versionID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid version path", "INVALID_ARGUMENT")
		return
	}

	data, err := h.storage.AccessSecretVersion(r.Context(), projectID, secretID, versionID)
	if err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		if err == storage.ErrVersionNotFound {
			message := models.FormatResourceNotFoundError("version", projectID, secretID+"/"+versionID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to access secret version", "INTERNAL")
		return
	}

	version, err := h.storage.GetSecretVersion(r.Context(), projectID, secretID, versionID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get version metadata", "INTERNAL")
		return
	}

	response := models.AccessSecretVersionResponse{
		Name: version.Name,
		Payload: &models.SecretPayload{
			Data:     data,
			Checksum: version.Checksum,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// ListSecretVersions handles GET requests to list all versions of a secret.
func (h *VersionsHandler) ListSecretVersions(w http.ResponseWriter, r *http.Request) {
	projectID, secretID := extractProjectAndSecretID(r.URL.Path)
	if projectID == "" || secretID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid secret path", "INVALID_ARGUMENT")
		return
	}

	pageSize := 100
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	pageToken := r.URL.Query().Get("pageToken")

	versions, nextPageToken, err := h.storage.ListSecretVersions(r.Context(), projectID, secretID, pageSize, pageToken)
	if err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to list secret versions", "INTERNAL")
		return
	}

	response := models.ListSecretVersionsResponse{
		Versions:      versions,
		NextPageToken: nextPageToken,
		TotalSize:     len(versions),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// DeleteSecretVersion handles DELETE requests to remove a specific secret version.
func (h *VersionsHandler) DeleteSecretVersion(w http.ResponseWriter, r *http.Request) {
	projectID, secretID, versionID := extractProjectSecretAndVersionID(r.URL.Path)
	if projectID == "" || secretID == "" || versionID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid version path", "INVALID_ARGUMENT")
		return
	}

	if err := h.storage.DeleteSecretVersion(r.Context(), projectID, secretID, versionID); err != nil {
		if err == storage.ErrSecretNotFound {
			message := models.FormatResourceNotFoundError("secret", projectID, secretID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		if err == storage.ErrVersionNotFound {
			message := models.FormatResourceNotFoundError("version", projectID, secretID+"/"+versionID)
			writeErrorResponse(w, http.StatusNotFound, message, "NOT_FOUND")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete secret version", "INTERNAL")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractProjectAndSecretFromAddVersionPath(path string) (string, string) {
	path = strings.TrimSuffix(path, ":addVersion")
	return extractProjectAndSecretID(path)
}

func extractProjectSecretAndVersionFromAccessPath(path string) (string, string, string) {
	path = strings.TrimSuffix(path, ":access")
	return extractProjectSecretAndVersionID(path)
}