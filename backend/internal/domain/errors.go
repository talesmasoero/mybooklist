package domain

import (
	"errors"
	"fmt"
)

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrReadingNotFound   = errors.New("reading not found")
	ErrAlreadyInLibrary  = errors.New("book already in user library")
	ErrExternalAPIFailed = errors.New("external API request failed")
)

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
