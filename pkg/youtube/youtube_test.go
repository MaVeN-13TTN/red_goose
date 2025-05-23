package youtube

import (
    "testing"
)

func TestParseURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantID  string
        wantErr bool
    }{
        {
            name:   "Standard YouTube URL",
            url:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            wantID: "dQw4w9WgXcQ",
        },
        {
            name:   "Short YouTube URL",
            url:    "https://youtu.be/dQw4w9WgXcQ",
            wantID: "dQw4w9WgXcQ",
        },
        {
            name:    "Invalid URL",
            url:     "not-a-url",
            wantErr: true,
        },
        {
            name:    "Non-YouTube URL",
            url:     "https://google.com",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            video, err := ParseURL(tt.url)
            if tt.wantErr && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            if !tt.wantErr && video.ID != tt.wantID {
                t.Errorf("Expected ID %s, got %s", tt.wantID, video.ID)
            }
        })
    }
}