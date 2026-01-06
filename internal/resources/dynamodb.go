package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBTable represents a DynamoDB table
type DynamoDBTable struct {
	Name         string
	Status       string
	PartitionKey string
	SortKey      string
	ItemCount    int64
	SizeBytes    int64
	BillingMode  string
	CreationDate string
}

// DynamoDBTables implements Resource for DynamoDB tables
type DynamoDBTables struct {
	tables []DynamoDBTable
}

// NewDynamoDBTables creates a new DynamoDBTables resource
func NewDynamoDBTables() *DynamoDBTables {
	return &DynamoDBTables{
		tables: make([]DynamoDBTable, 0),
	}
}

// Name returns the display name
func (d *DynamoDBTables) Name() string {
	return "DynamoDB Tables"
}

// Columns returns the column definitions
func (d *DynamoDBTables) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 35},
		{Name: "Status", Width: 12},
		{Name: "Partition Key", Width: 20},
		{Name: "Sort Key", Width: 20},
		{Name: "Items", Width: 12},
		{Name: "Size", Width: 15},
		{Name: "Billing Mode", Width: 15},
	}
}

// Fetch retrieves DynamoDB tables from AWS
func (d *DynamoDBTables) Fetch(ctx context.Context, c *client.Client) error {
	d.tables = make([]DynamoDBTable, 0)

	paginator := dynamodb.NewListTablesPaginator(c.DynamoDB(), &dynamodb.ListTablesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list DynamoDB tables: %w", err)
		}

		for _, tableName := range output.TableNames {
			// Get detailed table information
			describeOutput, err := c.DynamoDB().DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: &tableName,
			})
			if err != nil {
				continue
			}

			table := describeOutput.Table
			t := DynamoDBTable{
				Name:   tableName,
				Status: string(table.TableStatus),
			}

			if table.ItemCount != nil {
				t.ItemCount = *table.ItemCount
			}

			if table.TableSizeBytes != nil {
				t.SizeBytes = *table.TableSizeBytes
			}

			if table.BillingModeSummary != nil {
				t.BillingMode = string(table.BillingModeSummary.BillingMode)
			} else {
				t.BillingMode = "PROVISIONED"
			}

			// Get key schema
			for _, key := range table.KeySchema {
				keyName := stringValue(key.AttributeName)
				if key.KeyType == "HASH" {
					t.PartitionKey = keyName
				} else if key.KeyType == "RANGE" {
					t.SortKey = keyName
				}
			}

			if table.CreationDateTime != nil {
				t.CreationDate = table.CreationDateTime.Format("2006-01-02 15:04:05")
			}

			d.tables = append(d.tables, t)
		}
	}

	return nil
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Rows returns the table data
func (d *DynamoDBTables) Rows() [][]string {
	rows := make([][]string, len(d.tables))
	for i, table := range d.tables {
		rows[i] = []string{
			table.Name,
			table.Status,
			table.PartitionKey,
			table.SortKey,
			fmt.Sprintf("%d", table.ItemCount),
			formatSize(table.SizeBytes),
			table.BillingMode,
		}
	}
	return rows
}

// GetID returns the table name at the given index
func (d *DynamoDBTables) GetID(index int) string {
	if index >= 0 && index < len(d.tables) {
		return d.tables[index].Name
	}
	return ""
}
