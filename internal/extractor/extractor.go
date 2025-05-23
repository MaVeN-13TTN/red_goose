package extractor

import (
    "fmt"
    "sort"

    "github.com/kkdai/youtube/v2"
)

type VideoDetails struct {
    ID          string
    Title       string
    Author      string
    Duration    string
    Description string
    Thumbnail   string
    Formats     []FormatInfo
}

type FormatInfo struct {
    Quality   string
    MimeType  string
    URL       string
    Filesize  int64
    AudioOnly bool
    VideoOnly bool
}

type Extractor struct {
    client *youtube.Client
}

func New() *Extractor {
    return &Extractor{
        client: &youtube.Client{},
    }
}

func (e *Extractor) GetVideoDetails(videoID string) (*VideoDetails, error) {
    video, err := e.client.GetVideo(videoID)
    if err != nil {
        return nil, fmt.Errorf("failed to get video info: %w", err)
    }

    details := &VideoDetails{
        ID:          video.ID,
        Title:       video.Title,
        Author:      video.Author,
        Duration:    video.Duration.String(),
        Description: video.Description,
    }

    // Extract thumbnail
    if len(video.Thumbnails) > 0 {
        details.Thumbnail = video.Thumbnails[0].URL
    }

    // Process formats
    for _, format := range video.Formats {
        formatInfo := FormatInfo{
            Quality:   format.Quality,
            MimeType:  format.MimeType,
            URL:       format.URL,
            Filesize:  format.ContentLength,
            AudioOnly: format.AudioChannels > 0 && format.Width == 0,
            VideoOnly: format.AudioChannels == 0 && format.Width > 0,
        }
        details.Formats = append(details.Formats, formatInfo)
    }

    // Sort formats by quality (best first)
    sort.Slice(details.Formats, func(i, j int) bool {
        return details.Formats[i].Filesize > details.Formats[j].Filesize
    })

    return details, nil
}

func (e *Extractor) GetPlaylistDetails(playlistID string) (*youtube.Playlist, error) {
    playlist, err := e.client.GetPlaylist(playlistID)
    if err != nil {
        return nil, fmt.Errorf("failed to get playlist info: %w", err)
    }

    return playlist, nil
}

func (e *Extractor) SelectFormat(formats []FormatInfo, quality string, audioOnly bool) (*FormatInfo, error) {
    if len(formats) == 0 {
        return nil, fmt.Errorf("no formats available")
    }

    if audioOnly {
        for _, format := range formats {
            if format.AudioOnly {
                return &format, nil
            }
        }
        return nil, fmt.Errorf("no audio-only format found")
    }

    switch quality {
    case "best":
        return &formats[0], nil
    case "worst":
        return &formats[len(formats)-1], nil
    default:
        // Try to find specific quality
        for _, format := range formats {
            if format.Quality == quality {
                return &format, nil
            }
        }
        // Fallback to best
        return &formats[0], nil
    }
}
