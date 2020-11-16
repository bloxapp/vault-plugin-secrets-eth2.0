package keymanager

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// HTTPRequestError represents an HTTP request error.
type HTTPRequestError struct {
	URL          string `json:"url,required"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseBody []byte `json:"response_body,omitempty"`
	Message      string `json:"message,omitempty"`
}

// NewHTTPRequestError is the constructor of HTTPRequestError.
func NewHTTPRequestError(url string, statusCode int, responseBody []byte, message string) *HTTPRequestError {
	return &HTTPRequestError{
		URL:          url,
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Message:      message,
	}
}

// IsHTTPRequestError returns true if the given error is HTTPRequestError
func IsHTTPRequestError(err error) bool {
	_, ok := errors.Cause(err).(*HTTPRequestError)
	return ok
}

// Error implements error interface
func (e *HTTPRequestError) Error() string {
	return e.String()
}

// String returns a readable string representation of a HTTPRequestError struct.
func (e *HTTPRequestError) String() string {
	if e == nil {
		return ""
	}

	data, err := json.Marshal(e)
	if err != nil {
		logrus.Fatal(err)
	}
	return string(data)
}

// GenericError represents the generic error of keymanager.
type GenericError struct {
	ErrorMsg string `json:"error"`
}

// NewGenericError is the constructor of GenericError.
func NewGenericError(err error, desc string, args ...interface{}) *GenericError {
	return &GenericError{
		ErrorMsg: errors.Wrapf(err, desc, args...).Error(),
	}
}

// NewGenericErrorMessage is the constructor of GenericError.
func NewGenericErrorMessage(desc string, args ...interface{}) *GenericError {
	return &GenericError{
		ErrorMsg: fmt.Sprintf(desc, args...),
	}
}

// NewGenericErrorWithMessage is the constructor of GenericError.
func NewGenericErrorWithMessage(msg string) *GenericError {
	return &GenericError{
		ErrorMsg: msg,
	}
}

// IsGenericError returns true if the given error is GenericError
func IsGenericError(err error) bool {
	_, ok := errors.Cause(err).(*GenericError)
	return ok
}

// Error implements error interface.
func (e *GenericError) Error() string {
	return e.String()
}

// String implements fmt.Stringer interface.
func (e *GenericError) String() string {
	if e == nil {
		return ""
	}

	data, err := json.Marshal(e)
	if err != nil {
		logrus.Fatal(err)
	}
	return string(data)
}
