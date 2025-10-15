package tools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/slack-go/slack"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UploadImageTool struct {
	slackClient *slack.Client
	allowedDirs []string
	Definition  *mcp.Tool
}

func NewUploadImageTool(slackClient *slack.Client, allowedDirs []string) *UploadImageTool {
	return &UploadImageTool{
		slackClient: slackClient,
		allowedDirs: allowedDirs,
		Definition: &mcp.Tool{
			Description: "Tool for uploading images to Slack. Use this when you want to upload an image file to a Slack channel or thread.",
			Name:        "upload_image",
			Title:       "Upload Image",
		},
	}
}

type UploadImageInput struct {
	FilePath        string `json:"file_path" mcp:"description=Absolute file path of the image to upload."`
	ChannelID       string `json:"channel_id" mcp:"description=Channel ID where the image will be uploaded."`
	ThreadTimestamp string `json:"thread_timestamp,omitempty" mcp:"description=Optional timestamp of the parent message to reply in a thread."`
}

type UploadImageOutput struct {
	OK bool `json:"ok" mcp:"description=Whether the upload was successful."`
}

// validateUploadInput validates the required fields of UploadImageInput
func validateUploadInput(input UploadImageInput) error {
	if input.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}
	if input.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}
	return nil
}

// validateImageContent checks if the provided content is a valid image
func validateImageContent(content []byte) error {
	contentType := http.DetectContentType(content)
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("file is not an image: detected content type is %s", contentType)
	}
	return nil
}

// validateFilePath validates that the file path is within allowed directories
func validateFilePath(filePath string, allowedDirs []string) error {
	// Check if path is absolute
	if !filepath.IsAbs(filePath) {
		return fmt.Errorf("file_path must be absolute, got: %s", filePath)
	}

	// Resolve symlinks and clean path to prevent bypassing via symlinks
	resolved, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check each allowed directory
	for _, allowedDir := range allowedDirs {
		allowedDir = strings.TrimSpace(allowedDir)
		if allowedDir == "" {
			continue
		}

		// Resolve allowed directory symlinks as well
		resolvedAllowedDir, err := filepath.EvalSymlinks(allowedDir)
		if err != nil {
			// Skip invalid allowed directories
			continue
		}

		// Ensure allowed directory ends with separator for proper prefix matching
		if !strings.HasSuffix(resolvedAllowedDir, string(filepath.Separator)) {
			resolvedAllowedDir += string(filepath.Separator)
		}

		// Check if resolved path is within this allowed directory
		if strings.HasPrefix(resolved+string(filepath.Separator), resolvedAllowedDir) {
			return nil
		}
	}

	return fmt.Errorf("file path %s is outside allowed directories", filePath)
}

func (t *UploadImageTool) UploadImage(ctx context.Context, _ *mcp.CallToolRequest, input UploadImageInput) (
	*mcp.CallToolResult,
	UploadImageOutput,
	error,
) {
	// Validate required input
	if err := validateUploadInput(input); err != nil {
		return nil, UploadImageOutput{}, err
	}

	// Validate file path is within allowed directories
	if err := validateFilePath(input.FilePath, t.allowedDirs); err != nil {
		return nil, UploadImageOutput{}, err
	}

	// Read file content
	fileContent, err := os.ReadFile(input.FilePath)
	if err != nil {
		return nil, UploadImageOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file is an image by checking actual content
	if err := validateImageContent(fileContent); err != nil {
		return nil, UploadImageOutput{}, err
	}

	// Prepare upload parameters
	filename := filepath.Base(input.FilePath)
	params := slack.UploadFileV2Parameters{
		Filename:        filename,
		FileSize:        len(fileContent),
		Reader:          bytes.NewReader(fileContent),
		Channel:         input.ChannelID,
		ThreadTimestamp: input.ThreadTimestamp,
	}

	// Upload to Slack
	summary, err := t.slackClient.UploadFileV2Context(ctx, params)
	if err != nil {
		return nil, UploadImageOutput{OK: false}, fmt.Errorf("failed to upload file to slack: %w", err)
	}

	// Return successful result
	output := UploadImageOutput{
		OK: true,
	}

	textContent := mcp.TextContent{
		Text: fmt.Sprintf("Successfully uploaded file to Slack. File ID: %s, Title: %s", summary.ID, summary.Title),
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{&textContent},
	}

	return result, output, nil
}
