package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// VPC represents a VPC
type VPC struct {
	VpcID     string
	CIDRBlock string
	State     string
	IsDefault string
	Name      string
}

// VPCs implements Resource for VPCs
type VPCs struct {
	vpcs []VPC
}

// NewVPCs creates a new VPCs resource
func NewVPCs() *VPCs {
	return &VPCs{
		vpcs: make([]VPC, 0),
	}
}

// Name returns the display name
func (v *VPCs) Name() string {
	return "VPCs"
}

// Columns returns the column definitions
func (v *VPCs) Columns() []Column {
	return []Column{
		{Name: "VPC ID", Width: 25},
		{Name: "Name", Width: 30},
		{Name: "CIDR Block", Width: 20},
		{Name: "State", Width: 12},
		{Name: "Default", Width: 10},
	}
}

// Fetch retrieves VPCs from AWS
func (v *VPCs) Fetch(ctx context.Context, c *client.Client) error {
	v.vpcs = make([]VPC, 0)

	output, err := c.EC2().DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return fmt.Errorf("failed to describe VPCs: %w", err)
	}

	for _, vpc := range output.Vpcs {
		name := ""
		for _, tag := range vpc.Tags {
			if stringValue(tag.Key) == "Name" {
				name = stringValue(tag.Value)
				break
			}
		}

		isDefault := "No"
		if vpc.IsDefault != nil && *vpc.IsDefault {
			isDefault = "Yes"
		}

		v.vpcs = append(v.vpcs, VPC{
			VpcID:     stringValue(vpc.VpcId),
			CIDRBlock: stringValue(vpc.CidrBlock),
			State:     string(vpc.State),
			IsDefault: isDefault,
			Name:      name,
		})
	}

	return nil
}

// Rows returns the table data
func (v *VPCs) Rows() [][]string {
	rows := make([][]string, len(v.vpcs))
	for i, vpc := range v.vpcs {
		rows[i] = []string{
			vpc.VpcID,
			vpc.Name,
			vpc.CIDRBlock,
			vpc.State,
			vpc.IsDefault,
		}
	}
	return rows
}

// GetID returns the VPC ID at the given index
func (v *VPCs) GetID(index int) string {
	if index >= 0 && index < len(v.vpcs) {
		return v.vpcs[index].VpcID
	}
	return ""
}

// QuickActions returns the available quick actions for VPCs
func (v *VPCs) QuickActions() []QuickAction {
	return []QuickAction{}
}

// Subnet represents a subnet
type Subnet struct {
	SubnetID         string
	VpcID            string
	CIDRBlock        string
	AvailabilityZone string
	State            string
	Name             string
}

// Subnets implements Resource for subnets
type Subnets struct {
	subnets []Subnet
}

// NewSubnets creates a new Subnets resource
func NewSubnets() *Subnets {
	return &Subnets{
		subnets: make([]Subnet, 0),
	}
}

// Name returns the display name
func (s *Subnets) Name() string {
	return "Subnets"
}

// Columns returns the column definitions
func (s *Subnets) Columns() []Column {
	return []Column{
		{Name: "Subnet ID", Width: 25},
		{Name: "Name", Width: 30},
		{Name: "VPC ID", Width: 25},
		{Name: "CIDR Block", Width: 20},
		{Name: "AZ", Width: 15},
		{Name: "State", Width: 12},
	}
}

// Fetch retrieves subnets from AWS
func (s *Subnets) Fetch(ctx context.Context, c *client.Client) error {
	s.subnets = make([]Subnet, 0)

	output, err := c.EC2().DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return fmt.Errorf("failed to describe subnets: %w", err)
	}

	for _, subnet := range output.Subnets {
		name := ""
		for _, tag := range subnet.Tags {
			if stringValue(tag.Key) == "Name" {
				name = stringValue(tag.Value)
				break
			}
		}

		s.subnets = append(s.subnets, Subnet{
			SubnetID:         stringValue(subnet.SubnetId),
			VpcID:            stringValue(subnet.VpcId),
			CIDRBlock:        stringValue(subnet.CidrBlock),
			AvailabilityZone: stringValue(subnet.AvailabilityZone),
			State:            string(subnet.State),
			Name:             name,
		})
	}

	return nil
}

// Rows returns the table data
func (s *Subnets) Rows() [][]string {
	rows := make([][]string, len(s.subnets))
	for i, subnet := range s.subnets {
		rows[i] = []string{
			subnet.SubnetID,
			subnet.Name,
			subnet.VpcID,
			subnet.CIDRBlock,
			subnet.AvailabilityZone,
			subnet.State,
		}
	}
	return rows
}

// GetID returns the subnet ID at the given index
func (s *Subnets) GetID(index int) string {
	if index >= 0 && index < len(s.subnets) {
		return s.subnets[index].SubnetID
	}
	return ""
}

// QuickActions returns the available quick actions for subnets
func (s *Subnets) QuickActions() []QuickAction {
	return []QuickAction{}
}

// SecurityGroup represents a security group
type SecurityGroup struct {
	GroupID     string
	GroupName   string
	VpcID       string
	Description string
}

// SecurityGroups implements Resource for security groups
type SecurityGroups struct {
	groups []SecurityGroup
}

// NewSecurityGroups creates a new SecurityGroups resource
func NewSecurityGroups() *SecurityGroups {
	return &SecurityGroups{
		groups: make([]SecurityGroup, 0),
	}
}

// Name returns the display name
func (s *SecurityGroups) Name() string {
	return "Security Groups"
}

// Columns returns the column definitions
func (s *SecurityGroups) Columns() []Column {
	return []Column{
		{Name: "Group ID", Width: 25},
		{Name: "Group Name", Width: 30},
		{Name: "VPC ID", Width: 25},
		{Name: "Description", Width: 50},
	}
}

// Fetch retrieves security groups from AWS
func (s *SecurityGroups) Fetch(ctx context.Context, c *client.Client) error {
	s.groups = make([]SecurityGroup, 0)

	output, err := c.EC2().DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return fmt.Errorf("failed to describe security groups: %w", err)
	}

	for _, sg := range output.SecurityGroups {
		s.groups = append(s.groups, SecurityGroup{
			GroupID:     stringValue(sg.GroupId),
			GroupName:   stringValue(sg.GroupName),
			VpcID:       stringValue(sg.VpcId),
			Description: stringValue(sg.Description),
		})
	}

	return nil
}

// Rows returns the table data
func (s *SecurityGroups) Rows() [][]string {
	rows := make([][]string, len(s.groups))
	for i, sg := range s.groups {
		rows[i] = []string{
			sg.GroupID,
			sg.GroupName,
			sg.VpcID,
			sg.Description,
		}
	}
	return rows
}

// GetID returns the security group ID at the given index
func (s *SecurityGroups) GetID(index int) string {
	if index >= 0 && index < len(s.groups) {
		return s.groups[index].GroupID
	}
	return ""
}

// QuickActions returns the available quick actions for security groups
func (s *SecurityGroups) QuickActions() []QuickAction {
	return []QuickAction{}
}
