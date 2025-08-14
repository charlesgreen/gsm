// Package models contains data structures and types used throughout the Google Secret Manager emulator.
package models

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"math/rand"
	"strings"
	"time"
)

// ListSecretsResponse represents the response for listing secrets in a project.
type ListSecretsResponse struct {
	Secrets       []*Secret `json:"secrets"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
	TotalSize     int       `json:"totalSize"`
}

// ListSecretVersionsResponse represents the response for listing versions of a secret.
type ListSecretVersionsResponse struct {
	Versions      []*SecretVersion `json:"versions"`
	NextPageToken string           `json:"nextPageToken,omitempty"`
	TotalSize     int              `json:"totalSize"`
}

// CreateSecretRequest represents the request to create a new secret.
type CreateSecretRequest struct {
	SecretID string            `json:"secretId"`
	Secret   *CreateSecretData `json:"secret"`
}

// CreateSecretData contains the secret metadata for creation requests.
type CreateSecretData struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Replication *Replication      `json:"replication,omitempty"`
}

// AddSecretVersionRequest represents the request to add a new version to an existing secret.
type AddSecretVersionRequest struct {
	Payload *SecretPayload `json:"payload"`
}

// ErrorResponse represents an API error response following Google Cloud API standards.
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains the details of an API error with optional extended information.
type ErrorDetail struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Status  string        `json:"status"`
	Errors  []ErrorItem   `json:"errors,omitempty"`
	Details []interface{} `json:"details,omitempty"`
}

// ErrorItem represents individual error details in the errors array.
type ErrorItem struct {
	Domain       string `json:"domain"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	LocationType string `json:"locationType,omitempty"`
	Location     string `json:"location,omitempty"`
}

// ErrorInfo provides detailed error information following AIP-193 standard.
type ErrorInfo struct {
	Type     string            `json:"@type"`
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// NewErrorResponse creates a new error response with the given details.
func NewErrorResponse(code int, message, status string) *ErrorResponse {
	return &ErrorResponse{
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
		},
	}
}

// NewDetailedErrorResponse creates an error response with additional error details.
func NewDetailedErrorResponse(code int, message, status string, errors []ErrorItem) *ErrorResponse {
	return &ErrorResponse{
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
			Errors:  errors,
		},
	}
}

// NewErrorResponseWithInfo creates an error response with ErrorInfo details following AIP-193.
func NewErrorResponseWithInfo(code int, message, status, reason, domain string, metadata map[string]string) *ErrorResponse {
	errorInfo := ErrorInfo{
		Type:     "type.googleapis.com/google.rpc.ErrorInfo",
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	}

	return &ErrorResponse{
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
			Details: []interface{}{errorInfo},
		},
	}
}

// FormatResourceNotFoundError creates a properly formatted "not found" error message.
func FormatResourceNotFoundError(resourceType, projectID, resourceID string) string {
	switch resourceType {
	case "secret":
		return fmt.Sprintf("Secret [projects/%s/secrets/%s] not found.", projectID, resourceID)
	case "version":
		parts := strings.Split(resourceID, "/")
		if len(parts) == 2 {
			return fmt.Sprintf("Secret Version [projects/%s/secrets/%s/versions/%s] not found.", projectID, parts[0], parts[1])
		}
		return fmt.Sprintf("Secret Version [projects/%s/secrets/%s] not found.", projectID, resourceID)
	default:
		return fmt.Sprintf("Resource [projects/%s/%s] not found.", projectID, resourceID)
	}
}

// FormatResourceExistsError creates a properly formatted "already exists" error message.
func FormatResourceExistsError(resourceType, projectID, resourceID string) string {
	switch resourceType {
	case "secret":
		return fmt.Sprintf("Secret [projects/%s/secrets/%s] already exists.", projectID, resourceID)
	default:
		return fmt.Sprintf("Resource [projects/%s/%s] already exists.", projectID, resourceID)
	}
}

// FormatPermissionDeniedError creates a properly formatted permission denied error message.
func FormatPermissionDeniedError(permission, resourcePath string) string {
	return fmt.Sprintf("Permission '%s' denied on resource '%s'.", permission, resourcePath)
}

func generateEtag() string {
	return fmt.Sprintf(`"%x"`, md5.Sum([]byte(fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63()))))
}

func generateChecksum(data []byte) *SecretVersionChecksum {
	crc32Hash := crc32.ChecksumIEEE(data)
	sha256Hash := sha256.Sum256(data)
	
	return &SecretVersionChecksum{
		Crc32c: fmt.Sprintf("%08x", crc32Hash),
		Sha256: fmt.Sprintf("%x", sha256Hash),
	}
}