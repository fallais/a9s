package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Instance represents an EC2 instance
type EC2Instance struct {
	InstanceID       string
	Name             string
	State            string
	Type             string
	PrivateIP        string
	PublicIP         string
	AvailabilityZone string
	LaunchTime       string
}

// EC2Instances implements Resource for EC2 instances
type EC2Instances struct {
	instances []EC2Instance
}

// NewEC2Instances creates a new EC2Instances resource
func NewEC2Instances() *EC2Instances {
	return &EC2Instances{
		instances: make([]EC2Instance, 0),
	}
}

// Name returns the display name
func (e *EC2Instances) Name() string {
	return "EC2 Instances"
}

// Columns returns the column definitions
func (e *EC2Instances) Columns() []Column {
	return []Column{
		{Name: "ID", Width: 20},
		{Name: "Name", Width: 30},
		{Name: "State", Width: 12},
		{Name: "Type", Width: 15},
		{Name: "Private IP", Width: 16},
		{Name: "Public IP", Width: 16},
		{Name: "AZ", Width: 15},
		{Name: "Launch Time", Width: 20},
	}
}

// Fetch retrieves EC2 instances from AWS
func (e *EC2Instances) Fetch(ctx context.Context, c *client.Client) error {
	e.instances = make([]EC2Instance, 0)

	paginator := ec2.NewDescribeInstancesPaginator(c.EC2(), &ec2.DescribeInstancesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe EC2 instances: %w", err)
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				e.instances = append(e.instances, e.parseInstance(instance))
			}
		}
	}

	return nil
}

// parseInstance converts an AWS EC2 instance to our model
func (e *EC2Instances) parseInstance(instance types.Instance) EC2Instance {
	inst := EC2Instance{
		InstanceID: stringValue(instance.InstanceId),
		State:      string(instance.State.Name),
		Type:       string(instance.InstanceType),
		PrivateIP:  stringValue(instance.PrivateIpAddress),
		PublicIP:   stringValue(instance.PublicIpAddress),
	}

	// Get the Name tag
	for _, tag := range instance.Tags {
		if stringValue(tag.Key) == "Name" {
			inst.Name = stringValue(tag.Value)
			break
		}
	}

	if instance.Placement != nil {
		inst.AvailabilityZone = stringValue(instance.Placement.AvailabilityZone)
	}

	if instance.LaunchTime != nil {
		inst.LaunchTime = instance.LaunchTime.Format("2006-01-02 15:04:05")
	}

	return inst
}

// Rows returns the table data
func (e *EC2Instances) Rows() [][]string {
	rows := make([][]string, len(e.instances))
	for i, inst := range e.instances {
		rows[i] = []string{
			inst.InstanceID,
			inst.Name,
			inst.State,
			inst.Type,
			inst.PrivateIP,
			inst.PublicIP,
			inst.AvailabilityZone,
			inst.LaunchTime,
		}
	}
	return rows
}

// GetID returns the instance ID at the given index
func (e *EC2Instances) GetID(index int) string {
	if index >= 0 && index < len(e.instances) {
		return e.instances[index].InstanceID
	}
	return ""
}

// QuickActions returns the available quick actions for EC2 instances
func (e *EC2Instances) QuickActions() []QuickAction {
	return []QuickAction{
		{
			Key:             's',
			Label:           "stop",
			Description:     "Stop instance",
			NeedsSelection:  true,
			NeedsConfirm:    true,
			ConfirmTemplate: "[red]stop[-] instance [white]%s[-]?",
			Handler:         e.StopInstance,
		},
		{
			Key:             'S',
			Label:           "start",
			Description:     "Start instance",
			NeedsSelection:  true,
			NeedsConfirm:    true,
			ConfirmTemplate: "[green]start[-] instance [white]%s[-]?",
			Handler:         e.StartInstance,
		},
		{
			Key:             'R',
			Label:           "restart",
			Description:     "Restart instance",
			NeedsSelection:  true,
			NeedsConfirm:    true,
			ConfirmTemplate: "[yellow]restart[-] instance [white]%s[-]?",
			Handler:         e.RestartInstance,
		},
	}
}

// StopInstance stops an EC2 instance
func (e *EC2Instances) StopInstance(ctx context.Context, c *client.Client, instanceID string) error {
	_, err := c.EC2().StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance %s: %w", instanceID, err)
	}
	return nil
}

// StartInstance starts an EC2 instance
func (e *EC2Instances) StartInstance(ctx context.Context, c *client.Client, instanceID string) error {
	_, err := c.EC2().StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to start instance %s: %w", instanceID, err)
	}
	return nil
}

// RestartInstance restarts (reboots) an EC2 instance
func (e *EC2Instances) RestartInstance(ctx context.Context, c *client.Client, instanceID string) error {
	_, err := c.EC2().RebootInstances(ctx, &ec2.RebootInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to restart instance %s: %w", instanceID, err)
	}
	return nil
}
