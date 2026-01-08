package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Secret represents a Secrets Manager secret
type Secret struct {
	ARN              string
	Name             string
	Description      string
	RotationEnabled  bool
	LastAccessedDate string
	LastChangedDate  string
	CreatedDate      string
}

// Secrets implements Resource for Secrets Manager secrets
type Secrets struct {
	secrets []Secret
}

// NewSecrets creates a new Secrets resource
func NewSecrets() *Secrets {
	return &Secrets{
		secrets: make([]Secret, 0),
	}
}

// Name returns the display name
func (s *Secrets) Name() string {
	return "Secrets Manager"
}

// Columns returns the column definitions
func (s *Secrets) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 40},
		{Name: "Description", Width: 35},
		{Name: "Rotation", Width: 10},
		{Name: "Last Accessed", Width: 20},
		{Name: "Last Changed", Width: 20},
		{Name: "Created", Width: 20},
	}
}

// Fetch retrieves secrets from AWS Secrets Manager
func (s *Secrets) Fetch(ctx context.Context, c *client.Client) error {
	s.secrets = make([]Secret, 0)

	paginator := secretsmanager.NewListSecretsPaginator(c.SecretsManager(), &secretsmanager.ListSecretsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range output.SecretList {
			sec := Secret{
				ARN:             stringValue(secret.ARN),
				Name:            stringValue(secret.Name),
				Description:     stringValue(secret.Description),
				RotationEnabled: secret.RotationEnabled != nil && *secret.RotationEnabled,
			}

			if secret.LastAccessedDate != nil {
				sec.LastAccessedDate = secret.LastAccessedDate.Format("2006-01-02 15:04:05")
			}

			if secret.LastChangedDate != nil {
				sec.LastChangedDate = secret.LastChangedDate.Format("2006-01-02 15:04:05")
			}

			if secret.CreatedDate != nil {
				sec.CreatedDate = secret.CreatedDate.Format("2006-01-02 15:04:05")
			}

			s.secrets = append(s.secrets, sec)
		}
	}

	return nil
}

// Rows returns the table data
func (s *Secrets) Rows() [][]string {
	rows := make([][]string, len(s.secrets))
	for i, secret := range s.secrets {
		rotation := "No"
		if secret.RotationEnabled {
			rotation = "Yes"
		}
		rows[i] = []string{
			secret.Name,
			secret.Description,
			rotation,
			secret.LastAccessedDate,
			secret.LastChangedDate,
			secret.CreatedDate,
		}
	}
	return rows
}

// GetID returns the secret ARN at the given index
func (s *Secrets) GetID(index int) string {
	if index >= 0 && index < len(s.secrets) {
		return s.secrets[index].ARN
	}
	return ""
}

// QuickActions returns the available quick actions for secrets
func (s *Secrets) QuickActions() []QuickAction {
	return []QuickAction{}
}
