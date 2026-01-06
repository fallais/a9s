package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CognitoUserPool represents a Cognito User Pool
type CognitoUserPool struct {
	ID               string
	Name             string
	Status           string
	MFAConfiguration string
	UserCount        int
	CreationDate     string
	LastModifiedDate string
}

// CognitoUserPools implements Resource for Cognito User Pools
type CognitoUserPools struct {
	userPools []CognitoUserPool
}

// NewCognitoUserPools creates a new CognitoUserPools resource
func NewCognitoUserPools() *CognitoUserPools {
	return &CognitoUserPools{
		userPools: make([]CognitoUserPool, 0),
	}
}

// Name returns the display name
func (c *CognitoUserPools) Name() string {
	return "Cognito User Pools"
}

// Columns returns the column definitions
func (c *CognitoUserPools) Columns() []Column {
	return []Column{
		{Name: "ID", Width: 30},
		{Name: "Name", Width: 30},
		{Name: "Status", Width: 12},
		{Name: "MFA", Width: 12},
		{Name: "Users", Width: 10},
		{Name: "Created", Width: 20},
		{Name: "Modified", Width: 20},
	}
}

// Fetch retrieves Cognito User Pools from AWS
func (c *CognitoUserPools) Fetch(ctx context.Context, cl *client.Client) error {
	c.userPools = make([]CognitoUserPool, 0)

	maxResults := int32(60)
	paginator := cognitoidentityprovider.NewListUserPoolsPaginator(cl.Cognito(), &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: &maxResults,
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list Cognito user pools: %w", err)
		}

		for _, pool := range output.UserPools {
			up := CognitoUserPool{
				ID:     stringValue(pool.Id),
				Name:   stringValue(pool.Name),
				Status: string(pool.Status),
			}

			if pool.CreationDate != nil {
				up.CreationDate = pool.CreationDate.Format("2006-01-02 15:04:05")
			}

			if pool.LastModifiedDate != nil {
				up.LastModifiedDate = pool.LastModifiedDate.Format("2006-01-02 15:04:05")
			}

			// Get detailed information about the user pool
			describeOutput, err := cl.Cognito().DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: pool.Id,
			})
			if err == nil && describeOutput.UserPool != nil {
				up.MFAConfiguration = string(describeOutput.UserPool.MfaConfiguration)
				up.UserCount = int(describeOutput.UserPool.EstimatedNumberOfUsers)
			}

			c.userPools = append(c.userPools, up)
		}
	}

	return nil
}

// Rows returns the table data
func (c *CognitoUserPools) Rows() [][]string {
	rows := make([][]string, len(c.userPools))
	for i, pool := range c.userPools {
		rows[i] = []string{
			pool.ID,
			pool.Name,
			pool.Status,
			pool.MFAConfiguration,
			fmt.Sprintf("%d", pool.UserCount),
			pool.CreationDate,
			pool.LastModifiedDate,
		}
	}
	return rows
}

// GetID returns the user pool ID at the given index
func (c *CognitoUserPools) GetID(index int) string {
	if index >= 0 && index < len(c.userPools) {
		return c.userPools[index].ID
	}
	return ""
}
