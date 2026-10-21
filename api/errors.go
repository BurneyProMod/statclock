package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Errors callers can check for with errors.Is.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized, check FACEIT_API_KEY")
	ErrRateLimited  = errors.New("rate limited")
)

// APIError is returned when FACEIT responds with a non-200 status.
// It unwraps to one of the errors above when the status code matches.
type APIError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration // from the Retry-After header, if sent
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("http %d", e.StatusCode)
	if err := e.Unwrap(); err != nil {
		msg = fmt.Sprintf("%v (%s)", err, msg)
	}
	if e.Body != "" {
		msg += ": " + e.Body
	}
	return msg
}

func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}
	return nil
}
