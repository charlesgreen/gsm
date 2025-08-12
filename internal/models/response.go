// Package models contains data structures and types used throughout the Google Secret Manager emulator.
package models

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"math/rand"
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

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains the details of an API error.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
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