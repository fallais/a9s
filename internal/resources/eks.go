package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/eks"
)

// EKSCluster represents an EKS cluster
type EKSCluster struct {
	Name            string
	Status          string
	Version         string
	Endpoint        string
	RoleArn         string
	CreatedAt       string
	PlatformVersion string
}

// EKSClusters implements Resource for EKS clusters
type EKSClusters struct {
	clusters []EKSCluster
}

// NewEKSClusters creates a new EKSClusters resource
func NewEKSClusters() *EKSClusters {
	return &EKSClusters{
		clusters: make([]EKSCluster, 0),
	}
}

// Name returns the display name
func (e *EKSClusters) Name() string {
	return "EKS Clusters"
}

// Columns returns the column definitions
func (e *EKSClusters) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 30},
		{Name: "Status", Width: 12},
		{Name: "Version", Width: 10},
		{Name: "Platform Version", Width: 18},
		{Name: "Created At", Width: 20},
	}
}

// Fetch retrieves EKS clusters from AWS
func (e *EKSClusters) Fetch(ctx context.Context, c *client.Client) error {
	e.clusters = make([]EKSCluster, 0)

	// First, list all cluster names
	paginator := eks.NewListClustersPaginator(c.EKS(), &eks.ListClustersInput{})

	var clusterNames []string
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list EKS clusters: %w", err)
		}
		clusterNames = append(clusterNames, output.Clusters...)
	}

	// Then describe each cluster to get details
	for _, name := range clusterNames {
		describeOutput, err := c.EKS().DescribeCluster(ctx, &eks.DescribeClusterInput{
			Name: &name,
		})
		if err != nil {
			continue // Skip clusters we can't describe
		}

		cluster := describeOutput.Cluster
		eksCluster := EKSCluster{
			Name:            stringValue(cluster.Name),
			Status:          string(cluster.Status),
			Version:         stringValue(cluster.Version),
			Endpoint:        stringValue(cluster.Endpoint),
			RoleArn:         stringValue(cluster.RoleArn),
			PlatformVersion: stringValue(cluster.PlatformVersion),
		}

		if cluster.CreatedAt != nil {
			eksCluster.CreatedAt = cluster.CreatedAt.Format("2006-01-02 15:04:05")
		}

		e.clusters = append(e.clusters, eksCluster)
	}

	return nil
}

// Rows returns the table data
func (e *EKSClusters) Rows() [][]string {
	rows := make([][]string, len(e.clusters))
	for i, cluster := range e.clusters {
		rows[i] = []string{
			cluster.Name,
			cluster.Status,
			cluster.Version,
			cluster.PlatformVersion,
			cluster.CreatedAt,
		}
	}
	return rows
}

// GetID returns the cluster name at the given index
func (e *EKSClusters) GetID(index int) string {
	if index >= 0 && index < len(e.clusters) {
		return e.clusters[index].Name
	}
	return ""
}
