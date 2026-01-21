package common

import "fmt"

// CustomError represents a custom error with additional context
type CustomError struct {
	Code    string
	Message string
	Err     error
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewCustomError creates a new custom error
func NewCustomError(code, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes
const (
	ErrConfigLoad   = "CONFIG_LOAD_ERROR"
	ErrDBConnect    = "DB_CONNECT_ERROR"
	ErrCacheConnect = "CACHE_CONNECT_ERROR"
	ErrWSConnect    = "WS_CONNECT_ERROR"
	ErrInsert       = "INSERT_ERROR"
	ErrMarshal      = "MARSHAL_ERROR"
	ErrUnmarshal    = "UNMARSHAL_ERROR"
	ErrEnvLoad      = "ENV_LOAD_ERROR"
)
