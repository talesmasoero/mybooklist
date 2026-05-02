package domain

import "fmt"

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AppError struct {
	Code      int
	ErrorCode string
	Message   string
	Err       error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
