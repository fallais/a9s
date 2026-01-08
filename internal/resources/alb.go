package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// ALB represents an Application Load Balancer
type ALB struct {
	ARN               string
	Name              string
	DNSName           string
	Scheme            string
	State             string
	Type              string
	VpcID             string
	AvailabilityZones string
	CreatedTime       string
}

// ALBs implements Resource for Application Load Balancers
type ALBs struct {
	loadBalancers []ALB
}

// NewALBs creates a new ALBs resource
func NewALBs() *ALBs {
	return &ALBs{
		loadBalancers: make([]ALB, 0),
	}
}

// Name returns the display name
func (a *ALBs) Name() string {
	return "Load Balancers"
}

// Columns returns the column definitions
func (a *ALBs) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 30},
		{Name: "DNS Name", Width: 50},
		{Name: "Type", Width: 12},
		{Name: "Scheme", Width: 15},
		{Name: "State", Width: 10},
		{Name: "VPC ID", Width: 25},
		{Name: "Created", Width: 20},
	}
}

// Fetch retrieves load balancers from AWS
func (a *ALBs) Fetch(ctx context.Context, c *client.Client) error {
	a.loadBalancers = make([]ALB, 0)

	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(c.ELBv2(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe load balancers: %w", err)
		}

		for _, lb := range output.LoadBalancers {
			alb := ALB{
				ARN:     stringValue(lb.LoadBalancerArn),
				Name:    stringValue(lb.LoadBalancerName),
				DNSName: stringValue(lb.DNSName),
				Scheme:  string(lb.Scheme),
				Type:    string(lb.Type),
				VpcID:   stringValue(lb.VpcId),
			}

			if lb.State != nil {
				alb.State = string(lb.State.Code)
			}

			if lb.AvailabilityZones != nil {
				zones := ""
				for i, az := range lb.AvailabilityZones {
					if i > 0 {
						zones += ", "
					}
					zones += stringValue(az.ZoneName)
				}
				alb.AvailabilityZones = zones
			}

			if lb.CreatedTime != nil {
				alb.CreatedTime = lb.CreatedTime.Format("2006-01-02 15:04:05")
			}

			a.loadBalancers = append(a.loadBalancers, alb)
		}
	}

	return nil
}

// Rows returns the table data
func (a *ALBs) Rows() [][]string {
	rows := make([][]string, len(a.loadBalancers))
	for i, lb := range a.loadBalancers {
		rows[i] = []string{
			lb.Name,
			lb.DNSName,
			lb.Type,
			lb.Scheme,
			lb.State,
			lb.VpcID,
			lb.CreatedTime,
		}
	}
	return rows
}

// GetID returns the load balancer ARN at the given index
func (a *ALBs) GetID(index int) string {
	if index >= 0 && index < len(a.loadBalancers) {
		return a.loadBalancers[index].ARN
	}
	return ""
}

// QuickActions returns the available quick actions for ALBs
func (a *ALBs) QuickActions() []QuickAction {
	return []QuickAction{}
}
