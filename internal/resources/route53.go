package resources

import (
	"context"
	"fmt"
	"strings"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// HostedZone represents a Route53 hosted zone
type HostedZone struct {
	ID              string
	Name            string
	Type            string
	RecordSetCount  string
	Comment         string
}

// HostedZones implements Resource for Route53 hosted zones
type HostedZones struct {
	zones []HostedZone
}

// NewHostedZones creates a new HostedZones resource
func NewHostedZones() *HostedZones {
	return &HostedZones{
		zones: make([]HostedZone, 0),
	}
}

// Name returns the display name
func (h *HostedZones) Name() string {
	return "Route53 Hosted Zones"
}

// Columns returns the column definitions
func (h *HostedZones) Columns() []Column {
	return []Column{
		{Name: "Zone ID", Width: 25},
		{Name: "Name", Width: 40},
		{Name: "Type", Width: 12},
		{Name: "Records", Width: 10},
		{Name: "Comment", Width: 50},
	}
}

// Fetch retrieves Route53 hosted zones from AWS
func (h *HostedZones) Fetch(ctx context.Context, c *client.Client) error {
	h.zones = make([]HostedZone, 0)

	paginator := route53.NewListHostedZonesPaginator(c.Route53(), &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list Route53 hosted zones: %w", err)
		}

		for _, zone := range output.HostedZones {
			// Extract zone ID from the full ID path
			// Format: /hostedzone/Z1234567890ABC
			zoneID := stringValue(zone.Id)
			if strings.HasPrefix(zoneID, "/hostedzone/") {
				zoneID = strings.TrimPrefix(zoneID, "/hostedzone/")
			}

			zoneType := "Public"
			if zone.Config != nil && zone.Config.PrivateZone {
				zoneType = "Private"
			}

			comment := ""
			if zone.Config != nil && zone.Config.Comment != nil {
				comment = *zone.Config.Comment
			}

			h.zones = append(h.zones, HostedZone{
				ID:             zoneID,
				Name:           stringValue(zone.Name),
				Type:           zoneType,
				RecordSetCount: fmt.Sprintf("%d", ptrInt64Value(zone.ResourceRecordSetCount)),
				Comment:        comment,
			})
		}
	}

	return nil
}

// Rows returns the table data
func (h *HostedZones) Rows() [][]string {
	rows := make([][]string, len(h.zones))
	for i, zone := range h.zones {
		rows[i] = []string{
			zone.ID,
			zone.Name,
			zone.Type,
			zone.RecordSetCount,
			zone.Comment,
		}
	}
	return rows
}

// GetID returns the zone ID at the given index
func (h *HostedZones) GetID(index int) string {
	if index >= 0 && index < len(h.zones) {
		return h.zones[index].ID
	}
	return ""
}
