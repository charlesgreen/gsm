package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charlesgreen/gsm/internal/api/routes"
	"github.com/charlesgreen/gsm/internal/models"
	"github.com/charlesgreen/gsm/internal/storage"
)

// TestGSMEmulatorProductionParity tests that emulator behavior matches production
func TestGSMEmulatorProductionParity(t *testing.T) {
	storage := storage.NewMemoryStorage()
	router := routes.SetupRoutes(storage)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedError  *models.ErrorResponse
		description    string
		setup          func() // Optional setup function
	}{
		{
			name:           "CreateSecret_Success",
			method:         "POST",
			path:           "/v1/projects/test-project/secrets",
			body: map[string]interface{}{
				"secretId": "test-secret-1",
				"secret": map[string]interface{}{
					"labels": map[string]string{
						"type": "test-secret",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			description:    "Creating new secret should return 201",
		},
		{
			name:           "CreateSecret_AlreadyExists",
			method:         "POST",
			path:           "/v1/projects/test-project/secrets",
			body: map[string]interface{}{
				"secretId": "test-secret-1", // Same as above
				"secret": map[string]interface{}{
					"labels": map[string]string{
						"type": "test-secret",
					},
				},
			},
			expectedStatus: http.StatusConflict,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    409,
					Message: "Secret [projects/test-project/secrets/test-secret-1] already exists.",
					Status:  "ALREADY_EXISTS",
				},
			},
			description: "Creating duplicate secret should return 409 Conflict",
		},
		{
			name:           "GetSecret_NotFound",
			method:         "GET",
			path:           "/v1/projects/test-project/secrets/non-existent-secret",
			expectedStatus: http.StatusNotFound,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    404,
					Message: "Secret [projects/test-project/secrets/non-existent-secret] not found.",
					Status:  "NOT_FOUND",
				},
			},
			description: "Getting non-existent secret should return 404",
		},
		{
			name:           "GetSecret_Success",
			method:         "GET",
			path:           "/v1/projects/test-project/secrets/test-secret-1",
			expectedStatus: http.StatusOK,
			description:    "Getting existing secret should return 200",
		},
		{
			name:           "AccessVersion_SecretNotFound",
			method:         "GET",
			path:           "/v1/projects/test-project/secrets/non-existent-secret/versions/latest:access",
			expectedStatus: http.StatusNotFound,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    404,
					Message: "Secret [projects/test-project/secrets/non-existent-secret] not found.",
					Status:  "NOT_FOUND",
				},
			},
			description: "Accessing version of non-existent secret should return 404",
		},
		{
			name:           "AccessVersion_VersionNotFound",
			method:         "GET",
			path:           "/v1/projects/test-project/secrets/test-secret-1/versions/999:access",
			expectedStatus: http.StatusNotFound,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    404,
					Message: "Secret Version [projects/test-project/secrets/test-secret-1/versions/999] not found.",
					Status:  "NOT_FOUND",
				},
			},
			description: "Accessing non-existent version should return 404",
		},
		{
			name:           "CreateSecret_InvalidRequestBody",
			method:         "POST",
			path:           "/v1/projects/test-project/secrets",
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    400,
					Message: "Invalid request body",
					Status:  "INVALID_ARGUMENT",
				},
			},
			description: "Invalid request body should return 400",
		},
		{
			name:           "CreateSecret_MissingSecretId",
			method:         "POST",
			path:           "/v1/projects/test-project/secrets",
			body: map[string]interface{}{
				"secret": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError: &models.ErrorResponse{
				Error: &models.ErrorDetail{
					Code:    400,
					Message: "secretId is required",
					Status:  "INVALID_ARGUMENT",
				},
			},
			description: "Missing secretId should return 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			var body []byte
			var err error
			
			if tt.body != nil {
				if bodyStr, ok := tt.body.(string); ok {
					body = []byte(bodyStr)
				} else {
					body, err = json.Marshal(tt.body)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for %s",
					tt.expectedStatus, w.Code, tt.description)
				t.Logf("Response body: %s", w.Body.String())
			}

			// Check error response format if expected
			if tt.expectedError != nil {
				var errorResp models.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
					t.Errorf("Error response should be valid JSON: %v", err)
					t.Logf("Response body: %s", w.Body.String())
					return
				}

				if errorResp.Error.Code != tt.expectedError.Error.Code {
					t.Errorf("Expected error code %d, got %d", 
						tt.expectedError.Error.Code, errorResp.Error.Code)
				}

				if errorResp.Error.Status != tt.expectedError.Error.Status {
					t.Errorf("Expected error status %s, got %s",
						tt.expectedError.Error.Status, errorResp.Error.Status)
				}

				if errorResp.Error.Message != tt.expectedError.Error.Message {
					t.Errorf("Expected error message '%s', got '%s'",
						tt.expectedError.Error.Message, errorResp.Error.Message)
				}
			}
		})
	}
}

// TestErrorResponseFormat ensures error responses match production format
func TestErrorResponseFormat(t *testing.T) {
	storage := storage.NewMemoryStorage()
	router := routes.SetupRoutes(storage)

	// Test 404 error format for non-existent secret
	req := httptest.NewRequest("GET", "/v1/projects/test-project/secrets/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
		return
	}

	var errorResp models.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
		t.Errorf("Error response should be valid JSON: %v", err)
		return
	}

	// Check if error format matches Google Cloud error format
	if errorResp.Error == nil {
		t.Error("Error response should have 'error' field")
		return
	}

	// Verify required fields
	requiredFields := []struct {
		name  string
		value interface{}
	}{
		{"code", errorResp.Error.Code},
		{"message", errorResp.Error.Message},
		{"status", errorResp.Error.Status},
	}

	for _, field := range requiredFields {
		if field.value == nil || field.value == "" || field.value == 0 {
			t.Errorf("Error response missing required field: %s", field.name)
		}
	}

	// Verify Google Cloud format
	if errorResp.Error.Code != 404 {
		t.Errorf("Expected error code 404, got %d", errorResp.Error.Code)
	}

	if errorResp.Error.Status != "NOT_FOUND" {
		t.Errorf("Expected error status 'NOT_FOUND', got '%s'", errorResp.Error.Status)
	}

	expectedMessage := "Secret [projects/test-project/secrets/non-existent] not found."
	if errorResp.Error.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, errorResp.Error.Message)
	}
}

// TestSecretVersionErrorFormat tests version-specific error formats
func TestSecretVersionErrorFormat(t *testing.T) {
	storage := storage.NewMemoryStorage()
	router := routes.SetupRoutes(storage)

	tests := []struct {
		name            string
		path            string
		expectedMessage string
	}{
		{
			name:            "Version_SecretNotFound",
			path:            "/v1/projects/test-project/secrets/non-existent/versions/latest:access",
			expectedMessage: "Secret [projects/test-project/secrets/non-existent] not found.",
		},
		{
			name:            "Version_VersionNotFound",
			path:            "/v1/projects/test-project/secrets/existing-secret/versions/999:access",
			expectedMessage: "Secret Version [projects/test-project/secrets/existing-secret/versions/999] not found.",
		},
	}

	// Create a secret for the version not found test
	secretReq := map[string]interface{}{
		"secretId": "existing-secret",
		"secret":   map[string]interface{}{},
	}
	body, _ := json.Marshal(secretReq)
	req := httptest.NewRequest("POST", "/v1/projects/test-project/secrets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected 404, got %d", w.Code)
				return
			}

			var errorResp models.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
				t.Errorf("Error response should be valid JSON: %v", err)
				return
			}

			if errorResp.Error.Message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", 
					tt.expectedMessage, errorResp.Error.Message)
			}
		})
	}
}

// TestProductionParityIntegration runs the exact test cases provided in the bug report
func TestProductionParityIntegration(t *testing.T) {
	storage := storage.NewMemoryStorage()
	router := routes.SetupRoutes(storage)
	
	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	emulatorURL := server.URL
	projectID := "test-project"

	tests := []struct {
		name           string
		operation      string
		secretName     string
		expectedStatus int
		shouldMatch    bool
		description    string
	}{
		{
			name:           "CreateSecret_Success",
			operation:      "create",
			secretName:     "test-secret-1",
			expectedStatus: 201, // Fixed: should be 201 for creation
			shouldMatch:    true,
			description:    "Creating new secret should return 201",
		},
		{
			name:           "CreateSecret_AlreadyExists",
			operation:      "create",
			secretName:     "test-secret-1", // Same as above
			expectedStatus: 409,
			shouldMatch:    true,
			description:    "Creating duplicate secret should return 409 Conflict",
		},
		{
			name:           "AccessSecret_NotFound",
			operation:      "access",
			secretName:     "non-existent-secret",
			expectedStatus: 404,
			shouldMatch:    true,
			description:    "Accessing non-existent secret should return 404",
		},
		{
			name:           "AccessSecret_Success",
			operation:      "access", 
			secretName:     "test-secret-1",
			expectedStatus: 200,
			shouldMatch:    true,
			description:    "Accessing existing secret should return 200",
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			switch tt.operation {
			case "create":
				resp, err = createSecret(client, emulatorURL, projectID, tt.secretName)
			case "access":
				resp, err = accessSecret(client, emulatorURL, projectID, tt.secretName)
			default:
				t.Fatalf("Unknown operation: %s", tt.operation)
			}

			if err != nil {
				t.Fatalf("Operation failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for %s",
					tt.expectedStatus, resp.StatusCode, tt.description)

				// Log response body for debugging
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				t.Logf("Response body: %s", string(body[:n]))
			}
		})
	}
}

func createSecret(client *http.Client, baseURL, projectID, secretName string) (*http.Response, error) {
	createReq := map[string]interface{}{
		"secretId": secretName,
		"secret": map[string]interface{}{
			"labels": map[string]string{
				"type": "test-secret",
			},
		},
	}

	reqBody, _ := json.Marshal(createReq)
	url := fmt.Sprintf("%s/v1/projects/%s/secrets", baseURL, projectID)

	return client.Post(url, "application/json", bytes.NewBuffer(reqBody))
}

func accessSecret(client *http.Client, baseURL, projectID, secretName string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/projects/%s/secrets/%s", baseURL, projectID, secretName)
	return client.Get(url)
}