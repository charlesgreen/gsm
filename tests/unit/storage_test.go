package unit

import (
	"context"
	"testing"

	"github.com/charlesgreen/gsm/internal/models"
	"github.com/charlesgreen/gsm/internal/storage"
)

func TestMemoryStorage_CreateSecret(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	
	secret := models.NewSecret("test-project", "test-secret", nil)
	
	err := store.CreateSecret(ctx, "test-project", "test-secret", secret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	err = store.CreateSecret(ctx, "test-project", "test-secret", secret)
	if err != storage.ErrSecretExists {
		t.Fatalf("Expected ErrSecretExists, got %v", err)
	}
}

func TestMemoryStorage_GetSecret(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	
	_, err := store.GetSecret(ctx, "test-project", "nonexistent")
	if err != storage.ErrSecretNotFound {
		t.Fatalf("Expected ErrSecretNotFound, got %v", err)
	}
	
	secret := models.NewSecret("test-project", "test-secret", nil)
	err = store.CreateSecret(ctx, "test-project", "test-secret", secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}
	
	retrieved, err := store.GetSecret(ctx, "test-project", "test-secret")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if retrieved.Name != secret.Name {
		t.Fatalf("Expected secret name %s, got %s", secret.Name, retrieved.Name)
	}
}

func TestMemoryStorage_ListSecrets(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	
	secret1 := models.NewSecret("test-project", "secret1", nil)
	secret2 := models.NewSecret("test-project", "secret2", nil)
	
	_ = store.CreateSecret(ctx, "test-project", "secret1", secret1)
	_ = store.CreateSecret(ctx, "test-project", "secret2", secret2)
	
	secrets, nextToken, err := store.ListSecrets(ctx, "test-project", 10, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(secrets) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(secrets))
	}
	
	if nextToken != "" {
		t.Fatalf("Expected empty next token, got %s", nextToken)
	}
}

func TestMemoryStorage_AddSecretVersion(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	
	_, err := store.AddSecretVersion(ctx, "test-project", "nonexistent", []byte("data"))
	if err != storage.ErrSecretNotFound {
		t.Fatalf("Expected ErrSecretNotFound, got %v", err)
	}
	
	secret := models.NewSecret("test-project", "test-secret", nil)
	_ = store.CreateSecret(ctx, "test-project", "test-secret", secret)
	
	version, err := store.AddSecretVersion(ctx, "test-project", "test-secret", []byte("secret-data"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if version.GetVersionID() != "1" {
		t.Fatalf("Expected version ID '1', got %s", version.GetVersionID())
	}
	
	data, err := store.AccessSecretVersion(ctx, "test-project", "test-secret", "1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if string(data) != "secret-data" {
		t.Fatalf("Expected 'secret-data', got %s", string(data))
	}
}