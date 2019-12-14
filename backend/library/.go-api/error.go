package api

import (
	"fmt"
)

// Error is the HTTP response error object
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// ErrMessageClean is an error message that can be used for clients to hide technical details of the error
var ErrMessageClean = "There was an issue processing the request. Please see the logs."

// ErrEmptyBody is used when we expect to receive a request with some body but we don't
var ErrEmptyBody = fmt.Errorf("no content provided with the HTTP request")

// ErrInvalidJSON is used when we expect to receive a JSON request but we don't
var ErrInvalidJSON = fmt.Errorf("content is not a valid JSON")

// Error method makes handler.Error implement golang's error interface
func (e Error) Error() string {
	return fmt.Sprintf("Error Code %d: %s", e.Code, e.Message)
}

// NewError returns a new error instance
func NewError(code int, message string) Error {
	return Error{
		Code:    int32(code),
		Message: message,
	}
}

// CleanErrMessage prepends a clean user-friendly error text to the provided error message
func CleanErrMessage(msg string) string {
	return fmt.Sprintf("There was an error processing the request: %v", msg)
}
