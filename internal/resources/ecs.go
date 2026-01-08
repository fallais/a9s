package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ECSCluster represents an ECS cluster
type ECSCluster struct {
	ClusterName             string
	Status                  string
	RunningTasksCount       string
	PendingTasksCount       string
	ActiveServicesCount     string
	RegisteredContainerInst string
}

// ECSClusters implements Resource for ECS clusters
type ECSClusters struct {
	clusters []ECSCluster
}

// NewECSClusters creates a new ECSClusters resource
func NewECSClusters() *ECSClusters {
	return &ECSClusters{
		clusters: make([]ECSCluster, 0),
	}
}

// Name returns the display name
func (e *ECSClusters) Name() string {
	return "ECS Clusters"
}

// Columns returns the column definitions
func (e *ECSClusters) Columns() []Column {
	return []Column{
		{Name: "Cluster Name", Width: 35},
		{Name: "Status", Width: 12},
		{Name: "Running Tasks", Width: 14},
		{Name: "Pending Tasks", Width: 14},
		{Name: "Services", Width: 10},
		{Name: "Instances", Width: 10},
	}
}

// Fetch retrieves ECS clusters from AWS
func (e *ECSClusters) Fetch(ctx context.Context, c *client.Client) error {
	e.clusters = make([]ECSCluster, 0)

	// First, list all cluster ARNs
	listOutput, err := c.ECS().ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return fmt.Errorf("failed to list ECS clusters: %w", err)
	}

	if len(listOutput.ClusterArns) == 0 {
		return nil
	}

	// Then describe the clusters to get details
	describeOutput, err := c.ECS().DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: listOutput.ClusterArns,
	})
	if err != nil {
		return fmt.Errorf("failed to describe ECS clusters: %w", err)
	}

	for _, cluster := range describeOutput.Clusters {
		e.clusters = append(e.clusters, ECSCluster{
			ClusterName:             stringValue(cluster.ClusterName),
			Status:                  stringValue(cluster.Status),
			RunningTasksCount:       fmt.Sprintf("%d", cluster.RunningTasksCount),
			PendingTasksCount:       fmt.Sprintf("%d", cluster.PendingTasksCount),
			ActiveServicesCount:     fmt.Sprintf("%d", cluster.ActiveServicesCount),
			RegisteredContainerInst: fmt.Sprintf("%d", cluster.RegisteredContainerInstancesCount),
		})
	}

	return nil
}

// Rows returns the table data
func (e *ECSClusters) Rows() [][]string {
	rows := make([][]string, len(e.clusters))
	for i, cluster := range e.clusters {
		rows[i] = []string{
			cluster.ClusterName,
			cluster.Status,
			cluster.RunningTasksCount,
			cluster.PendingTasksCount,
			cluster.ActiveServicesCount,
			cluster.RegisteredContainerInst,
		}
	}
	return rows
}

// GetID returns the cluster name at the given index
func (e *ECSClusters) GetID(index int) string {
	if index >= 0 && index < len(e.clusters) {
		return e.clusters[index].ClusterName
	}
	return ""
}

// QuickActions returns the available quick actions for ECS clusters
func (e *ECSClusters) QuickActions() []QuickAction {
	return []QuickAction{}
}
