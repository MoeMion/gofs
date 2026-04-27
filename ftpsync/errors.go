package ftpsync

import (
	"errors"
	"strings"
)

// ErrorKind identifies a class of public FTPSyncService failures.
type ErrorKind string

const (
	// ErrValidation identifies invalid caller configuration.
	ErrValidation ErrorKind = "validation"
	// ErrCanceled identifies context cancellation before work is dispatched.
	ErrCanceled ErrorKind = "canceled"
	// ErrConnection identifies FTP connection setup failures.
	ErrConnection ErrorKind = "connection"
	// ErrAuthentication identifies FTP authentication failures.
	ErrAuthentication ErrorKind = "authentication"
	// ErrTransfer identifies file transfer failures.
	ErrTransfer ErrorKind = "transfer"
	// ErrUnsupportedCapability identifies known capabilities outside this API contract.
	ErrUnsupportedCapability ErrorKind = "unsupported_capability"
)

// Error is the public structured error type returned by FTPSyncService APIs.
type Error struct {
	KindValue ErrorKind
	Message   string
	Cause     error
}

// Error returns a password-safe error string with kind and contextual message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return string(e.KindValue)
	}
	return string(e.KindValue) + ": " + e.Message
}

// Unwrap returns the underlying cause for errors.Is/errors.As compatibility.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Kind returns the public error classification.
func (e *Error) Kind() ErrorKind {
	if e == nil {
		return ""
	}
	return e.KindValue
}

// IsKind reports whether err or any wrapped *Error has the requested kind.
func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if !errors.As(err, &target) {
		return false
	}
	return target.Kind() == kind
}

func newError(kind ErrorKind, message string, cause error) *Error {
	return &Error{KindValue: kind, Message: message, Cause: cause}
}

func newTransferError(message string, cause error) *Error {
	return newError(ErrTransfer, sanitizeErrorMessage(message), cause)
}

func sanitizeErrorMessage(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "transfer failed"
	}
	return message
}
