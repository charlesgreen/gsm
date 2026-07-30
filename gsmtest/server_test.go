package gsmtest_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/charlesgreen/gsm/gsmtest"
)

func TestTCP(t *testing.T) {
	gsm, err := gsmtest.New(t)
	if err != nil {
		t.Fatal(err)
	}
	testFlow(t, gsm)
}

func TestMem(t *testing.T) {
	gsm, err := gsmtest.New(t, gsmtest.InMemory())
	if err != nil {
		t.Fatal(err)
	}
	testFlow(t, gsm)
}

func TestPersistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage.json")
	gsm, err := gsmtest.New(t, gsmtest.InMemory(), gsmtest.StorageFile(path))
	if err != nil {
		t.Fatal(err)
	}
	testFlow(t, gsm)

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected changes to persist in file")
	}
}

func testFlow(t testing.TB, gsm *gsmtest.SecretManager) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = gsm.Start(ctx) }()

	client, err := gsm.Client(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Create an initial secret
	secret, err := client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   "projects/foo",
		SecretId: "bar",
		Secret:   &secretmanagerpb.Secret{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Append a new version
	data := []byte("shhhh")
	version, err := client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent:  secret.Name,
		Payload: &secretmanagerpb.SecretPayload{Data: data},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Confirm we receive the same version out
	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: version.Name,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(data, resp.Payload.Data) {
		t.Fatalf("expected %s, got %s", data, resp.Payload.Data)
	}
}
