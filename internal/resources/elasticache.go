package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

// ElastiCacheCluster represents an ElastiCache cluster
type ElastiCacheCluster struct {
	ClusterID     string
	Engine        string
	EngineVersion string
	CacheNodeType string
	NumCacheNodes string
	Status        string
	PreferredAZ   string
}

// ElastiCacheClusters implements Resource for ElastiCache clusters
type ElastiCacheClusters struct {
	clusters []ElastiCacheCluster
}

// NewElastiCacheClusters creates a new ElastiCacheClusters resource
func NewElastiCacheClusters() *ElastiCacheClusters {
	return &ElastiCacheClusters{
		clusters: make([]ElastiCacheCluster, 0),
	}
}

// Name returns the display name
func (e *ElastiCacheClusters) Name() string {
	return "ElastiCache Clusters"
}

// Columns returns the column definitions
func (e *ElastiCacheClusters) Columns() []Column {
	return []Column{
		{Name: "Cluster ID", Width: 30},
		{Name: "Engine", Width: 12},
		{Name: "Version", Width: 10},
		{Name: "Node Type", Width: 18},
		{Name: "Nodes", Width: 8},
		{Name: "Status", Width: 15},
		{Name: "AZ", Width: 15},
	}
}

// Fetch retrieves ElastiCache clusters from AWS
func (e *ElastiCacheClusters) Fetch(ctx context.Context, c *client.Client) error {
	e.clusters = make([]ElastiCacheCluster, 0)

	paginator := elasticache.NewDescribeCacheClustersPaginator(c.ElastiCache(), &elasticache.DescribeCacheClustersInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe ElastiCache clusters: %w", err)
		}

		for _, cluster := range output.CacheClusters {
			e.clusters = append(e.clusters, ElastiCacheCluster{
				ClusterID:     stringValue(cluster.CacheClusterId),
				Engine:        stringValue(cluster.Engine),
				EngineVersion: stringValue(cluster.EngineVersion),
				CacheNodeType: stringValue(cluster.CacheNodeType),
				NumCacheNodes: fmt.Sprintf("%d", ptrInt32Value(cluster.NumCacheNodes)),
				Status:        stringValue(cluster.CacheClusterStatus),
				PreferredAZ:   stringValue(cluster.PreferredAvailabilityZone),
			})
		}
	}

	return nil
}

// Rows returns the table data
func (e *ElastiCacheClusters) Rows() [][]string {
	rows := make([][]string, len(e.clusters))
	for i, cluster := range e.clusters {
		rows[i] = []string{
			cluster.ClusterID,
			cluster.Engine,
			cluster.EngineVersion,
			cluster.CacheNodeType,
			cluster.NumCacheNodes,
			cluster.Status,
			cluster.PreferredAZ,
		}
	}
	return rows
}

// GetID returns the cluster ID at the given index
func (e *ElastiCacheClusters) GetID(index int) string {
	if index >= 0 && index < len(e.clusters) {
		return e.clusters[index].ClusterID
	}
	return ""
}

// ElastiCacheReplicationGroup represents an ElastiCache replication group
type ElastiCacheReplicationGroup struct {
	ReplicationGroupID string
	Description        string
	Status             string
	ClusterEnabled     string
	NodeType           string
	NumNodeGroups      string
}

// ElastiCacheReplicationGroups implements Resource for ElastiCache replication groups
type ElastiCacheReplicationGroups struct {
	groups []ElastiCacheReplicationGroup
}

// NewElastiCacheReplicationGroups creates a new ElastiCacheReplicationGroups resource
func NewElastiCacheReplicationGroups() *ElastiCacheReplicationGroups {
	return &ElastiCacheReplicationGroups{
		groups: make([]ElastiCacheReplicationGroup, 0),
	}
}

// Name returns the display name
func (e *ElastiCacheReplicationGroups) Name() string {
	return "ElastiCache Replication Groups"
}

// Columns returns the column definitions
func (e *ElastiCacheReplicationGroups) Columns() []Column {
	return []Column{
		{Name: "Replication Group ID", Width: 30},
		{Name: "Description", Width: 40},
		{Name: "Status", Width: 15},
		{Name: "Node Type", Width: 18},
		{Name: "Node Groups", Width: 12},
		{Name: "Cluster Mode", Width: 12},
	}
}

// Fetch retrieves ElastiCache replication groups from AWS
func (e *ElastiCacheReplicationGroups) Fetch(ctx context.Context, c *client.Client) error {
	e.groups = make([]ElastiCacheReplicationGroup, 0)

	paginator := elasticache.NewDescribeReplicationGroupsPaginator(c.ElastiCache(), &elasticache.DescribeReplicationGroupsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe ElastiCache replication groups: %w", err)
		}

		for _, rg := range output.ReplicationGroups {
			clusterEnabled := "No"
			if rg.ClusterEnabled != nil && *rg.ClusterEnabled {
				clusterEnabled = "Yes"
			}

			numNodeGroups := "0"
			if len(rg.NodeGroups) > 0 {
				numNodeGroups = fmt.Sprintf("%d", len(rg.NodeGroups))
			}

			// Get node type from member clusters if available
			nodeType := ""
			if len(rg.MemberClusters) > 0 {
				nodeType = stringValue(rg.CacheNodeType)
			}

			e.groups = append(e.groups, ElastiCacheReplicationGroup{
				ReplicationGroupID: stringValue(rg.ReplicationGroupId),
				Description:        stringValue(rg.Description),
				Status:             stringValue(rg.Status),
				ClusterEnabled:     clusterEnabled,
				NodeType:           nodeType,
				NumNodeGroups:      numNodeGroups,
			})
		}
	}

	return nil
}

// Rows returns the table data
func (e *ElastiCacheReplicationGroups) Rows() [][]string {
	rows := make([][]string, len(e.groups))
	for i, rg := range e.groups {
		rows[i] = []string{
			rg.ReplicationGroupID,
			rg.Description,
			rg.Status,
			rg.NodeType,
			rg.NumNodeGroups,
			rg.ClusterEnabled,
		}
	}
	return rows
}

// GetID returns the replication group ID at the given index
func (e *ElastiCacheReplicationGroups) GetID(index int) string {
	if index >= 0 && index < len(e.groups) {
		return e.groups[index].ReplicationGroupID
	}
	return ""
}
