package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
)

// RestAPI represents a REST API Gateway
type RestAPI struct {
	ID          string
	Name        string
	Description string
	CreatedDate string
	Version     string
}

// RestAPIs implements Resource for API Gateway REST APIs
type RestAPIs struct {
	apis []RestAPI
}

// NewRestAPIs creates a new RestAPIs resource
func NewRestAPIs() *RestAPIs {
	return &RestAPIs{
		apis: make([]RestAPI, 0),
	}
}

// Name returns the display name
func (r *RestAPIs) Name() string {
	return "API Gateway (REST)"
}

// Columns returns the column definitions
func (r *RestAPIs) Columns() []Column {
	return []Column{
		{Name: "API ID", Width: 15},
		{Name: "Name", Width: 35},
		{Name: "Version", Width: 12},
		{Name: "Created", Width: 20},
		{Name: "Description", Width: 50},
	}
}

// Fetch retrieves REST APIs from AWS
func (r *RestAPIs) Fetch(ctx context.Context, c *client.Client) error {
	r.apis = make([]RestAPI, 0)

	paginator := apigateway.NewGetRestApisPaginator(c.APIGateway(), &apigateway.GetRestApisInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list REST APIs: %w", err)
		}

		for _, api := range output.Items {
			createdDate := ""
			if api.CreatedDate != nil {
				createdDate = api.CreatedDate.Format("2006-01-02 15:04:05")
			}

			r.apis = append(r.apis, RestAPI{
				ID:          stringValue(api.Id),
				Name:        stringValue(api.Name),
				Description: stringValue(api.Description),
				CreatedDate: createdDate,
				Version:     stringValue(api.Version),
			})
		}
	}

	return nil
}

// Rows returns the table data
func (r *RestAPIs) Rows() [][]string {
	rows := make([][]string, len(r.apis))
	for i, api := range r.apis {
		rows[i] = []string{
			api.ID,
			api.Name,
			api.Version,
			api.CreatedDate,
			api.Description,
		}
	}
	return rows
}

// GetID returns the API ID at the given index
func (r *RestAPIs) GetID(index int) string {
	if index >= 0 && index < len(r.apis) {
		return r.apis[index].ID
	}
	return ""
}

// HttpAPI represents an HTTP API Gateway (v2)
type HttpAPI struct {
	ID           string
	Name         string
	Description  string
	CreatedDate  string
	ProtocolType string
}

// HttpAPIs implements Resource for API Gateway v2 HTTP APIs
type HttpAPIs struct {
	apis []HttpAPI
}

// NewHttpAPIs creates a new HttpAPIs resource
func NewHttpAPIs() *HttpAPIs {
	return &HttpAPIs{
		apis: make([]HttpAPI, 0),
	}
}

// Name returns the display name
func (h *HttpAPIs) Name() string {
	return "API Gateway (HTTP/WebSocket)"
}

// Columns returns the column definitions
func (h *HttpAPIs) Columns() []Column {
	return []Column{
		{Name: "API ID", Width: 15},
		{Name: "Name", Width: 35},
		{Name: "Protocol", Width: 12},
		{Name: "Created", Width: 20},
		{Name: "Description", Width: 50},
	}
}

// Fetch retrieves HTTP APIs from AWS
func (h *HttpAPIs) Fetch(ctx context.Context, c *client.Client) error {
	h.apis = make([]HttpAPI, 0)

	var nextToken *string
	for {
		input := &apigatewayv2.GetApisInput{}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := c.APIGatewayV2().GetApis(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to list HTTP APIs: %w", err)
		}

		for _, api := range output.Items {
			createdDate := ""
			if api.CreatedDate != nil {
				createdDate = api.CreatedDate.Format("2006-01-02 15:04:05")
			}

			h.apis = append(h.apis, HttpAPI{
				ID:           stringValue(api.ApiId),
				Name:         stringValue(api.Name),
				Description:  stringValue(api.Description),
				CreatedDate:  createdDate,
				ProtocolType: string(api.ProtocolType),
			})
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return nil
}

// Rows returns the table data
func (h *HttpAPIs) Rows() [][]string {
	rows := make([][]string, len(h.apis))
	for i, api := range h.apis {
		rows[i] = []string{
			api.ID,
			api.Name,
			api.ProtocolType,
			api.CreatedDate,
			api.Description,
		}
	}
	return rows
}

// GetID returns the API ID at the given index
func (h *HttpAPIs) GetID(index int) string {
	if index >= 0 && index < len(h.apis) {
		return h.apis[index].ID
	}
	return ""
}
