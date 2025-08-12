package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charlesgreen/gsm/internal/models"
)

// PersistentStorage provides file-backed storage for secrets and versions.
type PersistentStorage struct {
	*MemoryStorage
	filePath string
	mu       sync.RWMutex
}

// Data represents the JSON structure for persisted storage data.
type Data struct {
	Secrets   map[string]*models.Secret `json:"secrets"`
	Timestamp time.Time                 `json:"timestamp"`
	Version   string                    `json:"version"`
}

// NewPersistentStorage creates a new persistent storage instance that saves data to the specified file.
func NewPersistentStorage(filePath string) (*PersistentStorage, error) {
	return &PersistentStorage{
		MemoryStorage: NewMemoryStorage(),
		filePath:      filePath,
	}, nil
}

// Load reads and restores secrets from the persistent storage file.
func (p *PersistentStorage) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	var storageData Data
	if err := json.Unmarshal(data, &storageData); err != nil {
		return fmt.Errorf("failed to parse storage file: %w", err)
	}

	p.mu.Lock()
	p.secrets = storageData.Secrets
	p.mu.Unlock()

	return nil
}

// Save writes the current state of secrets to the persistent storage file.
func (p *PersistentStorage) Save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.mu.RLock()
	storageData := Data{
		Secrets:   p.secrets,
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}
	p.mu.RUnlock()

	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal storage data: %w", err)
	}

	if err := os.WriteFile(p.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// CreateSecret creates a new secret and persists it to storage.
func (p *PersistentStorage) CreateSecret(ctx context.Context, projectID, secretID string, secret *models.Secret) error {
	if err := p.MemoryStorage.CreateSecret(ctx, projectID, secretID, secret); err != nil {
		return err
	}
	return p.Save()
}

// DeleteSecret removes a secret and persists the change to storage.
func (p *PersistentStorage) DeleteSecret(ctx context.Context, projectID, secretID string) error {
	if err := p.MemoryStorage.DeleteSecret(ctx, projectID, secretID); err != nil {
		return err
	}
	return p.Save()
}

// AddSecretVersion adds a new version to an existing secret and persists it to storage.
func (p *PersistentStorage) AddSecretVersion(ctx context.Context, projectID, secretID string, data []byte) (*models.SecretVersion, error) {
	version, err := p.MemoryStorage.AddSecretVersion(ctx, projectID, secretID, data)
	if err != nil {
		return nil, err
	}
	
	if err := p.Save(); err != nil {
		p.mu.Lock()
		key := fmt.Sprintf("%s/%s", projectID, secretID)
		if secret, exists := p.secrets[key]; exists {
			delete(secret.Versions, version.GetVersionID())
			secret.VersionCount--
		}
		p.mu.Unlock()
		return nil, err
	}
	
	return version, nil
}

// DeleteSecretVersion removes a secret version and persists the change to storage.
func (p *PersistentStorage) DeleteSecretVersion(ctx context.Context, projectID, secretID, versionID string) error {
	if err := p.MemoryStorage.DeleteSecretVersion(ctx, projectID, secretID, versionID); err != nil {
		return err
	}
	return p.Save()
}

// Close saves the current state to disk and releases resources.
func (p *PersistentStorage) Close() error {
	return p.Save()
}