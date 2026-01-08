package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// KMSKey represents a KMS key
type KMSKey struct {
	KeyID        string
	Alias        string
	Description  string
	KeyState     string
	KeyUsage     string
	KeySpec      string
	CreationDate string
}

// KMSKeys implements Resource for KMS keys
type KMSKeys struct {
	keys []KMSKey
}

// NewKMSKeys creates a new KMSKeys resource
func NewKMSKeys() *KMSKeys {
	return &KMSKeys{
		keys: make([]KMSKey, 0),
	}
}

// Name returns the display name
func (k *KMSKeys) Name() string {
	return "KMS Keys"
}

// Columns returns the column definitions
func (k *KMSKeys) Columns() []Column {
	return []Column{
		{Name: "Key ID", Width: 40},
		{Name: "Alias", Width: 30},
		{Name: "Description", Width: 30},
		{Name: "State", Width: 12},
		{Name: "Usage", Width: 18},
		{Name: "Spec", Width: 15},
	}
}

// Fetch retrieves KMS keys from AWS
func (k *KMSKeys) Fetch(ctx context.Context, c *client.Client) error {
	k.keys = make([]KMSKey, 0)

	// First, get all aliases to map them to keys
	aliasMap := make(map[string]string)
	aliasPaginator := kms.NewListAliasesPaginator(c.KMS(), &kms.ListAliasesInput{})
	for aliasPaginator.HasMorePages() {
		aliasOutput, err := aliasPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list KMS aliases: %w", err)
		}
		for _, alias := range aliasOutput.Aliases {
			if alias.TargetKeyId != nil && alias.AliasName != nil {
				aliasMap[*alias.TargetKeyId] = *alias.AliasName
			}
		}
	}

	// Get all keys
	paginator := kms.NewListKeysPaginator(c.KMS(), &kms.ListKeysInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list KMS keys: %w", err)
		}

		for _, key := range output.Keys {
			keyID := stringValue(key.KeyId)

			// Get detailed key information
			describeOutput, err := c.KMS().DescribeKey(ctx, &kms.DescribeKeyInput{
				KeyId: key.KeyId,
			})
			if err != nil {
				continue
			}

			metadata := describeOutput.KeyMetadata
			kmsKey := KMSKey{
				KeyID:       keyID,
				Alias:       aliasMap[keyID],
				Description: stringValue(metadata.Description),
				KeyState:    string(metadata.KeyState),
				KeyUsage:    string(metadata.KeyUsage),
				KeySpec:     string(metadata.KeySpec),
			}

			if metadata.CreationDate != nil {
				kmsKey.CreationDate = metadata.CreationDate.Format("2006-01-02 15:04:05")
			}

			k.keys = append(k.keys, kmsKey)
		}
	}

	return nil
}

// Rows returns the table data
func (k *KMSKeys) Rows() [][]string {
	rows := make([][]string, len(k.keys))
	for i, key := range k.keys {
		rows[i] = []string{
			key.KeyID,
			key.Alias,
			key.Description,
			key.KeyState,
			key.KeyUsage,
			key.KeySpec,
		}
	}
	return rows
}

// GetID returns the key ID at the given index
func (k *KMSKeys) GetID(index int) string {
	if index >= 0 && index < len(k.keys) {
		return k.keys[index].KeyID
	}
	return ""
}

// QuickActions returns the available quick actions for KMS keys
func (k *KMSKeys) QuickActions() []QuickAction {
	return []QuickAction{}
}
