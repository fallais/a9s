package resources

import (
	"context"
	"fmt"
	"strings"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSTopic represents an SNS topic
type SNSTopic struct {
	ARN                    string
	Name                   string
	SubscriptionsPending   string
	SubscriptionsConfirmed string
	SubscriptionsDeleted   string
}

// SNSTopics implements Resource for SNS topics
type SNSTopics struct {
	topics []SNSTopic
}

// NewSNSTopics creates a new SNSTopics resource
func NewSNSTopics() *SNSTopics {
	return &SNSTopics{
		topics: make([]SNSTopic, 0),
	}
}

// Name returns the display name
func (s *SNSTopics) Name() string {
	return "SNS Topics"
}

// Columns returns the column definitions
func (s *SNSTopics) Columns() []Column {
	return []Column{
		{Name: "Topic Name", Width: 40},
		{Name: "Confirmed", Width: 12},
		{Name: "Pending", Width: 12},
		{Name: "Deleted", Width: 12},
		{Name: "ARN", Width: 60},
	}
}

// Fetch retrieves SNS topics from AWS
func (s *SNSTopics) Fetch(ctx context.Context, c *client.Client) error {
	s.topics = make([]SNSTopic, 0)

	paginator := sns.NewListTopicsPaginator(c.SNS(), &sns.ListTopicsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list SNS topics: %w", err)
		}

		for _, topic := range output.Topics {
			arn := stringValue(topic.TopicArn)

			// Extract topic name from ARN
			// ARN format: arn:aws:sns:region:account-id:topic-name
			name := arn
			parts := strings.Split(arn, ":")
			if len(parts) >= 6 {
				name = parts[5]
			}

			t := SNSTopic{
				ARN:  arn,
				Name: name,
			}

			// Get topic attributes
			attrs, err := c.SNS().GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
				TopicArn: topic.TopicArn,
			})

			if err == nil && attrs.Attributes != nil {
				if val, ok := attrs.Attributes["SubscriptionsPending"]; ok {
					t.SubscriptionsPending = val
				}
				if val, ok := attrs.Attributes["SubscriptionsConfirmed"]; ok {
					t.SubscriptionsConfirmed = val
				}
				if val, ok := attrs.Attributes["SubscriptionsDeleted"]; ok {
					t.SubscriptionsDeleted = val
				}
			}

			s.topics = append(s.topics, t)
		}
	}

	return nil
}

// Rows returns the table data
func (s *SNSTopics) Rows() [][]string {
	rows := make([][]string, len(s.topics))
	for i, topic := range s.topics {
		rows[i] = []string{
			topic.Name,
			topic.SubscriptionsConfirmed,
			topic.SubscriptionsPending,
			topic.SubscriptionsDeleted,
			topic.ARN,
		}
	}
	return rows
}

// GetID returns the topic name at the given index
func (s *SNSTopics) GetID(index int) string {
	if index >= 0 && index < len(s.topics) {
		return s.topics[index].Name
	}
	return ""
}
