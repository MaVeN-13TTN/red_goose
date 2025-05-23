package errors

import (
    "fmt"
)

type ErrorType int

const (
    ErrorTypeNetwork ErrorType = iota
    ErrorTypeExtraction
    ErrorTypeDownload
    ErrorTypeFileSystem
    ErrorTypeConfiguration
    ErrorTypeValidation
)

type RedGooseError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *RedGooseError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e *RedGooseError) Unwrap() error {
    return e.Cause
}

func NewNetworkError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeNetwork,
        Message: message,
        Cause:   cause,
    }
}

func NewExtractionError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeExtraction,
        Message: message,
        Cause:   cause,
    }
}

func NewDownloadError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeDownload,
        Message: message,
        Cause:   cause,
    }
}

func NewFileSystemError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeFileSystem,
        Message: message,
        Cause:   cause,
    }
}

func NewConfigurationError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeConfiguration,
        Message: message,
        Cause:   cause,
    }
}

func NewValidationError(message string, cause error) *RedGooseError {
    return &RedGooseError{
        Type:    ErrorTypeValidation,
        Message: message,
        Cause:   cause,
    }
}

func IsRetryableError(err error) bool {
    if rge, ok := err.(*RedGooseError); ok {
        return rge.Type == ErrorTypeNetwork || rge.Type == ErrorTypeDownload
    }
    return false
}