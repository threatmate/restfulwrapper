package restfulwrapper

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/tekkamanendless/httperror"
)

// ErrorWriter can be used to implement custom error responses.
// These could be plain text, JSON, etc.
//
// This package provides a number of built-in error types that implement this interface,
// but you may implement your own as well.
type ErrorWriter interface {
	// WriteError writes the error to the response.
	WriteError(resp *restful.Response)
}

// APIResponseErrorOutput is the output structure for an error.
type APIResponseErrorOutput struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

// APIHeaderParameterErrorOutput is the output structure for a header parameter error.
type APIHeaderParameterErrorOutput struct {
	APIResponseErrorOutput
	Parameter string `json:"parameter"`
}

// APIPathParameterErrorOutput is the output structure for a path parameter error.
type APIPathParameterErrorOutput struct {
	APIResponseErrorOutput
	Parameter string `json:"parameter"`
}

// APIQueryParameterErrorOutput is the output structure for a query parameter error.
type APIQueryParameterErrorOutput struct {
	APIResponseErrorOutput
	Parameter string `json:"parameter"`
}

// APIBodyError is an error that represents a parsing issue with the POST body in some way.
//
// This will always be a 400-level error.
type APIBodyError struct {
	bodyError        error
	apiResponseError *APIResponseError
}

var _ error = (*APIBodyError)(nil)
var _ ErrorWriter = (*APIBodyError)(nil)

func (e *APIBodyError) Error() string {
	return e.bodyError.Error()
}

func (e *APIBodyError) WriteError(resp *restful.Response) {
	output := APIResponseErrorOutput{
		Type:    fmt.Sprintf("%T", e),
		Message: e.Error(),
	}
	resp.WriteHeaderAndEntity(e.apiResponseError.Code(), output)
}

func (e *APIBodyError) Unwrap() []error {
	return []error{e.bodyError, e.apiResponseError}
}

// NewAPIBodyError returns a new error relating to parsing the POST body.
//
// Call this with whatever error you got when parsing the body.
func NewAPIBodyError(bodyError error) error {
	err := &APIBodyError{
		bodyError: bodyError,
		apiResponseError: &APIResponseError{
			message:   bodyError.Error(),
			httpError: httperror.ErrorFromStatus(http.StatusBadRequest),
		},
	}
	return err
}

// APIHeaderParameterError is an error that represents a header parameter error.
//
// This will always be a 400-level error.
type APIHeaderParameterError struct {
	parameter        string
	parameterError   error
	apiResponseError *APIResponseError
}

var _ error = (*APIHeaderParameterError)(nil)
var _ ErrorWriter = (*APIHeaderParameterError)(nil)

func (e *APIHeaderParameterError) Error() string {
	return e.parameterError.Error()
}
func (e *APIHeaderParameterError) WriteError(resp *restful.Response) {
	output := APIHeaderParameterErrorOutput{
		APIResponseErrorOutput: APIResponseErrorOutput{
			Type:    fmt.Sprintf("%T", e),
			Message: e.apiResponseError.message,
		},
		Parameter: e.parameter,
	}
	resp.WriteHeaderAndEntity(e.apiResponseError.Code(), output)
}

func (e *APIHeaderParameterError) Unwrap() []error {
	return []error{e.parameterError, e.apiResponseError}
}

// NewAPIHeaderParameterError returns a new path parameter error.
//
// Call this any time there is any issue at all with a path parameter.
// For example, if it is required but missing; if it has an incorrect value; or
// if it needed to be parsed and could not be parsed.
func NewAPIHeaderParameterError(parameter string, parameterError error) error {
	err := &APIHeaderParameterError{
		parameter:      parameter,
		parameterError: parameterError,
		apiResponseError: &APIResponseError{
			message:   parameterError.Error(),
			httpError: httperror.ErrorFromStatus(http.StatusBadRequest),
		},
	}
	return err
}

// APIPathParameterError is an error that represents a path parameter error.
//
// This will always be a 400-level error.
type APIPathParameterError struct {
	parameter        string
	parameterError   error
	apiResponseError *APIResponseError
}

var _ error = (*APIPathParameterError)(nil)
var _ ErrorWriter = (*APIPathParameterError)(nil)

func (e *APIPathParameterError) Error() string {
	return e.parameterError.Error()
}
func (e *APIPathParameterError) WriteError(resp *restful.Response) {
	output := APIPathParameterErrorOutput{
		APIResponseErrorOutput: APIResponseErrorOutput{
			Type:    fmt.Sprintf("%T", e),
			Message: e.apiResponseError.message,
		},
		Parameter: e.parameter,
	}
	resp.WriteHeaderAndEntity(e.apiResponseError.Code(), output)
}

func (e *APIPathParameterError) Unwrap() []error {
	return []error{e.parameterError, e.apiResponseError}
}

// NewAPIPathParameterError returns a new path parameter error.
//
// Call this any time there is any issue at all with a path parameter.
// For example, if it is required but missing; if it has an incorrect value; or
// if it needed to be parsed and could not be parsed.
func NewAPIPathParameterError(parameter string, parameterError error) error {
	err := &APIPathParameterError{
		parameter:      parameter,
		parameterError: parameterError,
		apiResponseError: &APIResponseError{
			message:   parameterError.Error(),
			httpError: httperror.ErrorFromStatus(http.StatusBadRequest),
		},
	}
	return err
}

// APIQueryParameterError is an error that represents a query parameter error.
//
// This will always be a 400-level error.
type APIQueryParameterError struct {
	parameter        string
	parameterError   error
	apiResponseError *APIResponseError
}

var _ error = (*APIQueryParameterError)(nil)
var _ ErrorWriter = (*APIQueryParameterError)(nil)

func (e *APIQueryParameterError) Error() string {
	return e.parameterError.Error()
}

func (e *APIQueryParameterError) WriteError(resp *restful.Response) {
	output := APIQueryParameterErrorOutput{
		APIResponseErrorOutput: APIResponseErrorOutput{
			Type:    fmt.Sprintf("%T", e),
			Message: e.apiResponseError.message,
		},
		Parameter: e.parameter,
	}
	resp.WriteHeaderAndEntity(e.apiResponseError.Code(), output)
}

func (e *APIQueryParameterError) Unwrap() []error {
	return []error{e.parameterError, e.apiResponseError}
}

// NewAPIQueryParameterError returns a new query parameter error.
//
// Call this any time there is any issue at all with a query parameter.
// For example, if it is required but missing; if it has an incorrect value; or
// if it needed to be parsed and could not be parsed.
func NewAPIQueryParameterError(parameter string, parameterError error) error {
	err := &APIQueryParameterError{
		parameter:      parameter,
		parameterError: parameterError,
		apiResponseError: &APIResponseError{
			message:   parameterError.Error(),
			httpError: httperror.ErrorFromStatus(http.StatusBadRequest),
		},
	}
	return err
}

// APIResponseError is an error that respresents a general HTTP response failure.
//
// This can represent any HTTP error code.
type APIResponseError struct {
	message   string
	httpError error
}

var _ error = (*APIResponseError)(nil)
var _ ErrorWriter = (*APIResponseError)(nil)

func (e *APIResponseError) Error() string {
	return e.message
}

func (e *APIResponseError) WriteError(resp *restful.Response) {
	output := APIResponseErrorOutput{
		Type:    fmt.Sprintf("%T", e),
		Message: e.message,
	}
	resp.WriteHeaderAndEntity(e.Code(), output)
}

func (e *APIResponseError) Unwrap() error {
	return e.httpError
}

// Code returns the HTTP status code.
func (e *APIResponseError) Code() int {
	var httpError *httperror.Error
	ok := errors.As(e.httpError, &httpError)
	if !ok {
		return http.StatusInternalServerError
	}
	return httpError.Code()
}

// NewAPIResponseError returns a new general API response error.
//
// Whatever HTTP status code was given will be used for the response.
// An error structure will be rendered with the given message.
//
// If the message is empty, then a default one will be generated based
// on the HTTP status code.
func NewAPIResponseError(code int, message string) error {
	if message == "" {
		message = http.StatusText(code)
	}

	err := &APIResponseError{
		message:   message,
		httpError: httperror.ErrorFromStatus(code),
	}
	return err
}
