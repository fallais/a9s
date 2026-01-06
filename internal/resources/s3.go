package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Bucket represents an S3 bucket
type S3Bucket struct {
	Name         string
	CreationDate string
	Region       string
}

// S3Buckets implements Resource for S3 buckets
type S3Buckets struct {
	buckets []S3Bucket
}

// NewS3Buckets creates a new S3Buckets resource
func NewS3Buckets() *S3Buckets {
	return &S3Buckets{
		buckets: make([]S3Bucket, 0),
	}
}

// Name returns the display name
func (s *S3Buckets) Name() string {
	return "S3 Buckets"
}

// Columns returns the column definitions
func (s *S3Buckets) Columns() []Column {
	return []Column{
		{Name: "Name", Width: 50},
		{Name: "Creation Date", Width: 25},
		{Name: "Region", Width: 20},
	}
}

// Fetch retrieves S3 buckets from AWS
func (s *S3Buckets) Fetch(ctx context.Context, c *client.Client) error {
	s.buckets = make([]S3Bucket, 0)

	output, err := c.S3().ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list S3 buckets: %w", err)
	}

	for _, bucket := range output.Buckets {
		b := S3Bucket{
			Name: stringValue(bucket.Name),
		}

		if bucket.CreationDate != nil {
			b.CreationDate = bucket.CreationDate.Format("2006-01-02 15:04:05")
		}

		// Get bucket location
		location, err := c.S3().GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			if location.LocationConstraint == "" {
				b.Region = "us-east-1" // Default region when not specified
			} else {
				b.Region = string(location.LocationConstraint)
			}
		}

		s.buckets = append(s.buckets, b)
	}

	return nil
}

// Rows returns the table data
func (s *S3Buckets) Rows() [][]string {
	rows := make([][]string, len(s.buckets))
	for i, bucket := range s.buckets {
		rows[i] = []string{
			bucket.Name,
			bucket.CreationDate,
			bucket.Region,
		}
	}
	return rows
}

// GetID returns the bucket name at the given index
func (s *S3Buckets) GetID(index int) string {
	if index >= 0 && index < len(s.buckets) {
		return s.buckets[index].Name
	}
	return ""
}
