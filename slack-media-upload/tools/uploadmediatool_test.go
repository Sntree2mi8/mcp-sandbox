package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateUploadInput(t *testing.T) {
	tests := []struct {
		name    string
		input   UploadImageInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input with all fields",
			input: UploadImageInput{
				FilePath:        "/path/to/image.png",
				ChannelID:       "C1234567890",
				ThreadTimestamp: "1234567890.123456",
			},
			wantErr: false,
		},
		{
			name: "valid input without thread_timestamp",
			input: UploadImageInput{
				FilePath:  "/path/to/image.jpg",
				ChannelID: "C1234567890",
			},
			wantErr: false,
		},
		{
			name: "missing file_path",
			input: UploadImageInput{
				ChannelID: "C1234567890",
			},
			wantErr: true,
			errMsg:  "file_path is required",
		},
		{
			name: "empty file_path",
			input: UploadImageInput{
				FilePath:  "",
				ChannelID: "C1234567890",
			},
			wantErr: true,
			errMsg:  "file_path is required",
		},
		{
			name: "missing channel_id",
			input: UploadImageInput{
				FilePath: "/path/to/image.png",
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
		{
			name: "empty channel_id",
			input: UploadImageInput{
				FilePath:  "/path/to/image.png",
				ChannelID: "",
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
		{
			name:    "missing both file_path and channel_id",
			input:   UploadImageInput{},
			wantErr: true,
			errMsg:  "file_path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUploadInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateUploadInput() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("validateUploadInput() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateUploadInput() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateImageContent(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid PNG image",
			filePath: "testdata/test_image_png.png",
			wantErr:  false,
		},
		{
			name:     "valid JPEG image",
			filePath: "testdata/test_image_jpg.jpg",
			wantErr:  false,
		},
		{
			name:     "invalid text file with .txt extension",
			filePath: "testdata/text_txt.txt",
			wantErr:  true,
		},
		{
			name:     "invalid text file with .md extension",
			filePath: "testdata/text_md.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read file content
			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("failed to read test file %s: %v", tt.filePath, err)
			}

			// Test validateImageContent
			err = validateImageContent(content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateImageContent() expected error but got nil for file %s", tt.filePath)
				}
			} else {
				if err != nil {
					t.Errorf("validateImageContent() unexpected error = %v for file %s", err, tt.filePath)
				}
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a small file (within limit)
	smallFile := filepath.Join(tempDir, "small.txt")
	smallContent := make([]byte, 1024) // 1KB
	if err := os.WriteFile(smallFile, smallContent, 0644); err != nil {
		t.Fatalf("failed to create small test file: %v", err)
	}

	// Create a large file (over limit)
	largeFile := filepath.Join(tempDir, "large.txt")
	largeContent := make([]byte, MaxFileSize+1) // Just over the limit
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("failed to create large test file: %v", err)
	}

	// Create a file exactly at the limit
	exactFile := filepath.Join(tempDir, "exact.txt")
	exactContent := make([]byte, MaxFileSize) // Exactly at the limit
	if err := os.WriteFile(exactFile, exactContent, 0644); err != nil {
		t.Fatalf("failed to create exact size test file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "small file within limit",
			filePath: smallFile,
			wantErr:  false,
		},
		{
			name:     "file exactly at limit",
			filePath: exactFile,
			wantErr:  false,
		},
		{
			name:        "file over limit",
			filePath:    largeFile,
			wantErr:     true,
			errContains: "file too large",
		},
		{
			name:        "non-existent file",
			filePath:    filepath.Join(tempDir, "nonexistent.txt"),
			wantErr:     true,
			errContains: "failed to stat file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileSize(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateFileSize() expected error but got nil for file %s", tt.filePath)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("validateFileSize() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateFileSize() unexpected error = %v for file %s", err, tt.filePath)
				}
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	// Get current working directory for testing
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	// Get absolute path to testdata directory
	testdataDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("failed to get absolute path to testdata: %v", err)
	}

	// Create a temporary directory outside of cwd for negative testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		allowedDirs []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid file in current directory",
			filePath:    filepath.Join(testdataDir, "test_image_png.png"),
			allowedDirs: []string{cwd},
			wantErr:     false,
		},
		{
			name:        "valid file in explicitly allowed directory",
			filePath:    filepath.Join(testdataDir, "test_image_png.png"),
			allowedDirs: []string{testdataDir},
			wantErr:     false,
		},
		{
			name:        "valid file with multiple allowed directories",
			filePath:    filepath.Join(testdataDir, "test_image_png.png"),
			allowedDirs: []string{tempDir, testdataDir},
			wantErr:     false,
		},
		{
			name:        "file outside allowed directory",
			filePath:    tempFile,
			allowedDirs: []string{testdataDir},
			wantErr:     true,
			errContains: "outside allowed directories",
		},
		{
			name:        "relative path should fail",
			filePath:    "testdata/test_image_png.png",
			allowedDirs: []string{cwd},
			wantErr:     true,
			errContains: "must be absolute",
		},
		{
			name:        "non-existent file should fail",
			filePath:    filepath.Join(testdataDir, "nonexistent.png"),
			allowedDirs: []string{testdataDir},
			wantErr:     true,
			errContains: "invalid file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validateFilePath
			err := validateFilePath(tt.filePath, tt.allowedDirs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateFilePath() expected error but got nil for file %s", tt.filePath)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("validateFilePath() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateFilePath() unexpected error = %v for file %s", err, tt.filePath)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
