package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// ECRRepository represents an ECR repository
type ECRRepository struct {
	Name           string
	URI            string
	ImageCount     int
	ScanOnPush     bool
	TagMutability  string
	EncryptionType string
	CreatedAt      string
}

// ECRRepositories implements Resource for ECR repositories
type ECRRepositories struct {
	repositories []ECRRepository
}

// NewECRRepositories creates a new ECRRepositories resource
func NewECRRepositories() *ECRRepositories {
	return &ECRRepositories{
		repositories: make([]ECRRepository, 0),
	}
}

// Name returns the display name
func (e *ECRRepositories) Name() string {
	return "ECR Repositories"
}

// Columns returns the column definitions
func (e *ECRRepositories) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 35},
		{Name: "URI", Width: 60},
		{Name: "Images", Width: 8},
		{Name: "Scan", Width: 6},
		{Name: "Tag Mutability", Width: 15},
		{Name: "Encryption", Width: 12},
	}
}

// Fetch retrieves ECR repositories from AWS
func (e *ECRRepositories) Fetch(ctx context.Context, c *client.Client) error {
	e.repositories = make([]ECRRepository, 0)

	paginator := ecr.NewDescribeRepositoriesPaginator(c.ECR(), &ecr.DescribeRepositoriesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe ECR repositories: %w", err)
		}

		for _, repo := range output.Repositories {
			r := ECRRepository{
				Name:          stringValue(repo.RepositoryName),
				URI:           stringValue(repo.RepositoryUri),
				TagMutability: string(repo.ImageTagMutability),
			}

			if repo.ImageScanningConfiguration != nil {
				r.ScanOnPush = repo.ImageScanningConfiguration.ScanOnPush
			}

			if repo.EncryptionConfiguration != nil {
				r.EncryptionType = string(repo.EncryptionConfiguration.EncryptionType)
			}

			if repo.CreatedAt != nil {
				r.CreatedAt = repo.CreatedAt.Format("2006-01-02 15:04:05")
			}

			// Get image count
			imagesOutput, err := c.ECR().DescribeImages(ctx, &ecr.DescribeImagesInput{
				RepositoryName: repo.RepositoryName,
			})
			if err == nil {
				r.ImageCount = len(imagesOutput.ImageDetails)
			}

			e.repositories = append(e.repositories, r)
		}
	}

	return nil
}

// Rows returns the table data
func (e *ECRRepositories) Rows() [][]string {
	rows := make([][]string, len(e.repositories))
	for i, repo := range e.repositories {
		scan := "No"
		if repo.ScanOnPush {
			scan = "Yes"
		}
		rows[i] = []string{
			repo.Name,
			repo.URI,
			fmt.Sprintf("%d", repo.ImageCount),
			scan,
			repo.TagMutability,
			repo.EncryptionType,
		}
	}
	return rows
}

// GetID returns the repository name at the given index
func (e *ECRRepositories) GetID(index int) string {
	if index >= 0 && index < len(e.repositories) {
		return e.repositories[index].Name
	}
	return ""
}

// QuickActions returns the available quick actions for ECR repositories
func (e *ECRRepositories) QuickActions() []QuickAction {
	return []QuickAction{}
}
