# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server that extends Slack functionality with image upload capabilities. The project was created because the original Slack MCP lacks image upload functionality.

**Important**: The README notes that if the upstream Slack MCP adds this feature, that implementation should be preferred over this one.

## Architecture

The project implements an MCP tool for uploading images to Slack:

- **main.go**: Entry point that reads environment variables, creates MCP server, and registers the upload tool
- **uploadmediatool.go**: Core implementation of the `UploadImageTool` MCP tool
  - Uses the MCP Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`)
  - Uses Slack Go SDK (`github.com/slack-go/slack`)
  - Tool name: `upload_image`
  - Validates that files are images by checking actual file content (not just extension)
  - Supports uploading to channels and thread replies via optional `thread_timestamp`
- **uploadmediatool_test.go**: Unit tests for validation functions
  - Tests for `validateUploadInput` and `validateImageContent`
  - Uses test fixtures in `testdata/` directory

### Key Implementation Details

The `UploadImageTool` structure (uploadmediatool.go:15-18):
- Wraps a Slack client (`*slack.Client`)
- Exposes an MCP tool definition with name `upload_image`

Input parameters (`UploadImageInput`, uploadmediatool.go:31-35):
- `file_path` (required): Absolute file path of the image to upload
- `channel_id` (required): Channel ID where the image will be uploaded
- `thread_timestamp` (optional): Timestamp of the parent message to reply in a thread

Validation functions (pure functions, easy to test):
- `validateUploadInput()` (uploadmediatool.go:42-50): Validates required input fields
- `validateImageContent()` (uploadmediatool.go:52-59): Validates file content is an image using `http.DetectContentType()` on actual bytes, not just file extension
- `validateFilePath()` (uploadmediatool.go:61-112): Validates file path is within allowed directories to prevent Path Traversal attacks

Image validation approach:
- Uses `http.DetectContentType()` to check actual file content bytes
- More secure than extension-based validation (prevents `.jpg.txt` false positives)
- Only allows files with `image/*` content type

Security features:
- **Path Traversal Protection**: Files must be within directories specified by `ALLOWED_UPLOAD_DIRS` environment variable
  - **Required** environment variable (no default value)
  - Supports multiple directories (comma-separated)
  - Resolves symlinks using `filepath.EvalSymlinks()` to prevent bypassing
  - Rejects relative paths (only absolute paths accepted)
  - Example: `ALLOWED_UPLOAD_DIRS=/home/user/uploads,/tmp/images`

## Environment Variables

The server requires the following environment variables to be set:

- **SLACK_BOT_TOKEN** (required): Slack Bot User OAuth Token (starts with `xoxb-`)
  - Get from https://api.slack.com/apps
  - Required scopes: `files:write`, `chat:write`

- **ALLOWED_UPLOAD_DIRS** (required): Comma-separated list of absolute directory paths
  - Files can only be uploaded from these directories
  - Example: `/home/user/uploads,/tmp/images`

**Note**: This project does NOT use `.env` files. Environment variables must be set via shell export or system configuration.

## Development Commands

### Build
```bash
go build
```

### Run
```bash
# Set required environment variables first
export SLACK_BOT_TOKEN="xoxb-your-token"
export ALLOWED_UPLOAD_DIRS="$(pwd)/tools/testdata"

# Run the server
go run main.go
```

### Test
```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestValidateUploadInput
go test -v -run TestValidateImageContent
```

### Install dependencies
```bash
go mod download
```

### Add new dependencies
```bash
go get <package>
go mod tidy
```

## Testing

The project includes unit tests for validation functions in `uploadmediatool_test.go`:

### Test Data
- `testdata/test_image_png.png` - Valid PNG image for testing
- `testdata/test_image_jpg.jpg` - Valid JPEG image for testing
- `testdata/text_txt.txt` - Text file to test rejection of non-images
- `testdata/text_md.md` - Markdown file to test rejection of non-images

### Test Coverage
- `TestValidateUploadInput`: 7 test cases covering required field validation
- `TestValidateImageContent`: 4 test cases using actual test files to verify content-based image detection
- `TestValidateFilePath`: 6 test cases covering path traversal protection including:
  - Allowed directories passed as parameters
  - Explicitly allowed directories
  - Multiple allowed directories
  - Files outside allowed directories (should fail)
  - Relative paths (should fail)
  - Non-existent files (should fail)

When adding new validation logic, extract it into pure functions (like `validateUploadInput`, `validateImageContent`, and `validateFilePath`) to make testing easier without mocking Slack API or file system dependencies.

## Project Status

The project is in active development (branch: `impl_media_upload_feature`):
- ✅ Core upload tool implementation is complete in `uploadmediatool.go`
- ✅ Unit tests for validation functions are complete
- ✅ Test fixtures are available in `testdata/`
- ✅ Main entry point implemented with environment variable validation
- ✅ Path Traversal protection implemented with directory restrictions
- ✅ Installation instructions, environment setup, and examples are complete in README.md