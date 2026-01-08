package resources

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// BillingEntry represents a billing line item
type BillingEntry struct {
	Service    string
	Amount     float64
	Currency   string
	Percentage float64
}

// Billing implements Resource for AWS billing information
type Billing struct {
	entries     []BillingEntry
	totalAmount float64
	currency    string
	periodStart string
	periodEnd   string
}

// NewBilling creates a new Billing resource
func NewBilling() *Billing {
	return &Billing{
		entries: make([]BillingEntry, 0),
	}
}

// Name returns the display name
func (b *Billing) Name() string {
	return "Billing (Current Month)"
}

// Columns returns the column definitions
func (b *Billing) Columns() []Column {
	return []Column{
		{Name: "Service", Width: 40},
		{Name: "Cost", Width: 15},
		{Name: "%", Width: 8},
		{Name: "Distribution", Width: 30},
	}
}

// Fetch retrieves billing information from AWS Cost Explorer
func (b *Billing) Fetch(ctx context.Context, c *client.Client) error {
	b.entries = make([]BillingEntry, 0)
	b.totalAmount = 0

	// Get current month date range
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	b.periodStart = startOfMonth.Format("2006-01-02")
	b.periodEnd = endOfMonth.Format("2006-01-02")

	// Get cost by service
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(b.periodStart),
			End:   aws.String(b.periodEnd),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	output, err := c.CostExplorer().GetCostAndUsage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get billing data: %w", err)
	}

	// Parse results
	for _, result := range output.ResultsByTime {
		for _, group := range result.Groups {
			serviceName := ""
			if len(group.Keys) > 0 {
				serviceName = group.Keys[0]
			}

			if cost, ok := group.Metrics["UnblendedCost"]; ok {
				amount, _ := strconv.ParseFloat(aws.ToString(cost.Amount), 64)
				currency := aws.ToString(cost.Unit)

				if amount > 0.001 { // Filter out negligible amounts
					b.entries = append(b.entries, BillingEntry{
						Service:  serviceName,
						Amount:   amount,
						Currency: currency,
					})
					b.totalAmount += amount
					b.currency = currency
				}
			}
		}
	}

	// Sort by amount descending
	sort.Slice(b.entries, func(i, j int) bool {
		return b.entries[i].Amount > b.entries[j].Amount
	})

	// Calculate percentages
	for i := range b.entries {
		if b.totalAmount > 0 {
			b.entries[i].Percentage = (b.entries[i].Amount / b.totalAmount) * 100
		}
	}

	return nil
}

// Rows returns the data rows for the table
func (b *Billing) Rows() [][]string {
	rows := make([][]string, 0, len(b.entries)+2)

	// Add header row with total
	rows = append(rows, []string{
		fmt.Sprintf("📊 TOTAL (%s to %s)", b.periodStart, b.periodEnd),
		fmt.Sprintf("%.2f %s", b.totalAmount, b.currency),
		"100%",
		"████████████████████████████████",
	})

	// Add separator
	rows = append(rows, []string{
		"────────────────────────────────────────",
		"───────────────",
		"────────",
		"──────────────────────────────",
	})

	// Add service entries
	for _, entry := range b.entries {
		rows = append(rows, []string{
			entry.Service,
			fmt.Sprintf("%.2f %s", entry.Amount, entry.Currency),
			fmt.Sprintf("%.1f%%", entry.Percentage),
			b.renderBar(entry.Percentage),
		})
	}

	return rows
}

// renderBar creates a simple text-based bar chart
func (b *Billing) renderBar(percentage float64) string {
	maxWidth := 30
	filled := int((percentage / 100) * float64(maxWidth))
	if filled < 1 && percentage > 0 {
		filled = 1
	}

	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", maxWidth-filled)

	return bar + empty
}

// GetID returns the ID of the resource at the given index
func (b *Billing) GetID(index int) string {
	// Adjust for header rows
	if index < 2 {
		return ""
	}
	actualIndex := index - 2
	if actualIndex >= 0 && actualIndex < len(b.entries) {
		return b.entries[actualIndex].Service
	}
	return ""
}

// QuickActions returns the available quick actions for billing
func (b *Billing) QuickActions() []QuickAction {
	return []QuickAction{}
}
