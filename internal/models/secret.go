package models

import (
	"fmt"
	"time"
)

// Secret represents a Google Secret Manager secret resource.
type Secret struct {
	Name         string            `json:"name"`
	CreateTime   time.Time         `json:"createTime"`
	Labels       map[string]string `json:"labels,omitempty"`
	Replication  Replication       `json:"replication"`
	Etag         string            `json:"etag"`
	Versions     map[string]*SecretVersion `json:"-"`
	VersionCount int               `json:"-"`
}

// Replication describes the replication policy for a secret.
type Replication struct {
	Automatic *AutomaticReplication `json:"automatic,omitempty"`
	UserManaged *UserManagedReplication `json:"userManaged,omitempty"`
}

// AutomaticReplication represents Google-managed replication policy.
type AutomaticReplication struct {
	CustomerManagedEncryption *CustomerManagedEncryption `json:"customerManagedEncryption,omitempty"`
}

// UserManagedReplication represents user-managed replication policy.
type UserManagedReplication struct {
	Replicas []*Replica `json:"replicas"`
}

// Replica represents a single replica location in user-managed replication.
type Replica struct {
	Location                  string                     `json:"location"`
	CustomerManagedEncryption *CustomerManagedEncryption `json:"customerManagedEncryption,omitempty"`
}

// CustomerManagedEncryption represents customer-managed encryption configuration.
type CustomerManagedEncryption struct {
	KmsKeyName string `json:"kmsKeyName"`
}

// NewSecret creates a new secret with the given project ID, secret ID, and labels.
func NewSecret(projectID, secretID string, labels map[string]string) *Secret {
	name := fmt.Sprintf("projects/%s/secrets/%s", projectID, secretID)
	
	return &Secret{
		Name:       name,
		CreateTime: time.Now().UTC(),
		Labels:     labels,
		Replication: Replication{
			Automatic: &AutomaticReplication{},
		},
		Etag:         generateEtag(),
		Versions:     make(map[string]*SecretVersion),
		VersionCount: 0,
	}
}

// GetProjectID extracts the project ID from the secret's resource name.
func (s *Secret) GetProjectID() string {
	return extractProjectID(s.Name)
}

// GetSecretID extracts the secret ID from the secret's resource name.
func (s *Secret) GetSecretID() string {
	return extractSecretID(s.Name)
}

func extractProjectID(name string) string {
	if len(name) < 10 || name[:9] != "projects/" {
		return ""
	}
	
	end := len("projects/")
	for i := end; i < len(name); i++ {
		if name[i] == '/' {
			return name[end:i]
		}
	}
	return name[end:]
}

func extractSecretID(name string) string {
	secretsPrefix := "/secrets/"
	for i := len(name) - 1; i >= 0; i-- {
		if i+len(secretsPrefix) < len(name) && name[i:i+len(secretsPrefix)] == secretsPrefix {
			return name[i+len(secretsPrefix):]
		}
	}
	return ""
}