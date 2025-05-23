package downloader

import (
    "context"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

func TestSanitizeFilename(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "Normal filename",
            input:    "test-file.mp4",
            expected: "test-file.mp4",
        },
        {
            name:     "Filename with invalid characters",
            input:    "test/file:with*invalid?chars.mp4",
            expected: "test_file_with_invalid_chars.mp4",
        },
        {
            name:     "Very long filename",
            input:    "very_long_filename_" + strings.Repeat("a_very_long_string_that_repeats_itself_many_times_", 10) + ".mp4",
            expected: "very_long_filename_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_a_very_long_string_that_repeats_itself_many_times_",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SanitizeFilename(tt.input)
            if result != tt.expected {
                t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}

func TestGetFileExtension(t *testing.T) {
    tests := []struct {
        name     string
        mimeType string
        expected string
    }{
        {
            name:     "MP4 video",
            mimeType: "video/mp4",
            expected: ".mp4",
        },
        {
            name:     "WebM video",
            mimeType: "video/webm",
            expected: ".webm",
        },
        {
            name:     "MP3 audio",
            mimeType: "audio/mp3",
            expected: ".mp3",
        },
        {
            name:     "M4A audio",
            mimeType: "audio/m4a",
            expected: ".m4a",
        },
        {
            name:     "Unknown type",
            mimeType: "application/octet-stream",
            expected: ".unknown",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := GetFileExtension(tt.mimeType)
            if result != tt.expected {
                t.Errorf("GetFileExtension(%q) = %q, want %q", tt.mimeType, result, tt.expected)
            }
        })
    }
}

func TestDownload(t *testing.T) {
    // Create a test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test content"))
    }))
    defer server.Close()

    // Create a temporary directory for downloads
    tempDir, err := os.MkdirTemp("", "red-goose-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Create downloader
    dl := New()

    // Test download
    opts := DownloadOptions{
        URL:          server.URL,
        OutputDir:    tempDir,
        Filename:     "test-file.txt",
        ShowProgress: false,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = dl.Download(ctx, opts)
    if err != nil {
        t.Fatalf("Download failed: %v", err)
    }

    // Verify file was downloaded
    filePath := filepath.Join(tempDir, "test-file.txt")
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        t.Fatalf("Downloaded file does not exist: %v", err)
    }

    // Verify content
    content, err := os.ReadFile(filePath)
    if err != nil {
        t.Fatalf("Failed to read downloaded file: %v", err)
    }

    if string(content) != "test content" {
        t.Errorf("Downloaded content = %q, want %q", string(content), "test content")
    }
}
