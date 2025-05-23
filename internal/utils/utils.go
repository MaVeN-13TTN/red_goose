package utils

import (
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "runtime/debug"
    "syscall"
    "time"
)

// Logger provides a simple logging interface
type Logger struct {
    Verbose bool
}

// NewLogger creates a new logger
func NewLogger(verbose bool) *Logger {
    return &Logger{
        Verbose: verbose,
    }
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
    fmt.Printf(format+"\n", args...)
}

// Debug logs a debug message if verbose mode is enabled
func (l *Logger) Debug(format string, args ...interface{}) {
    if l.Verbose {
        fmt.Printf("[DEBUG] "+format+"\n", args...)
    }
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
}

// Fatal logs an error message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
    l.Error(format, args...)
    os.Exit(1)
}

// SetupCleanupHandler sets up a handler for SIGINT and SIGTERM signals
// to perform cleanup before exiting
func SetupCleanupHandler(cleanup func()) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        fmt.Println("\nReceived termination signal. Cleaning up...")
        cleanup()
        os.Exit(1)
    }()
}

// RecoverFromPanic recovers from panics and logs the error
func RecoverFromPanic(logger *Logger) {
    if r := recover(); r != nil {
        logger.Error("Recovered from panic: %v\n%s", r, debug.Stack())
    }
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
    return os.MkdirAll(path, 0755)
}

// TempFileName generates a temporary file name with the given prefix and extension
func TempFileName(prefix, ext string) string {
    return filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), ext))
}

// FileExists checks if a file exists
func FileExists(path string) bool {
    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return false
    }
    return info.IsDir()
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int, initialDelay time.Duration) error {
    var err error
    delay := initialDelay
    
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil
        }
        
        time.Sleep(delay)
        delay *= 2 // Exponential backoff
    }
    
    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}