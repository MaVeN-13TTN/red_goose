package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MaVeN-13TTN/red_goose/internal/errors"
	"github.com/MaVeN-13TTN/red_goose/internal/utils"
	"github.com/schollz/progressbar/v3"
)

type DownloadOptions struct {
	URL          string
	OutputDir    string
	Filename     string
	ShowProgress bool
}

type Downloader struct {
	client *http.Client
}

func New() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

type ProgressCallback func(downloaded, total int64, speed float64)

type ProgressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	callback   ProgressCallback
	lastUpdate time.Time
}

func NewProgressReader(reader io.Reader, total int64, callback ProgressCallback) *ProgressReader {
	return &ProgressReader{
		reader:     reader,
		total:      total,
		callback:   callback,
		lastUpdate: time.Now(),
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.downloaded, int64(n))

		// Update progress every 100ms to avoid too frequent updates
		if time.Since(pr.lastUpdate) > 100*time.Millisecond {
			downloaded := atomic.LoadInt64(&pr.downloaded)
			if pr.callback != nil {
				speed := float64(downloaded) / time.Since(pr.lastUpdate).Seconds()
				pr.callback(downloaded, pr.total, speed)
			}
			pr.lastUpdate = time.Now()
		}
	}
	return n, err
}

func (d *Downloader) Download(ctx context.Context, opts DownloadOptions) error {
	// Create output directory if it doesn't exist
	if err := utils.EnsureDir(opts.OutputDir); err != nil {
		return errors.NewFileSystemError("failed to create output directory", err)
	}

	// Retry the download operation with exponential backoff
	return utils.RetryOperation(func() error {
		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", opts.URL, nil)
		if err != nil {
			return errors.NewNetworkError("failed to create request", err)
		}

		// Set user agent to avoid blocking
		req.Header.Set("User-Agent", "red-goose/1.0")

		// Make request
		resp, err := d.client.Do(req)
		if err != nil {
			return errors.NewNetworkError("failed to download", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.NewNetworkError(fmt.Sprintf("bad status: %s", resp.Status), nil)
		}

		// Create output file
		outputPath := filepath.Join(opts.OutputDir, opts.Filename)
		file, err := os.Create(outputPath)
		if err != nil {
			return errors.NewFileSystemError("failed to create file", err)
		}
		defer file.Close()

		// Setup progress bar if requested
		var reader io.Reader = resp.Body
		if opts.ShowProgress {
			bar := progressbar.DefaultBytes(
				resp.ContentLength,
				fmt.Sprintf("Downloading %s", opts.Filename),
			)

			reader = NewProgressReader(resp.Body, resp.ContentLength, func(downloaded, total int64, speed float64) {
				bar.Set64(downloaded)
			})
		}

		// Copy response body to file
		_, err = io.Copy(file, reader)
		if err != nil {
			// If copy fails, try to remove the partial file
			os.Remove(outputPath)
			return errors.NewDownloadError("failed to save file", err)
		}

		fmt.Printf("\nDownload completed: %s\n", outputPath)
		return nil
	}, 3, 2*time.Second)
}

func (d *Downloader) DownloadResumable(ctx context.Context, opts DownloadOptions) error {
	// Create output directory if it doesn't exist
	if err := utils.EnsureDir(opts.OutputDir); err != nil {
		return errors.NewFileSystemError("failed to create output directory", err)
	}

	outputPath := filepath.Join(opts.OutputDir, opts.Filename)

	// Check if partial file exists
	var startByte int64 = 0
	if stat, err := os.Stat(outputPath); err == nil {
		startByte = stat.Size()
		fmt.Printf("Resuming download from byte %d\n", startByte)
	}

	// Retry the download operation with exponential backoff
	return utils.RetryOperation(func() error {
		// Create HTTP request with Range header
		req, err := http.NewRequestWithContext(ctx, "GET", opts.URL, nil)
		if err != nil {
			return errors.NewNetworkError("failed to create request", err)
		}

		if startByte > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
		}
		req.Header.Set("User-Agent", "red-goose/1.0")

		resp, err := d.client.Do(req)
		if err != nil {
			return errors.NewNetworkError("failed to download", err)
		}
		defer resp.Body.Close()

		// Check response status
		if startByte > 0 && resp.StatusCode != http.StatusPartialContent {
			if resp.StatusCode == http.StatusOK {
				// Server doesn't support resume, start over
				startByte = 0
			} else {
				return errors.NewNetworkError(fmt.Sprintf("bad status: %s", resp.Status), nil)
			}
		} else if startByte == 0 && resp.StatusCode != http.StatusOK {
			return errors.NewNetworkError(fmt.Sprintf("bad status: %s", resp.Status), nil)
		}

		// Open file for writing
		var file *os.File
		if startByte > 0 {
			file, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			file, err = os.Create(outputPath)
		}
		if err != nil {
			return errors.NewFileSystemError("failed to open file", err)
		}
		defer file.Close()

		// Setup progress tracking
		totalSize := startByte + resp.ContentLength
		var writer io.Writer = file

		if opts.ShowProgress {
			bar := progressbar.NewOptions64(totalSize,
				progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", opts.Filename)),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(50),
			)
			bar.Add64(startByte) // Set initial progress
			writer = io.MultiWriter(file, bar)
		}

		// Copy response body to file
		_, err = io.Copy(writer, resp.Body)
		if err != nil {
			return errors.NewDownloadError("failed to save file", err)
		}

		fmt.Printf("\nDownload completed: %s\n", outputPath)
		return nil
	}, 3, 2*time.Second)
}

func (d *Downloader) DownloadWithProgress(ctx context.Context, opts DownloadOptions, callback ProgressCallback) error {
	// Create output directory if it doesn't exist
	if err := utils.EnsureDir(opts.OutputDir); err != nil {
		return errors.NewFileSystemError("failed to create output directory", err)
	}

	// Retry the download operation with exponential backoff
	return utils.RetryOperation(func() error {
		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", opts.URL, nil)
		if err != nil {
			return errors.NewNetworkError("failed to create request", err)
		}

		// Set user agent to avoid blocking
		req.Header.Set("User-Agent", "red-goose/1.0")

		// Make request
		resp, err := d.client.Do(req)
		if err != nil {
			return errors.NewNetworkError("failed to download", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.NewNetworkError(fmt.Sprintf("bad status: %s", resp.Status), nil)
		}

		// Create output file
		outputPath := filepath.Join(opts.OutputDir, opts.Filename)
		file, err := os.Create(outputPath)
		if err != nil {
			return errors.NewFileSystemError("failed to create file", err)
		}
		defer file.Close()

		// Setup progress reader
		var reader io.Reader = resp.Body
		if callback != nil {
			reader = NewProgressReader(resp.Body, resp.ContentLength, callback)
		}

		// Copy with progress tracking
		_, err = io.Copy(file, reader)
		if err != nil {
			// If copy fails, try to remove the partial file
			os.Remove(outputPath)
			return errors.NewDownloadError("failed to save file", err)
		}

		return nil
	}, 3, 2*time.Second)
}

func SanitizeFilename(filename string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	sanitized := filename

	for _, char := range invalid {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	// Limit length
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}

	return sanitized
}

func GetFileExtension(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "mp4"):
		return ".mp4"
	case strings.Contains(mimeType, "webm"):
		return ".webm"
	case strings.Contains(mimeType, "mp3"):
		return ".mp3"
	case strings.Contains(mimeType, "m4a"):
		return ".m4a"
	default:
		return ".unknown"
	}
}

type BatchDownloader struct {
	downloader *Downloader
	maxWorkers int
	semaphore  chan struct{}
}

func NewBatchDownloader(maxWorkers int) *BatchDownloader {
	return &BatchDownloader{
		downloader: New(),
		maxWorkers: maxWorkers,
		semaphore:  make(chan struct{}, maxWorkers),
	}
}

func (bd *BatchDownloader) DownloadAll(ctx context.Context, tasks []DownloadOptions) error {
	// Create a logger
	logger := utils.NewLogger(false)

	// Setup cleanup handler
	utils.SetupCleanupHandler(func() {
		logger.Info("Canceling downloads...")
		// Nothing to clean up here, just let the goroutines finish
	})

	errChan := make(chan error, len(tasks))

	for i, task := range tasks {
		go func(idx int, opts DownloadOptions) {
			// Recover from panics in goroutines
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Recovered from panic in download task %d: %v", idx, r)
					errChan <- fmt.Errorf("task %d panicked: %v", idx, r)
				}
			}()

			bd.semaphore <- struct{}{}        // Acquire
			defer func() { <-bd.semaphore }() // Release

			err := bd.downloader.Download(ctx, opts)
			errChan <- err
		}(i, task)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(tasks); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(tasks) {
			return fmt.Errorf("all downloads failed: %v", errors[0])
		}
		return fmt.Errorf("%d of %d downloads failed", len(errors), len(tasks))
	}

	return nil
}
