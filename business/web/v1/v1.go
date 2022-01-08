package v1

import "errors"

// ErrorResponse is the form used for API responses from failure in the API.
type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
}

func NewRequestError(err error, status int) error {
	return &RequestError{err, status}
}

func (err *RequestError) Error() string {
	return err.Err.Error()
}

func IsRequestError(err error) bool {
	var re *RequestError
	return errors.As(err, &re)
}

func GetRequestError(err error) *RequestError {
	var re *RequestError
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
