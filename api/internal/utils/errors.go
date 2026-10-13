package utils

import "time"

type ErrorKind int

const (
	ErrInvalid ErrorKind = iota + 1
	ErrUnauthorized
	ErrForbidden
	ErrNotFound
	ErrConflict
	ErrTooManyRequests
)

// Error is a failure the caller can do something about. Services return it so that
// controllers can tell it apart from a server fault, without knowing about HTTP themselves.
type Error struct {
	Kind       ErrorKind
	Message    string
	RetryAfter time.Duration
}

func (failure *Error) Error() string {
	return failure.Message
}

func Invalid(message string) error {
	return &Error{Kind: ErrInvalid, Message: message}
}

func Unauthorized(message string) error {
	return &Error{Kind: ErrUnauthorized, Message: message}
}

func Forbidden(message string) error {
	return &Error{Kind: ErrForbidden, Message: message}
}

func NotFound(message string) error {
	return &Error{Kind: ErrNotFound, Message: message}
}

func Conflict(message string) error {
	return &Error{Kind: ErrConflict, Message: message}
}

func TooManyRequests(retryAfter time.Duration) error {
	return &Error{Kind: ErrTooManyRequests, Message: "Too many requests, please try again later", RetryAfter: retryAfter}
}
