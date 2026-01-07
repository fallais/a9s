package client

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Client wraps AWS SDK clients for various services
type Client struct {
	cfg                  aws.Config
	ec2Client            *ec2.Client
	s3Client             *s3.Client
	lambdaClient         *lambda.Client
	ecsClient            *ecs.Client
	eksClient            *eks.Client
	rdsClient            *rds.Client
	acmClient            *acm.Client
	costExplorerClient   *costexplorer.Client
	cloudfrontClient     *cloudfront.Client
	elbv2Client          *elasticloadbalancingv2.Client
	dynamodbClient       *dynamodb.Client
	secretsmanagerClient *secretsmanager.Client
	kmsClient            *kms.Client
	ecrClient            *ecr.Client
	cognitoClient        *cognitoidentityprovider.Client
	iamClient            *iam.Client
	sqsClient            *sqs.Client
	snsClient            *sns.Client
	apiGatewayClient     *apigateway.Client
	apiGatewayV2Client   *apigatewayv2.Client
	elasticacheClient    *elasticache.Client
	route53Client        *route53.Client
	region               string
	profile              string
}

// New creates a new AWS client with the default configuration
func New(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Get profile from environment variable
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}

	return &Client{
		cfg:                  cfg,
		ec2Client:            ec2.NewFromConfig(cfg),
		s3Client:             s3.NewFromConfig(cfg),
		lambdaClient:         lambda.NewFromConfig(cfg),
		ecsClient:            ecs.NewFromConfig(cfg),
		eksClient:            eks.NewFromConfig(cfg),
		rdsClient:            rds.NewFromConfig(cfg),
		acmClient:            acm.NewFromConfig(cfg),
		costExplorerClient:   costexplorer.NewFromConfig(cfg),
		cloudfrontClient:     cloudfront.NewFromConfig(cfg),
		elbv2Client:          elasticloadbalancingv2.NewFromConfig(cfg),
		dynamodbClient:       dynamodb.NewFromConfig(cfg),
		secretsmanagerClient: secretsmanager.NewFromConfig(cfg),
		kmsClient:            kms.NewFromConfig(cfg),
		ecrClient:            ecr.NewFromConfig(cfg),
		cognitoClient:        cognitoidentityprovider.NewFromConfig(cfg),
		iamClient:            iam.NewFromConfig(cfg),
		sqsClient:            sqs.NewFromConfig(cfg),
		snsClient:            sns.NewFromConfig(cfg),
		apiGatewayClient:     apigateway.NewFromConfig(cfg),
		apiGatewayV2Client:   apigatewayv2.NewFromConfig(cfg),
		elasticacheClient:    elasticache.NewFromConfig(cfg),
		route53Client:        route53.NewFromConfig(cfg),
		region:               cfg.Region,
		profile:              profile,
	}, nil
}

// NewWithRegion creates a new AWS client for a specific region
func NewWithRegion(ctx context.Context, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	// Get profile from environment variable
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}

	return &Client{
		cfg:                  cfg,
		ec2Client:            ec2.NewFromConfig(cfg),
		s3Client:             s3.NewFromConfig(cfg),
		lambdaClient:         lambda.NewFromConfig(cfg),
		ecsClient:            ecs.NewFromConfig(cfg),
		eksClient:            eks.NewFromConfig(cfg),
		rdsClient:            rds.NewFromConfig(cfg),
		acmClient:            acm.NewFromConfig(cfg),
		costExplorerClient:   costexplorer.NewFromConfig(cfg),
		cloudfrontClient:     cloudfront.NewFromConfig(cfg),
		elbv2Client:          elasticloadbalancingv2.NewFromConfig(cfg),
		dynamodbClient:       dynamodb.NewFromConfig(cfg),
		secretsmanagerClient: secretsmanager.NewFromConfig(cfg),
		kmsClient:            kms.NewFromConfig(cfg),
		ecrClient:            ecr.NewFromConfig(cfg),
		cognitoClient:        cognitoidentityprovider.NewFromConfig(cfg),
		iamClient:            iam.NewFromConfig(cfg),
		sqsClient:            sqs.NewFromConfig(cfg),
		snsClient:            sns.NewFromConfig(cfg),
		apiGatewayClient:     apigateway.NewFromConfig(cfg),
		apiGatewayV2Client:   apigatewayv2.NewFromConfig(cfg),
		elasticacheClient:    elasticache.NewFromConfig(cfg),
		route53Client:        route53.NewFromConfig(cfg),
		region:               region,
		profile:              profile,
	}, nil
}

// Region returns the current AWS region
func (c *Client) Region() string {
	return c.region
}

// Profile returns the current AWS profile
func (c *Client) Profile() string {
	return c.profile
}

// SetRegion changes the region and reinitializes clients
func (c *Client) SetRegion(ctx context.Context, region string) error {
	opts := []func(*config.LoadOptions) error{config.WithRegion(region)}
	if c.profile != "" && c.profile != "default" {
		opts = append(opts, config.WithSharedConfigProfile(c.profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return err
	}

	c.cfg = cfg
	c.ec2Client = ec2.NewFromConfig(cfg)
	c.s3Client = s3.NewFromConfig(cfg)
	c.lambdaClient = lambda.NewFromConfig(cfg)
	c.ecsClient = ecs.NewFromConfig(cfg)
	c.eksClient = eks.NewFromConfig(cfg)
	c.rdsClient = rds.NewFromConfig(cfg)
	c.acmClient = acm.NewFromConfig(cfg)
	c.costExplorerClient = costexplorer.NewFromConfig(cfg)
	c.cloudfrontClient = cloudfront.NewFromConfig(cfg)
	c.elbv2Client = elasticloadbalancingv2.NewFromConfig(cfg)
	c.dynamodbClient = dynamodb.NewFromConfig(cfg)
	c.secretsmanagerClient = secretsmanager.NewFromConfig(cfg)
	c.kmsClient = kms.NewFromConfig(cfg)
	c.ecrClient = ecr.NewFromConfig(cfg)
	c.cognitoClient = cognitoidentityprovider.NewFromConfig(cfg)
	c.iamClient = iam.NewFromConfig(cfg)
	c.sqsClient = sqs.NewFromConfig(cfg)
	c.snsClient = sns.NewFromConfig(cfg)
	c.apiGatewayClient = apigateway.NewFromConfig(cfg)
	c.apiGatewayV2Client = apigatewayv2.NewFromConfig(cfg)
	c.elasticacheClient = elasticache.NewFromConfig(cfg)
	c.route53Client = route53.NewFromConfig(cfg)
	c.region = region
	return nil
}

// SetProfile changes the profile and reinitializes clients
func (c *Client) SetProfile(ctx context.Context, profile string) error {
	opts := []func(*config.LoadOptions) error{config.WithSharedConfigProfile(profile)}
	if c.region != "" {
		opts = append(opts, config.WithRegion(c.region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return err
	}

	c.cfg = cfg
	c.ec2Client = ec2.NewFromConfig(cfg)
	c.s3Client = s3.NewFromConfig(cfg)
	c.lambdaClient = lambda.NewFromConfig(cfg)
	c.ecsClient = ecs.NewFromConfig(cfg)
	c.eksClient = eks.NewFromConfig(cfg)
	c.rdsClient = rds.NewFromConfig(cfg)
	c.acmClient = acm.NewFromConfig(cfg)
	c.costExplorerClient = costexplorer.NewFromConfig(cfg)
	c.cloudfrontClient = cloudfront.NewFromConfig(cfg)
	c.elbv2Client = elasticloadbalancingv2.NewFromConfig(cfg)
	c.dynamodbClient = dynamodb.NewFromConfig(cfg)
	c.secretsmanagerClient = secretsmanager.NewFromConfig(cfg)
	c.kmsClient = kms.NewFromConfig(cfg)
	c.ecrClient = ecr.NewFromConfig(cfg)
	c.cognitoClient = cognitoidentityprovider.NewFromConfig(cfg)
	c.iamClient = iam.NewFromConfig(cfg)
	c.sqsClient = sqs.NewFromConfig(cfg)
	c.snsClient = sns.NewFromConfig(cfg)
	c.apiGatewayClient = apigateway.NewFromConfig(cfg)
	c.apiGatewayV2Client = apigatewayv2.NewFromConfig(cfg)
	c.elasticacheClient = elasticache.NewFromConfig(cfg)
	c.route53Client = route53.NewFromConfig(cfg)
	c.profile = profile
	return nil
}

// EC2 returns the EC2 client
func (c *Client) EC2() *ec2.Client {
	return c.ec2Client
}

// S3 returns the S3 client
func (c *Client) S3() *s3.Client {
	return c.s3Client
}

// Lambda returns the Lambda client
func (c *Client) Lambda() *lambda.Client {
	return c.lambdaClient
}

// ECS returns the ECS client
func (c *Client) ECS() *ecs.Client {
	return c.ecsClient
}

// EKS returns the EKS client
func (c *Client) EKS() *eks.Client {
	return c.eksClient
}

// RDS returns the RDS client
func (c *Client) RDS() *rds.Client {
	return c.rdsClient
}

// ACM returns the ACM client
func (c *Client) ACM() *acm.Client {
	return c.acmClient
}

// CostExplorer returns the Cost Explorer client
func (c *Client) CostExplorer() *costexplorer.Client {
	return c.costExplorerClient
}

// CloudFront returns the CloudFront client
func (c *Client) CloudFront() *cloudfront.Client {
	return c.cloudfrontClient
}

// ELBv2 returns the Elastic Load Balancing v2 client
func (c *Client) ELBv2() *elasticloadbalancingv2.Client {
	return c.elbv2Client
}

// DynamoDB returns the DynamoDB client
func (c *Client) DynamoDB() *dynamodb.Client {
	return c.dynamodbClient
}

// SecretsManager returns the Secrets Manager client
func (c *Client) SecretsManager() *secretsmanager.Client {
	return c.secretsmanagerClient
}

// KMS returns the KMS client
func (c *Client) KMS() *kms.Client {
	return c.kmsClient
}

// ECR returns the ECR client
func (c *Client) ECR() *ecr.Client {
	return c.ecrClient
}

// Cognito returns the Cognito Identity Provider client
func (c *Client) Cognito() *cognitoidentityprovider.Client {
	return c.cognitoClient
}

// IAM returns the IAM client
func (c *Client) IAM() *iam.Client {
	return c.iamClient
}

// SQS returns the SQS client
func (c *Client) SQS() *sqs.Client {
	return c.sqsClient
}

// SNS returns the SNS client
func (c *Client) SNS() *sns.Client {
	return c.snsClient
}

// APIGateway returns the API Gateway client
func (c *Client) APIGateway() *apigateway.Client {
	return c.apiGatewayClient
}

// APIGatewayV2 returns the API Gateway V2 client
func (c *Client) APIGatewayV2() *apigatewayv2.Client {
	return c.apiGatewayV2Client
}

// ElastiCache returns the ElastiCache client
func (c *Client) ElastiCache() *elasticache.Client {
	return c.elasticacheClient
}

// Route53 returns the Route53 client
func (c *Client) Route53() *route53.Client {
	return c.route53Client
}
