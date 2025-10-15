package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sntree2mi8/mcp-sandbox/slack-media-upload/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/slack-go/slack"
)

func main() {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackBotToken == "" {
		fmt.Fprintln(os.Stderr, "SLACK_BOT_TOKEN is not set")
		os.Exit(1)
	}
	slackClient := slack.New(slackBotToken)

	// Read and validate allowed upload directories
	allowedDirsStr := os.Getenv("ALLOWED_UPLOAD_DIRS")
	if allowedDirsStr == "" {
		fmt.Fprintln(os.Stderr, "ALLOWED_UPLOAD_DIRS is not set")
		os.Exit(1)
	}

	// Split by comma and trim spaces
	var allowedDirs []string
	for _, dir := range strings.Split(allowedDirsStr, ",") {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			allowedDirs = append(allowedDirs, dir)
		}
	}

	if len(allowedDirs) == 0 {
		fmt.Fprintln(os.Stderr, "ALLOWED_UPLOAD_DIRS must contain at least one directory")
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "slack-image-upload",
		Version: "v0.0.1",
	}, nil)

	uploadImageTool := tools.NewUploadImageTool(slackClient, allowedDirs)
	mcp.AddTool[tools.UploadImageInput, tools.UploadImageOutput](
		server, uploadImageTool.Definition, uploadImageTool.UploadImage,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		if !errors.Is(err, context.Canceled) {
			// Only log the error if it's not due to context cancellation
			fmt.Fprintf(os.Stderr, "Error running MCP server: %v\n", err)
		}
	}
}
