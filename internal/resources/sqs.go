package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSQueue represents an SQS queue
type SQSQueue struct {
	URL                          string
	Name                         string
	ApproximateMessages          string
	ApproximateMessagesNotVisible string
	MessageRetentionPeriod       string
}

// SQSQueues implements Resource for SQS queues
type SQSQueues struct {
	queues []SQSQueue
}

// NewSQSQueues creates a new SQSQueues resource
func NewSQSQueues() *SQSQueues {
	return &SQSQueues{
		queues: make([]SQSQueue, 0),
	}
}

// Name returns the display name
func (s *SQSQueues) Name() string {
	return "SQS Queues"
}

// Columns returns the column definitions
func (s *SQSQueues) Columns() []Column {
	return []Column{
		{Name: "Queue Name", Width: 40},
		{Name: "Messages", Width: 12},
		{Name: "In Flight", Width: 12},
		{Name: "Retention (s)", Width: 15},
		{Name: "URL", Width: 60},
	}
}

// Fetch retrieves SQS queues from AWS
func (s *SQSQueues) Fetch(ctx context.Context, c *client.Client) error {
	s.queues = make([]SQSQueue, 0)

	paginator := sqs.NewListQueuesPaginator(c.SQS(), &sqs.ListQueuesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list SQS queues: %w", err)
		}

		for _, url := range output.QueueUrls {
			// Get queue attributes
			attrs, err := c.SQS().GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
				QueueUrl: &url,
				AttributeNames: []sqstypes.QueueAttributeName{
					sqstypes.QueueAttributeNameQueueArn,
					sqstypes.QueueAttributeNameApproximateNumberOfMessages,
					sqstypes.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
					sqstypes.QueueAttributeNameMessageRetentionPeriod,
				},
			})

			queue := SQSQueue{
				URL: url,
			}

			// Extract queue name from URL
			// URL format: https://sqs.region.amazonaws.com/account-id/queue-name
		for i := len(url) - 1; i >= 0; i-- {
				if url[i] == '/' {
					queue.Name = url[i+1:]
					break
				}
			}
			if queue.Name == "" {
				queue.Name = url
			}

			if err == nil && attrs.Attributes != nil {
				if val, ok := attrs.Attributes["ApproximateNumberOfMessages"]; ok {
					queue.ApproximateMessages = val
				}
				if val, ok := attrs.Attributes["ApproximateNumberOfMessagesNotVisible"]; ok {
					queue.ApproximateMessagesNotVisible = val
				}
				if val, ok := attrs.Attributes["MessageRetentionPeriod"]; ok {
					queue.MessageRetentionPeriod = val
				}
			}

			s.queues = append(s.queues, queue)
		}
	}

	return nil
}

// Rows returns the table data
func (s *SQSQueues) Rows() [][]string {
	rows := make([][]string, len(s.queues))
	for i, queue := range s.queues {
		rows[i] = []string{
			queue.Name,
			queue.ApproximateMessages,
			queue.ApproximateMessagesNotVisible,
			queue.MessageRetentionPeriod,
			queue.URL,
		}
	}
	return rows
}

// GetID returns the queue name at the given index
func (s *SQSQueues) GetID(index int) string {
	if index >= 0 && index < len(s.queues) {
		return s.queues[index].Name
	}
	return ""
}
