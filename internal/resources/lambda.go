package resources

import (
	"context"
	"fmt"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaFunction represents a Lambda function
type LambdaFunction struct {
	FunctionName string
	Runtime      string
	Handler      string
	MemorySize   string
	Timeout      string
	LastModified string
	Description  string
}

// LambdaFunctions implements Resource for Lambda functions
type LambdaFunctions struct {
	functions []LambdaFunction
}

// NewLambdaFunctions creates a new LambdaFunctions resource
func NewLambdaFunctions() *LambdaFunctions {
	return &LambdaFunctions{
		functions: make([]LambdaFunction, 0),
	}
}

// Name returns the display name
func (l *LambdaFunctions) Name() string {
	return "Lambda Functions"
}

// Columns returns the column definitions
func (l *LambdaFunctions) Columns() []Column {
	return []Column{
		{Name: "Function Name", Width: 40},
		{Name: "Runtime", Width: 15},
		{Name: "Handler", Width: 30},
		{Name: "Memory (MB)", Width: 12},
		{Name: "Timeout (s)", Width: 12},
		{Name: "Last Modified", Width: 25},
	}
}

// Fetch retrieves Lambda functions from AWS
func (l *LambdaFunctions) Fetch(ctx context.Context, c *client.Client) error {
	l.functions = make([]LambdaFunction, 0)

	paginator := lambda.NewListFunctionsPaginator(c.Lambda(), &lambda.ListFunctionsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list Lambda functions: %w", err)
		}

		for _, fn := range output.Functions {
			l.functions = append(l.functions, LambdaFunction{
				FunctionName: stringValue(fn.FunctionName),
				Runtime:      string(fn.Runtime),
				Handler:      stringValue(fn.Handler),
				MemorySize:   fmt.Sprintf("%d", ptrInt32Value(fn.MemorySize)),
				Timeout:      fmt.Sprintf("%d", ptrInt32Value(fn.Timeout)),
				LastModified: stringValue(fn.LastModified),
				Description:  stringValue(fn.Description),
			})
		}
	}

	return nil
}

// Rows returns the table data
func (l *LambdaFunctions) Rows() [][]string {
	rows := make([][]string, len(l.functions))
	for i, fn := range l.functions {
		rows[i] = []string{
			fn.FunctionName,
			fn.Runtime,
			fn.Handler,
			fn.MemorySize,
			fn.Timeout,
			fn.LastModified,
		}
	}
	return rows
}

// GetID returns the function name at the given index
func (l *LambdaFunctions) GetID(index int) string {
	if index >= 0 && index < len(l.functions) {
		return l.functions[index].FunctionName
	}
	return ""
}
