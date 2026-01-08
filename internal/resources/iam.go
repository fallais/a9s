package resources

import (
	"context"
	"fmt"
	"strings"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// IAMUser represents an IAM user
type IAMUser struct {
	UserName   string
	UserID     string
	CreateDate string
	ARN        string
}

// IAMUsers implements Resource for IAM users
type IAMUsers struct {
	users []IAMUser
}

// NewIAMUsers creates a new IAMUsers resource
func NewIAMUsers() *IAMUsers {
	return &IAMUsers{
		users: make([]IAMUser, 0),
	}
}

// Name returns the display name
func (i *IAMUsers) Name() string {
	return "IAM Users"
}

// Columns returns the column definitions
func (i *IAMUsers) Columns() []Column {
	return []Column{
		{Name: "User Name", Width: 30},
		{Name: "User ID", Width: 25},
		{Name: "Created", Width: 20},
		{Name: "ARN", Width: 60},
	}
}

// Fetch retrieves IAM users from AWS
func (i *IAMUsers) Fetch(ctx context.Context, c *client.Client) error {
	i.users = make([]IAMUser, 0)

	paginator := iam.NewListUsersPaginator(c.IAM(), &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list IAM users: %w", err)
		}

		for _, user := range output.Users {
			createDate := ""
			if user.CreateDate != nil {
				createDate = user.CreateDate.Format("2006-01-02 15:04:05")
			}

			i.users = append(i.users, IAMUser{
				UserName:   stringValue(user.UserName),
				UserID:     stringValue(user.UserId),
				CreateDate: createDate,
				ARN:        stringValue(user.Arn),
			})
		}
	}

	return nil
}

// Rows returns the table data
func (i *IAMUsers) Rows() [][]string {
	rows := make([][]string, len(i.users))
	for idx, user := range i.users {
		rows[idx] = []string{
			user.UserName,
			user.UserID,
			user.CreateDate,
			user.ARN,
		}
	}
	return rows
}

// GetID returns the user name at the given index
func (i *IAMUsers) GetID(index int) string {
	if index >= 0 && index < len(i.users) {
		return i.users[index].UserName
	}
	return ""
}

// IAMRole represents an IAM role
type IAMRole struct {
	RoleName   string
	RoleID     string
	CreateDate string
	ARN        string
}

// IAMRoles implements Resource for IAM roles
type IAMRoles struct {
	roles []IAMRole
}

// NewIAMRoles creates a new IAMRoles resource
func NewIAMRoles() *IAMRoles {
	return &IAMRoles{
		roles: make([]IAMRole, 0),
	}
}

// Name returns the display name
func (i *IAMRoles) Name() string {
	return "IAM Roles"
}

// Columns returns the column definitions
func (i *IAMRoles) Columns() []Column {
	return []Column{
		{Name: "Role Name", Width: 40},
		{Name: "Role ID", Width: 25},
		{Name: "Created", Width: 20},
		{Name: "ARN", Width: 60},
	}
}

// Fetch retrieves IAM roles from AWS
func (i *IAMRoles) Fetch(ctx context.Context, c *client.Client) error {
	i.roles = make([]IAMRole, 0)

	paginator := iam.NewListRolesPaginator(c.IAM(), &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list IAM roles: %w", err)
		}

		for _, role := range output.Roles {
			createDate := ""
			if role.CreateDate != nil {
				createDate = role.CreateDate.Format("2006-01-02 15:04:05")
			}

			i.roles = append(i.roles, IAMRole{
				RoleName:   stringValue(role.RoleName),
				RoleID:     stringValue(role.RoleId),
				CreateDate: createDate,
				ARN:        stringValue(role.Arn),
			})
		}
	}

	return nil
}

// Rows returns the table data
func (i *IAMRoles) Rows() [][]string {
	rows := make([][]string, len(i.roles))
	for idx, role := range i.roles {
		rows[idx] = []string{
			role.RoleName,
			role.RoleID,
			role.CreateDate,
			role.ARN,
		}
	}
	return rows
}

// GetID returns the role name at the given index
func (i *IAMRoles) GetID(index int) string {
	if index >= 0 && index < len(i.roles) {
		return i.roles[index].RoleName
	}
	return ""
}

// IAMPolicy represents an IAM policy
type IAMPolicy struct {
	PolicyName      string
	PolicyID        string
	ARN             string
	AttachmentCount string
	CreateDate      string
}

// IAMPolicies implements Resource for IAM policies
type IAMPolicies struct {
	policies []IAMPolicy
}

// NewIAMPolicies creates a new IAMPolicies resource
func NewIAMPolicies() *IAMPolicies {
	return &IAMPolicies{
		policies: make([]IAMPolicy, 0),
	}
}

// Name returns the display name
func (i *IAMPolicies) Name() string {
	return "IAM Policies"
}

// Columns returns the column definitions
func (i *IAMPolicies) Columns() []Column {
	return []Column{
		{Name: "Policy Name", Width: 40},
		{Name: "Policy ID", Width: 25},
		{Name: "Attachments", Width: 12},
		{Name: "Created", Width: 20},
		{Name: "ARN", Width: 60},
	}
}

// Fetch retrieves IAM policies from AWS
func (i *IAMPolicies) Fetch(ctx context.Context, c *client.Client) error {
	i.policies = make([]IAMPolicy, 0)

	// Only fetch customer managed policies
	paginator := iam.NewListPoliciesPaginator(c.IAM(), &iam.ListPoliciesInput{
		Scope: "Local",
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list IAM policies: %w", err)
		}

		for _, policy := range output.Policies {
			createDate := ""
			if policy.CreateDate != nil {
				createDate = policy.CreateDate.Format("2006-01-02 15:04:05")
			}

			i.policies = append(i.policies, IAMPolicy{
				PolicyName:      stringValue(policy.PolicyName),
				PolicyID:        stringValue(policy.PolicyId),
				ARN:             stringValue(policy.Arn),
				AttachmentCount: fmt.Sprintf("%d", ptrInt32Value(policy.AttachmentCount)),
				CreateDate:      createDate,
			})
		}
	}

	return nil
}

// Rows returns the table data
func (i *IAMPolicies) Rows() [][]string {
	rows := make([][]string, len(i.policies))
	for idx, policy := range i.policies {
		rows[idx] = []string{
			policy.PolicyName,
			policy.PolicyID,
			policy.AttachmentCount,
			policy.CreateDate,
			policy.ARN,
		}
	}
	return rows
}

// GetID returns the policy ARN at the given index
func (i *IAMPolicies) GetID(index int) string {
	if index >= 0 && index < len(i.policies) {
		// Extract policy name from ARN for cleaner display
		parts := strings.Split(i.policies[index].ARN, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return i.policies[index].ARN
	}
	return ""
}
