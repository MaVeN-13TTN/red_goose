package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MaVeN-13TTN/red_goose/internal/config"
	"github.com/MaVeN-13TTN/red_goose/internal/downloader"
	"github.com/MaVeN-13TTN/red_goose/internal/extractor"
	"github.com/MaVeN-13TTN/red_goose/pkg/youtube"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	outputDir  string
	quality    string
	audioOnly  bool
	playlist   bool
	verbose    bool
	maxWorkers int
	skipErrors bool

	// Version information
	version   string = "dev"
	buildTime string = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "red-goose [URL]",
	Short: "A fast YouTube video and playlist downloader",
	Long: `Red-Goose is a command-line tool built in Go for downloading 
YouTube videos and playlists efficiently and reliably.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return downloadVideo(args[0])
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	configShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showConfig()
		},
	}

	configSaveCmd = &cobra.Command{
		Use:   "save",
		Short: "Save current configuration to file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return saveConfig()
		},
	}

	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initDefaultConfig()
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Red-Goose version %s\n", version)
			fmt.Printf("Build time: %s\n", buildTime)
		},
	}

	playlistCmd = &cobra.Command{
		Use:   "playlist [URL]",
		Short: "Download entire YouTube playlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return downloadPlaylist(args[0])
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.red-goose.yaml)")

	// Root command flags
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "./downloads",
		"output directory for downloads")
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "best",
		"video quality (best, worst, 720p, 1080p, etc.)")
	rootCmd.Flags().BoolVarP(&audioOnly, "audio-only", "a", false,
		"download audio only")
	rootCmd.Flags().BoolVarP(&playlist, "playlist", "p", false,
		"download entire playlist")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"verbose output")

	// Add subcommands
	rootCmd.AddCommand(playlistCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)

	// Config subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSaveCmd)
	configCmd.AddCommand(configInitCmd)

	// Playlist command flags
	playlistCmd.Flags().StringVarP(&outputDir, "output", "o", "./downloads",
		"output directory for downloads")
	playlistCmd.Flags().StringVarP(&quality, "quality", "q", "best",
		"video quality (best, worst, 720p, 1080p, etc.)")
	playlistCmd.Flags().BoolVarP(&audioOnly, "audio-only", "a", false,
		"download audio only")
	playlistCmd.Flags().IntVarP(&maxWorkers, "workers", "w", 3,
		"number of concurrent downloads")
	playlistCmd.Flags().BoolVar(&skipErrors, "skip-errors", false,
		"continue downloading even if some videos fail")
}

var appConfig *config.Config

func initConfig() {
	var err error
	appConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		appConfig = config.DefaultConfig()
	}

	// Override config with command line flags if provided
	if rootCmd.Flags().Changed("output") {
		appConfig.Output.Directory = outputDir
	} else {
		outputDir = appConfig.Output.Directory
	}

	if rootCmd.Flags().Changed("quality") {
		appConfig.Download.DefaultQuality = quality
	} else {
		quality = appConfig.Download.DefaultQuality
	}

	if rootCmd.Flags().Changed("audio-only") {
		appConfig.Download.AudioOnly = audioOnly
	} else {
		audioOnly = appConfig.Download.AudioOnly
	}

	if rootCmd.Flags().Changed("verbose") {
		// Verbose is not part of the new config structure
		// We'll keep it as a command-line flag only
	}

	if verbose && viper.ConfigFileUsed() != "" {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func downloadVideo(url string) error {
	video, err := youtube.ParseURL(url)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	ext := extractor.New()
	details, err := ext.GetVideoDetails(video.ID)
	if err != nil {
		return fmt.Errorf("failed to extract video info: %w", err)
	}

	fmt.Printf("Title: %s\n", details.Title)
	fmt.Printf("Author: %s\n", details.Author)
	fmt.Printf("Duration: %s\n", details.Duration)
	fmt.Printf("Available formats: %d\n", len(details.Formats))

	selectedFormat, err := ext.SelectFormat(details.Formats, quality, audioOnly)
	if err != nil {
		return fmt.Errorf("failed to select format: %w", err)
	}

	fmt.Printf("Selected quality: %s\n", selectedFormat.Quality)
	fmt.Printf("File size: %d bytes\n", selectedFormat.Filesize)

	// Create filename
	var filename string
	if appConfig.Output.NamingPattern == "{title}" {
		filename = downloader.SanitizeFilename(details.Title) +
			downloader.GetFileExtension(selectedFormat.MimeType)
	} else {
		// Support for other naming patterns could be added here
		filename = downloader.SanitizeFilename(details.Title) +
			downloader.GetFileExtension(selectedFormat.MimeType)
	}

	// Download
	dl := downloader.New()
	opts := downloader.DownloadOptions{
		URL:          selectedFormat.URL,
		OutputDir:    outputDir,
		Filename:     filename,
		ShowProgress: true,
	}

	// Create context with timeout from config
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(appConfig.Network.Timeout)*time.Second)
	defer cancel()

	// Note: The downloader doesn't currently support setting user agent, retries, or rate limit
	// These would need to be added to the Downloader struct if needed

	return dl.Download(ctx, opts)
}

func downloadPlaylist(url string) error {
	if !youtube.IsPlaylistURL(url) {
		return fmt.Errorf("not a playlist URL")
	}

	playlistID := youtube.ExtractPlaylistID(url)
	if playlistID == "" {
		return fmt.Errorf("could not extract playlist ID")
	}

	ext := extractor.New()
	playlist, err := ext.GetPlaylistDetails(playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist info: %w", err)
	}

	fmt.Printf("Playlist: %s\n", playlist.Title)
	fmt.Printf("Videos: %d\n", len(playlist.Videos))
	fmt.Printf("Author: %s\n", playlist.Author)

	// Create download tasks
	var tasks []downloader.DownloadOptions

	for i, video := range playlist.Videos {
		fmt.Printf("Processing video %d/%d: %s\n", i+1, len(playlist.Videos), video.Title)

		details, err := ext.GetVideoDetails(video.ID)
		if err != nil {
			if skipErrors {
				fmt.Printf("Skipping video %s: %v\n", video.ID, err)
				continue
			}
			return fmt.Errorf("failed to get video details for %s: %w", video.ID, err)
		}

		selectedFormat, err := ext.SelectFormat(details.Formats, quality, audioOnly)
		if err != nil {
			if skipErrors {
				fmt.Printf("Skipping video %s: %v\n", video.ID, err)
				continue
			}
			return fmt.Errorf("failed to select format for %s: %w", video.ID, err)
		}

		filename := fmt.Sprintf("%03d - %s%s",
			i+1,
			downloader.SanitizeFilename(details.Title),
			downloader.GetFileExtension(selectedFormat.MimeType))

		tasks = append(tasks, downloader.DownloadOptions{
			URL:          selectedFormat.URL,
			OutputDir:    filepath.Join(outputDir, downloader.SanitizeFilename(playlist.Title)),
			Filename:     filename,
			ShowProgress: false, // Disable individual progress for batch
		})
	}

	// Download all videos
	batchDownloader := downloader.NewBatchDownloader(appConfig.Download.MaxWorkers)

	// Create context with timeout from config
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(appConfig.Network.Timeout)*time.Second)
	defer cancel()

	fmt.Printf("Starting download of %d videos with %d workers...\n",
		len(tasks), appConfig.Download.MaxWorkers)
	return batchDownloader.DownloadAll(ctx, tasks)
}

func showConfig() error {
	fmt.Println("Current configuration:")
	fmt.Printf("  Download:\n")
	fmt.Printf("    Default Quality: %s\n", appConfig.Download.DefaultQuality)
	fmt.Printf("    Max Workers: %d\n", appConfig.Download.MaxWorkers)
	fmt.Printf("    Skip Errors: %t\n", appConfig.Download.SkipErrors)
	fmt.Printf("    Audio Only: %t\n", appConfig.Download.AudioOnly)
	fmt.Printf("  Output:\n")
	fmt.Printf("    Directory: %s\n", appConfig.Output.Directory)
	fmt.Printf("    Create Subfolders: %t\n", appConfig.Output.CreateSubfolders)
	fmt.Printf("    Naming Pattern: %s\n", appConfig.Output.NamingPattern)
	fmt.Printf("  Network:\n")
	fmt.Printf("    Timeout: %d seconds\n", appConfig.Network.Timeout)
	fmt.Printf("    Retries: %d\n", appConfig.Network.Retries)
	fmt.Printf("    User Agent: %s\n", appConfig.Network.UserAgent)
	fmt.Printf("    Rate Limit: %d ms\n", appConfig.Network.RateLimit)

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("No config file in use (using defaults)")
	}

	return nil
}

func saveConfig() error {
	// Update config with current values
	appConfig.Output.Directory = outputDir
	appConfig.Download.DefaultQuality = quality
	appConfig.Download.AudioOnly = audioOnly
	appConfig.Download.MaxWorkers = maxWorkers
	// Verbose is not part of the new config structure

	if err := config.SaveConfig(appConfig, cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved successfully")
	if cfgFile != "" {
		fmt.Printf("Config file: %s\n", cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		fmt.Printf("Config file: %s\n", filepath.Join(home, ".red-goose.yaml"))
	}

	return nil
}

// SetVersionInfo sets the version and build time information
func SetVersionInfo(v, bt string) {
	version = v
	buildTime = bt
}

// GetVersionInfo returns the current version and build time
func GetVersionInfo() (string, string) {
	return version, buildTime
}

// initDefaultConfig creates a default configuration file
func initDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".red-goose.yaml")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		return nil
	}

	if err := config.CreateDefaultConfig(configPath); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("Created default configuration at %s\n", configPath)
	return nil
}
