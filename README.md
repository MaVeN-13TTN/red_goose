# Red-Goose YouTube Downloader

A fast and efficient command-line tool built in Go for downloading YouTube videos and playlists.

## Features

- Download single videos or entire playlists
- Select video quality
- Download audio-only
- Track download progress
- Concurrent downloads for playlists

## Usage

```bash
# Download a single video
red-goose https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download a video with specific quality
red-goose --quality 720p https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download audio only
red-goose --audio-only https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download a playlist
red-goose playlist https://www.youtube.com/playlist?list=PLxxx
```

## Installation

```bash
go install github.com/MaVeN-13TTN/red_goose@latest
```

## License

MIT