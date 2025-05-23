package youtube

import (
    "fmt"
    "net/url"
    "regexp"
    "strings"
)

type VideoType int

const (
    VideoTypeSingle VideoType = iota
    VideoTypePlaylist
    VideoTypeChannel
)

type VideoInfo struct {
    ID       string
    URL      string
    Type     VideoType
    Title    string
    Duration string
}

type PlaylistInfo struct {
    ID     string
    URL    string
    Title  string
    Videos []VideoInfo
}

var (
    videoIDRegex    = regexp.MustCompile(`(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})`)
    playlistIDRegex = regexp.MustCompile(`[?&]list=([a-zA-Z0-9_-]+)`)
)

func ParseURL(inputURL string) (*VideoInfo, error) {
    parsedURL, err := url.Parse(inputURL)
    if err != nil {
        return nil, fmt.Errorf("invalid URL: %w", err)
    }
    
    // Check if it's a YouTube URL
    if !strings.Contains(parsedURL.Host, "youtube.com") && 
       !strings.Contains(parsedURL.Host, "youtu.be") {
        return nil, fmt.Errorf("not a YouTube URL")
    }
    
    video := &VideoInfo{
        URL: inputURL,
    }
    
    // Extract video ID
    matches := videoIDRegex.FindStringSubmatch(inputURL)
    if len(matches) > 1 {
        video.ID = matches[1]
        video.Type = VideoTypeSingle
    }
    
    // Check for playlist
    playlistMatches := playlistIDRegex.FindStringSubmatch(inputURL)
    if len(playlistMatches) > 1 {
        video.Type = VideoTypePlaylist
    }
    
    if video.ID == "" && video.Type != VideoTypePlaylist {
        return nil, fmt.Errorf("could not extract video ID from URL")
    }
    
    return video, nil
}

func IsPlaylistURL(url string) bool {
    return playlistIDRegex.MatchString(url)
}

func ExtractVideoID(url string) string {
    matches := videoIDRegex.FindStringSubmatch(url)
    if len(matches) > 1 {
        return matches[1]
    }
    return ""
}

func ExtractPlaylistID(url string) string {
    matches := playlistIDRegex.FindStringSubmatch(url)
    if len(matches) > 1 {
        return matches[1]
    }
    return ""
}