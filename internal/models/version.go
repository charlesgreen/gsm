package models

import (
	"fmt"
	"time"
)

// SecretVersion represents a version of a secret with its data and metadata.
type SecretVersion struct {
	Name        string                `json:"name"`
	CreateTime  time.Time             `json:"createTime"`
	State       SecretVersionState    `json:"state"`
	Etag        string                `json:"etag"`
	Data        []byte                `json:"-"`
	Checksum    *SecretVersionChecksum `json:"checksum,omitempty"`
}

// SecretVersionState represents the state of a secret version.
type SecretVersionState string

const (
	// StateEnabled indicates the version is enabled and accessible.
	StateEnabled   SecretVersionState = "ENABLED"
	// StateDisabled indicates the version is disabled and cannot be accessed.
	StateDisabled  SecretVersionState = "DISABLED"
	// StateDestroyed indicates the version has been permanently destroyed.
	StateDestroyed SecretVersionState = "DESTROYED"
)

// SecretVersionChecksum contains checksums for verifying secret data integrity.
type SecretVersionChecksum struct {
	Crc32c string `json:"crc32c,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

// AccessSecretVersionResponse represents the response for accessing a secret version.
type AccessSecretVersionResponse struct {
	Name     string                     `json:"name"`
	Payload  *SecretPayload             `json:"payload"`
}

// SecretPayload contains the actual secret data and its checksums.
type SecretPayload struct {
	Data     []byte                     `json:"data"`
	Checksum *SecretVersionChecksum     `json:"checksum,omitempty"`
}

// NewSecretVersion creates a new secret version with the given parameters and data.
func NewSecretVersion(projectID, secretID string, versionID string, data []byte) *SecretVersion {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, versionID)
	
	return &SecretVersion{
		Name:       name,
		CreateTime: time.Now().UTC(),
		State:      StateEnabled,
		Etag:       generateEtag(),
		Data:       data,
		Checksum:   generateChecksum(data),
	}
}

// GetProjectID extracts the project ID from the version's resource name.
func (v *SecretVersion) GetProjectID() string {
	return extractProjectID(v.Name)
}

// GetSecretID extracts the secret ID from the version's resource name.
func (v *SecretVersion) GetSecretID() string {
	return extractSecretID(v.Name)
}

// GetVersionID extracts the version ID from the version's resource name.
func (v *SecretVersion) GetVersionID() string {
	versionsPrefix := "/versions/"
	for i := len(v.Name) - 1; i >= 0; i-- {
		if i+len(versionsPrefix) < len(v.Name) && v.Name[i:i+len(versionsPrefix)] == versionsPrefix {
			return v.Name[i+len(versionsPrefix):]
		}
	}
	return ""
}