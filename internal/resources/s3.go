package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// CreateBucket creates a new S3 bucket
func (s *S3Buckets) CreateBucket(ctx context.Context, c *client.Client, bucketName string) error {
	input := &s3.CreateBucketInput{
		Bucket: &bucketName,
	}

	// For regions other than us-east-1, we need to specify the LocationConstraint
	region := c.Region()
	if region != "us-east-1" && region != "" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	_, err := c.S3().CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}

	return nil
}

// DeleteBucket deletes an S3 bucket
func (s *S3Buckets) DeleteBucket(ctx context.Context, c *client.Client, bucketName string) error {
	_, err := c.S3().DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", bucketName, err)
	}

	return nil
}

// EmptyBucket deletes all objects (including versions) from an S3 bucket
func (s *S3Buckets) EmptyBucket(ctx context.Context, c *client.Client, bucketName string) error {
	// Delete all object versions (handles versioned buckets)
	if err := s.deleteAllVersions(ctx, c, bucketName); err != nil {
		return err
	}

	// Delete remaining objects (for non-versioned buckets)
	if err := s.deleteAllObjects(ctx, c, bucketName); err != nil {
		return err
	}

	return nil
}

// deleteAllVersions deletes all object versions and delete markers
func (s *S3Buckets) deleteAllVersions(ctx context.Context, c *client.Client, bucketName string) error {
	var keyMarker *string
	var versionIDMarker *string

	for {
		listOutput, err := c.S3().ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
			Bucket:          &bucketName,
			KeyMarker:       keyMarker,
			VersionIdMarker: versionIDMarker,
		})
		if err != nil {
			return fmt.Errorf("failed to list object versions: %w", err)
		}

		// Collect objects to delete
		var objectsToDelete []s3types.ObjectIdentifier

		for _, version := range listOutput.Versions {
			objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
				Key:       version.Key,
				VersionId: version.VersionId,
			})
		}

		for _, marker := range listOutput.DeleteMarkers {
			objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
				Key:       marker.Key,
				VersionId: marker.VersionId,
			})
		}

		// Delete objects in batches of 1000
		if len(objectsToDelete) > 0 {
			if err := s.deleteBatch(ctx, c, bucketName, objectsToDelete); err != nil {
				return err
			}
		}

		// Check if there are more versions to process
		if listOutput.IsTruncated == nil || !*listOutput.IsTruncated {
			break
		}
		keyMarker = listOutput.NextKeyMarker
		versionIDMarker = listOutput.NextVersionIdMarker
	}

	return nil
}

// deleteAllObjects deletes all objects (for non-versioned buckets)
func (s *S3Buckets) deleteAllObjects(ctx context.Context, c *client.Client, bucketName string) error {
	var continuationToken *string

	for {
		listOutput, err := c.S3().ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &bucketName,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		if len(listOutput.Contents) == 0 {
			break
		}

		// Collect objects to delete
		var objectsToDelete []s3types.ObjectIdentifier
		for _, obj := range listOutput.Contents {
			objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		// Delete the batch
		if err := s.deleteBatch(ctx, c, bucketName, objectsToDelete); err != nil {
			return err
		}

		// Check if there are more objects to process
		if listOutput.IsTruncated == nil || !*listOutput.IsTruncated {
			break
		}
		continuationToken = listOutput.NextContinuationToken
	}

	return nil
}

// deleteBatch deletes a batch of objects (max 1000 per call)
func (s *S3Buckets) deleteBatch(ctx context.Context, c *client.Client, bucketName string, objects []s3types.ObjectIdentifier) error {
	// S3 DeleteObjects supports max 1000 objects per request
	const maxBatchSize = 1000

	for i := 0; i < len(objects); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(objects) {
			end = len(objects)
		}

		batch := objects[i:end]
		_, err := c.S3().DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucketName,
			Delete: &s3types.Delete{
				Objects: batch,
				Quiet:   boolPtr(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	return nil
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
