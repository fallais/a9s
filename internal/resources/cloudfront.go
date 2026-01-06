package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

// CloudFrontDistribution represents a CloudFront distribution
type CloudFrontDistribution struct {
	ID           string
	DomainName   string
	Status       string
	Enabled      bool
	Origins      int
	PriceClass   string
	Aliases      string
	LastModified string
}

// CloudFrontDistributions implements Resource for CloudFront distributions
type CloudFrontDistributions struct {
	distributions []CloudFrontDistribution
}

// NewCloudFrontDistributions creates a new CloudFrontDistributions resource
func NewCloudFrontDistributions() *CloudFrontDistributions {
	return &CloudFrontDistributions{
		distributions: make([]CloudFrontDistribution, 0),
	}
}

// Name returns the display name
func (c *CloudFrontDistributions) Name() string {
	return "CloudFront Distributions"
}

// Columns returns the column definitions
func (c *CloudFrontDistributions) Columns() []Column {
	return []Column{
		{Name: "ID", Width: 16},
		{Name: "Domain Name", Width: 40},
		{Name: "Status", Width: 12},
		{Name: "Enabled", Width: 8},
		{Name: "Origins", Width: 8},
		{Name: "Price Class", Width: 20},
		{Name: "Aliases", Width: 30},
	}
}

// Fetch retrieves CloudFront distributions from AWS
func (c *CloudFrontDistributions) Fetch(ctx context.Context, cl *client.Client) error {
	c.distributions = make([]CloudFrontDistribution, 0)

	paginator := cloudfront.NewListDistributionsPaginator(cl.CloudFront(), &cloudfront.ListDistributionsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list CloudFront distributions: %w", err)
		}

		if output.DistributionList == nil || output.DistributionList.Items == nil {
			continue
		}

		for _, dist := range output.DistributionList.Items {
			d := CloudFrontDistribution{
				ID:         stringValue(dist.Id),
				DomainName: stringValue(dist.DomainName),
				Status:     stringValue(dist.Status),
				Enabled:    dist.Enabled != nil && *dist.Enabled,
				PriceClass: string(dist.PriceClass),
			}

			if dist.Origins != nil && dist.Origins.Quantity != nil {
				d.Origins = int(*dist.Origins.Quantity)
			}

			if dist.Aliases != nil && dist.Aliases.Items != nil {
				aliases := ""
				for i, alias := range dist.Aliases.Items {
					if i > 0 {
						aliases += ", "
					}
					aliases += alias
				}
				d.Aliases = aliases
			}

			if dist.LastModifiedTime != nil {
				d.LastModified = dist.LastModifiedTime.Format("2006-01-02 15:04:05")
			}

			c.distributions = append(c.distributions, d)
		}
	}

	return nil
}

// Rows returns the table data
func (c *CloudFrontDistributions) Rows() [][]string {
	rows := make([][]string, len(c.distributions))
	for i, dist := range c.distributions {
		enabled := "No"
		if dist.Enabled {
			enabled = "Yes"
		}
		rows[i] = []string{
			dist.ID,
			dist.DomainName,
			dist.Status,
			enabled,
			fmt.Sprintf("%d", dist.Origins),
			dist.PriceClass,
			dist.Aliases,
		}
	}
	return rows
}

// GetID returns the distribution ID at the given index
func (c *CloudFrontDistributions) GetID(index int) string {
	if index >= 0 && index < len(c.distributions) {
		return c.distributions[index].ID
	}
	return ""
}
