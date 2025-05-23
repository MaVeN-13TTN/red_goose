# Red-Goose Usage Guide

Red-Goose is a command-line tool for downloading YouTube videos and playlists. This guide explains how to use the various features of Red-Goose.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/MaVeN-13TTN/red_goose.git
cd red-goose

# Build the binary
go build -o red-goose ./cmd/red-goose

# Optionally, install it to your GOPATH
go install ./cmd/red-goose
```

## Basic Usage

### Download a Single Video

```bash
# Download a video with default settings
red-goose https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download a video with specific quality
red-goose --quality 720p https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download audio only
red-goose --audio-only https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Specify output directory
red-goose --output ~/Videos https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Download a Playlist

```bash
# Download an entire playlist
red-goose playlist https://www.youtube.com/playlist?list=PLxxx

# Download a playlist with specific quality
red-goose playlist --quality 480p https://www.youtube.com/playlist?list=PLxxx

# Download a playlist with audio only
red-goose playlist --audio-only https://www.youtube.com/playlist?list=PLxxx

# Control the number of concurrent downloads
red-goose playlist --workers 5 https://www.youtube.com/playlist?list=PLxxx

# Continue downloading even if some videos fail
red-goose playlist --skip-errors https://www.youtube.com/playlist?list=PLxxx
```

## Configuration

Red-Goose supports configuration files to set default options.

### View Current Configuration

```bash
red-goose config show
```

### Save Current Configuration

```bash
# Save current settings to config file
red-goose config save
```

### Configuration File

The default configuration file is located at `~/.red-goose.yaml`. You can specify a different configuration file using the `--config` flag.

Example configuration file:

```yaml
output_dir: "~/Videos/YouTube"
quality: "1080p"
audio_only: false
max_workers: 3
verbose: true
```

## Command-Line Options

### Global Options

- `--config`: Specify a configuration file (default is `~/.red-goose.yaml`)
- `--verbose, -v`: Enable verbose output

### Download Options

- `--output, -o`: Output directory for downloads (default is `./downloads`)
- `--quality, -q`: Video quality (best, worst, 720p, 1080p, etc.) (default is `best`)
- `--audio-only, -a`: Download audio only
- `--playlist, -p`: Download entire playlist

### Playlist Options

- `--workers, -w`: Number of concurrent downloads (default is `3`)
- `--skip-errors`: Continue downloading even if some videos fail

## Examples

### Download the Best Quality Version of a Video

```bash
red-goose --quality best https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Download Only the Audio from a Video

```bash
red-goose --audio-only https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Download an Entire Playlist with 5 Concurrent Downloads

```bash
red-goose playlist --workers 5 https://www.youtube.com/playlist?list=PLxxx
```

### Save Your Preferred Settings as Default

```bash
# Set your preferred options
red-goose --output ~/Videos --quality 720p --audio-only https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Save these settings for future use
red-goose config save
```

## Troubleshooting

If you encounter issues:

1. Make sure you have the latest version of Red-Goose
2. Try with the `--verbose` flag to see more detailed output
3. Check that the YouTube URL is valid and accessible
4. Ensure you have write permissions to the output directory