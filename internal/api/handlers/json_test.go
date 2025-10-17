package handlers

import (
	"bytes"
	"testing"

	"github.com/charlesgreen/gsm/internal/models"
)

// example structure to verify we don't normalize user-specified field naming, while
// adjusting the rest.
type example struct {
	Labels         map[string]int `json:"labels"`
	Annotations    map[string]int `json:"annotations"`
	VersionAliases map[string]int `json:"versionAliases"`
	SecretID       int            `json:"secretId"`
	Secret         models.Secret  `json:"secret"`
}

func TestDecode(t *testing.T) {
	data := []byte(`{
		"labels": {
			"keep_me": 1
		},
		"annotations": {
			"keep_me": 1
		},
		"version_aliases": {
			"keep_me": 1
		},
		"secret_id": 1,
		"secret": {
			"replication": {
				"user_managed": {
					"replicas": [{"location": "foo"}]
				}
			}
		}
	}`)

	buf := bytes.NewBuffer(data)
	var req example
	if err := decodeJSON(buf, &req); err != nil {
		t.Fatal(err)
	}

	expected := 1

	got := req.Labels["keep_me"]
	if got != expected {
		t.Fatalf("expected %d, got %d", expected, got)
	}
	got = req.Annotations["keep_me"]
	if got != expected {
		t.Fatalf("expected %d, got %d", expected, got)
	}
	got = req.VersionAliases["keep_me"]
	if got != expected {
		t.Fatalf("expected %d, got %d", expected, got)
	}
	got = req.SecretID
	if got != expected {
		t.Fatalf("expected %d, got %d", expected, got)
	}

	replicas := req.Secret.Replication.UserManaged.Replicas
	if len(replicas) != 1 {
		t.Fatalf("expected len == 1")
	}
	if replicas[0].Location != "foo" {
		t.Fatalf("expected %s, got %s", "foo", replicas[0].Location)
	}
}
