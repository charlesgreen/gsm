package storage

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/charlesgreen/gsm/internal/models"
)

// MemoryStorage provides in-memory storage for secrets and versions with thread safety.
type MemoryStorage struct {
	mu      sync.RWMutex
	secrets map[string]*models.Secret // key: "projectID/secretID"
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		secrets: make(map[string]*models.Secret),
	}
}

// CreateSecret stores a new secret in memory.
func (m *MemoryStorage) CreateSecret(_ context.Context, projectID, secretID string, secret *models.Secret) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	if _, exists := m.secrets[key]; exists {
		return ErrSecretExists
	}

	m.secrets[key] = secret
	return nil
}

// GetSecret retrieves a secret from memory by project and secret ID.
func (m *MemoryStorage) GetSecret(_ context.Context, projectID, secretID string) (*models.Secret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	secret, exists := m.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}

	return secret, nil
}

// ListSecrets retrieves all secrets for a project with pagination support.
func (m *MemoryStorage) ListSecrets(_ context.Context, projectID string, pageSize int, pageToken string) ([]*models.Secret, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var secrets []*models.Secret
	prefix := projectID + "/"
	
	for key, secret := range m.secrets {
		if strings.HasPrefix(key, prefix) {
			secrets = append(secrets, secret)
		}
	}

	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].Name < secrets[j].Name
	})

	start := 0
	if pageToken != "" {
		startIdx, err := strconv.Atoi(pageToken)
		if err == nil && startIdx >= 0 && startIdx < len(secrets) {
			start = startIdx
		}
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	end := start + pageSize
	if end > len(secrets) {
		end = len(secrets)
	}

	result := secrets[start:end]
	
	var nextPageToken string
	if end < len(secrets) {
		nextPageToken = strconv.Itoa(end)
	}

	return result, nextPageToken, nil
}

// DeleteSecret removes a secret from memory.
func (m *MemoryStorage) DeleteSecret(_ context.Context, projectID, secretID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	if _, exists := m.secrets[key]; !exists {
		return ErrSecretNotFound
	}

	delete(m.secrets, key)
	return nil
}

// AddSecretVersion adds a new version to an existing secret in memory.
func (m *MemoryStorage) AddSecretVersion(_ context.Context, projectID, secretID string, data []byte) (*models.SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	secret, exists := m.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}

	secret.VersionCount++
	versionID := strconv.Itoa(secret.VersionCount)
	
	version := models.NewSecretVersion(projectID, secretID, versionID, data)
	secret.Versions[versionID] = version

	return version, nil
}

// GetSecretVersion retrieves a specific version of a secret from memory.
func (m *MemoryStorage) GetSecretVersion(_ context.Context, projectID, secretID, versionID string) (*models.SecretVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	secret, exists := m.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}

	if versionID == "latest" {
		if secret.VersionCount == 0 {
			return nil, ErrVersionNotFound
		}
		versionID = strconv.Itoa(secret.VersionCount)
	}

	version, exists := secret.Versions[versionID]
	if !exists {
		return nil, ErrVersionNotFound
	}

	return version, nil
}

// ListSecretVersions retrieves all versions of a secret with pagination support.
func (m *MemoryStorage) ListSecretVersions(_ context.Context, projectID, secretID string, pageSize int, pageToken string) ([]*models.SecretVersion, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	secret, exists := m.secrets[key]
	if !exists {
		return nil, "", ErrSecretNotFound
	}

	versions := make([]*models.SecretVersion, 0, len(secret.Versions))
	for _, version := range secret.Versions {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool {
		iVersion := versions[i].GetVersionID()
		jVersion := versions[j].GetVersionID()
		
		iNum, iErr := strconv.Atoi(iVersion)
		jNum, jErr := strconv.Atoi(jVersion)
		
		if iErr == nil && jErr == nil {
			return iNum > jNum // Latest first
		}
		return iVersion > jVersion
	})

	start := 0
	if pageToken != "" {
		startIdx, err := strconv.Atoi(pageToken)
		if err == nil && startIdx >= 0 && startIdx < len(versions) {
			start = startIdx
		}
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	end := start + pageSize
	if end > len(versions) {
		end = len(versions)
	}

	result := versions[start:end]
	
	var nextPageToken string
	if end < len(versions) {
		nextPageToken = strconv.Itoa(end)
	}

	return result, nextPageToken, nil
}

// DeleteSecretVersion removes a specific version of a secret from memory.
func (m *MemoryStorage) DeleteSecretVersion(_ context.Context, projectID, secretID, versionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s", projectID, secretID)
	secret, exists := m.secrets[key]
	if !exists {
		return ErrSecretNotFound
	}

	if _, exists := secret.Versions[versionID]; !exists {
		return ErrVersionNotFound
	}

	delete(secret.Versions, versionID)
	return nil
}

// AccessSecretVersion retrieves the raw data of a specific secret version.
func (m *MemoryStorage) AccessSecretVersion(_ context.Context, projectID, secretID, versionID string) ([]byte, error) {
	version, err := m.GetSecretVersion(context.TODO(), projectID, secretID, versionID)
	if err != nil {
		return nil, err
	}

	return version.Data, nil
}

// Close releases any resources used by the memory storage (no-op for memory storage).
func (m *MemoryStorage) Close() error {
	return nil
}