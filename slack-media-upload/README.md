# Slack Media Upload MCP

本家slack mcpに画像アップロードの機能がないので機能を追加したもの。
本家の方にその機能が追加されたら絶対そっち使った方がいい。

## Install

```bash
go install github.com/Sntree2mi8/mcp-sandbox/slack-media-upload@latest
```

## Environment Variables

このMCPサーバーは以下の環境変数が必要です:

```bash
# Slack Bot Token (required)
# Get your token from https://api.slack.com/apps
SLACK_BOT_TOKEN="xoxb-your-bot-token-here"

# Allowed Upload Directories (required)
# Comma-separated list of absolute paths where files can be uploaded from
# Example: Single directory
ALLOWED_UPLOAD_DIRS="/home/user/uploads"

# Example: Multiple directories
ALLOWED_UPLOAD_DIRS="/home/user/uploads,/tmp/images,/var/data/media"
```

### Required Slack Permissions

Bot Token Scopesで以下の権限が必要です:
- `files:write` - ファイルのアップロード
- `chat:write` - チャンネルへの投稿

## Usage

### 1. MCPの追加

```bash
claude mcp add slack-media-upload --scope project --env SLACK_BOT_TOKEN="xoxb-your-token-here" --env ALLOWED_UPLOAD_DIRS="/path/to/allowed/directory" -- slack-media-upload
```

### 2. MCPサーバーを起動

Step1で追加しておけば、Agentが自動で起動してくれます。

### 3. Claudeから使用

Claudeに以下のようなプロンプトで画像をアップロード:

```
/path/to/image.png を Slack の #general チャンネルにアップロードしてください
```

## Development

### Build

```bash
go build
```

### Test

```bash
go test -v ./...
```

### Run locally

```bash
export SLACK_BOT_TOKEN="xoxb-your-token"
export ALLOWED_UPLOAD_DIRS="$(pwd)/tools/testdata"
go run main.go
```
