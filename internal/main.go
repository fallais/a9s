package internal

import (
	"context"
	"fmt"
	"os"

	"a9s/internal/client"
	"a9s/internal/view"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Initialize AWS client
	c, err := client.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize AWS client: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure your AWS credentials are configured.\n")
		os.Exit(1)
	}

	// Create and run the application
	app := view.New(ctx, c)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
