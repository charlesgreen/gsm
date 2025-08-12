// Package client provides example usage of the Google Secret Manager emulator with the official client library.
package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

// ExampleUsage demonstrates comprehensive usage of the Secret Manager emulator with various operations.
func ExampleUsage() {
	ctx := context.Background()

	client, err := newSecretManagerClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	projectID := getProjectID()
	
	fmt.Println("=== Google Secret Manager Emulator Example ===")

	if err := createSecretExample(ctx, client, projectID); err != nil {
		log.Printf("Create secret failed: %v", err)
	}

	if err := addVersionExample(ctx, client, projectID); err != nil {
		log.Printf("Add version failed: %v", err)
	}

	if err := accessSecretExample(ctx, client, projectID); err != nil {
		log.Printf("Access secret failed: %v", err)
	}

	if err := listSecretsExample(ctx, client, projectID); err != nil {
		log.Printf("List secrets failed: %v", err)
	}

	if err := listVersionsExample(ctx, client, projectID); err != nil {
		log.Printf("List versions failed: %v", err)
	}
}

func newSecretManagerClient(ctx context.Context) (*secretmanager.Client, error) {
	if emulatorHost := os.Getenv("SECRET_MANAGER_EMULATOR_HOST"); emulatorHost != "" {
		fmt.Printf("Using Secret Manager Emulator at: %s\n", emulatorHost)
		return secretmanager.NewClient(ctx,
			option.WithEndpoint("http://"+emulatorHost),
			option.WithoutAuthentication(),
		)
	}

	fmt.Println("Using production Secret Manager")
	return secretmanager.NewClient(ctx)
}

func getProjectID() string {
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		return projectID
	}
	if projectID := os.Getenv("GSM_PROJECT_ID"); projectID != "" {
		return projectID
	}
	return "my-test-project"
}

func createSecretExample(ctx context.Context, client *secretmanager.Client, projectID string) error {
	fmt.Println("\n--- Creating Secret ---")
	
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: "example-secret",
		Secret: &secretmanagerpb.Secret{
			Labels: map[string]string{
				"env":        "development",
				"created-by": "gsm-example",
			},
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	result, err := client.CreateSecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	fmt.Printf("Created secret: %s\n", result.Name)
	return nil
}

func addVersionExample(ctx context.Context, client *secretmanager.Client, projectID string) error {
	fmt.Println("\n--- Adding Secret Version ---")
	
	secretData := "my-super-secret-value"
	
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/example-secret", projectID),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretData),
		},
	}

	result, err := client.AddSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	fmt.Printf("Added secret version: %s\n", result.Name)
	return nil
}

func accessSecretExample(ctx context.Context, client *secretmanager.Client, projectID string) error {
	fmt.Println("\n--- Accessing Secret ---")
	
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/example-secret/versions/latest", projectID),
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to access secret: %w", err)
	}

	fmt.Printf("Secret value: %s\n", string(result.Payload.Data))
	if result.Payload.DataCrc32C != nil {
		fmt.Printf("CRC32C checksum: %d\n", *result.Payload.DataCrc32C)
	}
	return nil
}

func listSecretsExample(ctx context.Context, client *secretmanager.Client, projectID string) error {
	fmt.Println("\n--- Listing Secrets ---")
	
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
	}

	it := client.ListSecrets(ctx, req)
	for {
		secret, err := it.Next()
		if err != nil {
			if err.Error() == "no more items in iterator" {
				break
			}
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		fmt.Printf("Secret: %s\n", secret.Name)
		fmt.Printf("  Created: %v\n", secret.CreateTime.AsTime())
		if len(secret.Labels) > 0 {
			fmt.Printf("  Labels: %v\n", secret.Labels)
		}
	}
	
	return nil
}

func listVersionsExample(ctx context.Context, client *secretmanager.Client, projectID string) error {
	fmt.Println("\n--- Listing Secret Versions ---")
	
	req := &secretmanagerpb.ListSecretVersionsRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/example-secret", projectID),
	}

	it := client.ListSecretVersions(ctx, req)
	for {
		version, err := it.Next()
		if err != nil {
			if err.Error() == "no more items in iterator" {
				break
			}
			return fmt.Errorf("failed to list versions: %w", err)
		}

		fmt.Printf("Version: %s\n", version.Name)
		fmt.Printf("  State: %s\n", version.State.String())
		fmt.Printf("  Created: %v\n", version.CreateTime.AsTime())
	}
	
	return nil
}

// SimpleExample shows basic secret creation and access with minimal setup.
func SimpleExample() {
	fmt.Println("=== Simple Secret Manager Example ===")
	
	ctx := context.Background()
	projectID := "my-project"
	
	_ = os.Setenv("SECRET_MANAGER_EMULATOR_HOST", "localhost:8085")
	
	client, err := secretmanager.NewClient(ctx,
		option.WithEndpoint("http://localhost:8085"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	secretName := fmt.Sprintf("projects/%s/secrets/my-secret", projectID)
	
	_, err = client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: "my-secret",
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Create secret error (may already exist): %v", err)
	}

	_, err = client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("Hello, Secret Manager!"),
		},
	})
	if err != nil {
		log.Printf("Add version failed: %v", err)
		return
	}

	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/latest", secretName),
	})
	if err != nil {
		log.Printf("Access secret failed: %v", err)
		return
	}

	fmt.Printf("Secret value: %s\n", string(result.Payload.Data))
}

// Base64Example demonstrates working with base64-encoded secret data like credentials.
func Base64Example() {
	fmt.Println("=== Base64 Encoded Secret Example ===")
	
	ctx := context.Background()
	_ = os.Setenv("SECRET_MANAGER_EMULATOR_HOST", "localhost:8085")
	
	client, err := secretmanager.NewClient(ctx,
		option.WithEndpoint("http://localhost:8085"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	secretData := map[string]string{
		"username": "admin",
		"password": "super-secret-password",
		"api_key":  "abcd1234567890",
	}
	
	jsonData := fmt.Sprintf(`{"username":"%s","password":"%s","api_key":"%s"}`, 
		secretData["username"], secretData["password"], secretData["api_key"])
	
	encodedData := base64.StdEncoding.EncodeToString([]byte(jsonData))
	
	projectID := "my-project"
	
	_, err = client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: "database-credentials",
		Secret: &secretmanagerpb.Secret{
			Labels: map[string]string{
				"type": "credentials",
				"env":  "production",
			},
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Create secret error (may already exist): %v", err)
	}

	decodedData, _ := base64.StdEncoding.DecodeString(encodedData)
	
	_, err = client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/database-credentials", projectID),
		Payload: &secretmanagerpb.SecretPayload{
			Data: decodedData,
		},
	})
	if err != nil {
		log.Printf("Add version failed: %v", err)
		return
	}

	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/database-credentials/versions/latest", projectID),
	})
	if err != nil {
		log.Printf("Access secret failed: %v", err)
		return
	}

	fmt.Printf("Retrieved credentials: %s\n", string(result.Payload.Data))
}