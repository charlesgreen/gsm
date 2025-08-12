// Package storage provides data storage interfaces and implementations for the GSM emulator.
package storage

import (
	"context"
	"errors"
	"github.com/charlesgreen/gsm/internal/models"
)

var (
	// ErrSecretNotFound is returned when a requested secret does not exist.
	ErrSecretNotFound  = errors.New("secret not found")
	// ErrVersionNotFound is returned when a requested secret version does not exist.
	ErrVersionNotFound = errors.New("version not found")
	// ErrSecretExists is returned when attempting to create a secret that already exists.
	ErrSecretExists    = errors.New("secret already exists")
)

// Storage defines the interface for secret storage operations.
type Storage interface {
	CreateSecret(ctx context.Context, projectID, secretID string, secret *models.Secret) error
	GetSecret(ctx context.Context, projectID, secretID string) (*models.Secret, error)
	ListSecrets(ctx context.Context, projectID string, pageSize int, pageToken string) ([]*models.Secret, string, error)
	DeleteSecret(ctx context.Context, projectID, secretID string) error
	
	AddSecretVersion(ctx context.Context, projectID, secretID string, data []byte) (*models.SecretVersion, error)
	GetSecretVersion(ctx context.Context, projectID, secretID, versionID string) (*models.SecretVersion, error)
	ListSecretVersions(ctx context.Context, projectID, secretID string, pageSize int, pageToken string) ([]*models.SecretVersion, string, error)
	DeleteSecretVersion(ctx context.Context, projectID, secretID, versionID string) error
	
	AccessSecretVersion(ctx context.Context, projectID, secretID, versionID string) ([]byte, error)
	
	Close() error
}