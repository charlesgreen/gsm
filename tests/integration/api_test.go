package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charlesgreen/gsm/internal/api/routes"
	"github.com/charlesgreen/gsm/internal/models"
	"github.com/charlesgreen/gsm/internal/storage"
)

func TestHealthEndpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var health models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if health.Status != "OK" {
		t.Errorf("Expected status OK, got %s", health.Status)
	}
}

func TestCreateSecret(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	createReq := models.CreateSecretRequest{
		SecretID: "test-secret",
		Secret: &models.CreateSecretData{
			Labels: map[string]string{"env": "test"},
		},
	}

	body, _ := json.Marshal(createReq)
	req, err := http.NewRequest("POST", "/v1/projects/test-project/secrets", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}

	var secret models.Secret
	if err := json.Unmarshal(rr.Body.Bytes(), &secret); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedName := "projects/test-project/secrets/test-secret"
	if secret.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, secret.Name)
	}
}

func TestGetSecret(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	secret := models.NewSecret("test-project", "test-secret", map[string]string{"env": "test"})
	_ = store.CreateSecret(context.Background(), "test-project", "test-secret", secret)

	req, err := http.NewRequest("GET", "/v1/projects/test-project/secrets/test-secret", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var retrievedSecret models.Secret
	if err := json.Unmarshal(rr.Body.Bytes(), &retrievedSecret); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if retrievedSecret.Name != secret.Name {
		t.Errorf("Expected name %s, got %s", secret.Name, retrievedSecret.Name)
	}
}

func TestAddSecretVersion(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	secret := models.NewSecret("test-project", "test-secret", nil)
	_ = store.CreateSecret(context.Background(), "test-project", "test-secret", secret)

	addVersionReq := models.AddSecretVersionRequest{
		Payload: &models.SecretPayload{
			Data: []byte("my-secret-value"),
		},
	}

	body, _ := json.Marshal(addVersionReq)
	req, err := http.NewRequest("POST", "/v1/projects/test-project/secrets/test-secret:addVersion", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}

	var version models.SecretVersion
	if err := json.Unmarshal(rr.Body.Bytes(), &version); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedName := "projects/test-project/secrets/test-secret/versions/1"
	if version.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, version.Name)
	}
}

func TestAccessSecretVersion(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	secret := models.NewSecret("test-project", "test-secret", nil)
	_ = store.CreateSecret(context.Background(), "test-project", "test-secret", secret)

	secretData := []byte("my-secret-value")
	_, _ = store.AddSecretVersion(context.Background(), "test-project", "test-secret", secretData)

	req, err := http.NewRequest("GET", "/v1/projects/test-project/secrets/test-secret/versions/1:access", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var accessResp models.AccessSecretVersionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &accessResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if string(accessResp.Payload.Data) != string(secretData) {
		t.Errorf("Expected data %s, got %s", string(secretData), string(accessResp.Payload.Data))
	}
}

func TestListSecrets(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	secret1 := models.NewSecret("test-project", "secret1", nil)
	secret2 := models.NewSecret("test-project", "secret2", nil)
	_ = store.CreateSecret(context.Background(), "test-project", "secret1", secret1)
	_ = store.CreateSecret(context.Background(), "test-project", "secret2", secret2)

	req, err := http.NewRequest("GET", "/v1/projects/test-project/secrets", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var listResp models.ListSecretsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(listResp.Secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(listResp.Secrets))
	}
}

func TestNotFoundEndpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	router := routes.SetupRoutes(store)

	req, err := http.NewRequest("GET", "/v1/projects/test-project/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
	}
}
