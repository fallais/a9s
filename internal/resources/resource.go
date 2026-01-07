package resources

import (
	"context"

	"a9s/internal/client"
)

// Column represents a table column definition
type Column struct {
	Name  string
	Width int
}

// Resource defines the interface for all AWS resources
type Resource interface {
	// Name returns the display name of the resource type
	Name() string

	// Columns returns the column definitions for the table
	Columns() []Column

	// Fetch retrieves the resources from AWS
	Fetch(ctx context.Context, client *client.Client) error

	// Rows returns the data rows for the table
	Rows() [][]string

	// GetID returns the ID of the resource at the given index
	GetID(index int) string
}

// Registry holds all available resource types
type Registry struct {
	resources map[string]Resource
}

// NewRegistry creates a new resource registry
func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[string]Resource),
	}
}

// Register adds a resource to the registry
func (r *Registry) Register(key string, resource Resource) {
	r.resources[key] = resource
}

// Get returns a resource by key
func (r *Registry) Get(key string) (Resource, bool) {
	res, ok := r.resources[key]
	return res, ok
}

// List returns all registered resource keys
func (r *Registry) List() []string {
	keys := make([]string, 0, len(r.resources))
	for k := range r.resources {
		keys = append(keys, k)
	}
	return keys
}

// DefaultRegistry creates a registry with all default resources
func DefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.Register("ec2", NewEC2Instances())
	reg.Register("s3", NewS3Buckets())
	reg.Register("lambda", NewLambdaFunctions())
	reg.Register("ecs", NewECSClusters())
	reg.Register("eks", NewEKSClusters())
	reg.Register("rds", NewRDSInstances())
	reg.Register("acm", NewACMCertificates())
	reg.Register("billing", NewBilling())
	reg.Register("cloudfront", NewCloudFrontDistributions())
	reg.Register("alb", NewALBs())
	reg.Register("dynamodb", NewDynamoDBTables())
	reg.Register("secrets", NewSecrets())
	reg.Register("kms", NewKMSKeys())
	reg.Register("ecr", NewECRRepositories())
	reg.Register("cognito", NewCognitoUserPools())
	reg.Register("iam-users", NewIAMUsers())
	reg.Register("iam-roles", NewIAMRoles())
	reg.Register("iam-policies", NewIAMPolicies())
	reg.Register("vpc", NewVPCs())
	reg.Register("subnets", NewSubnets())
	reg.Register("security-groups", NewSecurityGroups())
	reg.Register("sqs", NewSQSQueues())
	reg.Register("sns", NewSNSTopics())
	reg.Register("api-gateway", NewRestAPIs())
	reg.Register("api-gateway-v2", NewHttpAPIs())
	reg.Register("elasticache-clusters", NewElastiCacheClusters())
	reg.Register("elasticache-groups", NewElastiCacheReplicationGroups())
	reg.Register("route53", NewHostedZones())
	return reg
}
