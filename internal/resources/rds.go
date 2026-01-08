package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// RDSInstance represents an RDS database instance
type RDSInstance struct {
	DBInstanceID     string
	DBInstanceClass  string
	Engine           string
	EngineVersion    string
	Status           string
	Endpoint         string
	AvailabilityZone string
	MultiAZ          string
	StorageType      string
	AllocatedStorage string
}

// RDSInstances implements Resource for RDS instances
type RDSInstances struct {
	instances []RDSInstance
}

// NewRDSInstances creates a new RDSInstances resource
func NewRDSInstances() *RDSInstances {
	return &RDSInstances{
		instances: make([]RDSInstance, 0),
	}
}

// Name returns the display name
func (r *RDSInstances) Name() string {
	return "RDS Instances"
}

// Columns returns the column definitions
func (r *RDSInstances) Columns() []Column {
	return []Column{
		{Name: "DB Instance ID", Width: 30},
		{Name: "Class", Width: 18},
		{Name: "Engine", Width: 15},
		{Name: "Version", Width: 12},
		{Name: "Status", Width: 15},
		{Name: "Endpoint", Width: 50},
		{Name: "AZ", Width: 15},
		{Name: "Multi-AZ", Width: 10},
	}
}

// Fetch retrieves RDS instances from AWS
func (r *RDSInstances) Fetch(ctx context.Context, c *client.Client) error {
	r.instances = make([]RDSInstance, 0)

	paginator := rds.NewDescribeDBInstancesPaginator(c.RDS(), &rds.DescribeDBInstancesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe RDS instances: %w", err)
		}

		for _, db := range output.DBInstances {
			instance := RDSInstance{
				DBInstanceID:     stringValue(db.DBInstanceIdentifier),
				DBInstanceClass:  stringValue(db.DBInstanceClass),
				Engine:           stringValue(db.Engine),
				EngineVersion:    stringValue(db.EngineVersion),
				Status:           stringValue(db.DBInstanceStatus),
				AvailabilityZone: stringValue(db.AvailabilityZone),
				MultiAZ:          fmt.Sprintf("%t", ptrBoolValue(db.MultiAZ)),
				StorageType:      stringValue(db.StorageType),
				AllocatedStorage: fmt.Sprintf("%d GB", db.AllocatedStorage),
			}

			if db.Endpoint != nil {
				instance.Endpoint = fmt.Sprintf("%s:%d", stringValue(db.Endpoint.Address), db.Endpoint.Port)
			}

			r.instances = append(r.instances, instance)
		}
	}

	return nil
}

// Rows returns the table data
func (r *RDSInstances) Rows() [][]string {
	rows := make([][]string, len(r.instances))
	for i, instance := range r.instances {
		rows[i] = []string{
			instance.DBInstanceID,
			instance.DBInstanceClass,
			instance.Engine,
			instance.EngineVersion,
			instance.Status,
			instance.Endpoint,
			instance.AvailabilityZone,
			instance.MultiAZ,
		}
	}
	return rows
}

// GetID returns the DB instance ID at the given index
func (r *RDSInstances) GetID(index int) string {
	if index >= 0 && index < len(r.instances) {
		return r.instances[index].DBInstanceID
	}
	return ""
}

// QuickActions returns the available quick actions for RDS instances
func (r *RDSInstances) QuickActions() []QuickAction {
	return []QuickAction{}
}

// ptrBoolValue safely dereferences a bool pointer
func ptrBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
