package schema

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/riposo/riposo/pkg/riposo"
)

// Error is an error response.
type Error struct {
	StatusCode int            `json:"code"`
	ErrCode    riposo.ErrCode `json:"errno"`
	Text       string         `json:"error"`
	Message    string         `json:"message,omitempty"`
	Info       string         `json:"info,omitempty"`
	Details    interface{}    `json:"details,omitempty"`
}

// HTTPStatus returns the http status code.
func (e *Error) HTTPStatus() int { return e.StatusCode }

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Text
}

// --------------------------------------------------------------------

// InternalError generates an Error from an internal error.
func InternalError(err error) *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		ErrCode:    riposo.ErrCodeUndefined,
		Text:       http.StatusText(http.StatusInternalServerError),
		Message:    err.Error(),
	}
}

// BadRequest generates an Error.
//
//nolint:errorlint
func BadRequest(err error) *Error {
	if errors.Is(err, gzip.ErrHeader) || errors.Is(err, gzip.ErrChecksum) {
		return InvalidBody("", "Invalid gzip encoding")
	}

	switch e := err.(type) {
	case flate.CorruptInputError, flate.InternalError:
		return InvalidBody("", "Invalid flate encoding")
	case *json.SyntaxError:
		return InvalidBody("", "Invalid JSON")
	case *json.UnmarshalTypeError:
		if e.Field == "" {
			return InvalidBody("", "Invalid JSON")
		}
		return InvalidBody(e.Field, "Invalid type")
	}
	return InvalidBody("", err.Error())
}

// InvalidBody generates an Error.
func InvalidBody(field, description string) *Error {
	return invalidParams("body", field, description)
}

// InvalidQuery generates an Error.
func InvalidQuery(description string) *Error {
	return invalidParams("querystring", "", description)
}

// InvalidPath generates an Error.
func InvalidPath(description string) *Error {
	return invalidParams("path", "", description)
}

// invalidParamsDetails are used by invalidParams.
type invalidParamsDetails struct {
	Name        string `json:"name,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

func invalidParams(location, name, description string) *Error {
	message := description
	if name != "" && location != "" {
		message = name + " in " + location + ": " + description
	} else if location != "" {
		message = location + ": " + description
	}

	return &Error{
		StatusCode: http.StatusBadRequest,
		ErrCode:    riposo.ErrCodeInvalidParameters,
		Text:       "Invalid parameters",
		Message:    message,
		Details: []invalidParamsDetails{
			{
				Location:    location,
				Name:        name,
				Description: description,
			},
		},
	}
}

// invalidResourceDetails are used by InvalidResource.
type invalidResourceDetails struct {
	ID           string `json:"id"`
	ResourceName string `json:"resource_name"`
}

// MissingResource generates an Error.
func MissingResource(id, resourceName string) *Error {
	return &Error{
		StatusCode: 404,
		ErrCode:    riposo.ErrCodeMissingResource,
		Text:       "Not Found",
		Details:    invalidResourceDetails{ID: id, ResourceName: resourceName},
	}
}

// NotFound is a standard not found error response.
var NotFound = &Error{
	StatusCode: 404,
	ErrCode:    riposo.ErrCodeMissingResource,
	Text:       "Not Found",
	Message:    "The resource you are looking for could not be found.",
}

// MissingAuthToken is a standard unauthorized error response.
var MissingAuthToken = &Error{
	StatusCode: http.StatusUnauthorized,
	ErrCode:    riposo.ErrCodeMissingAuthToken,
	Text:       "Unauthorized",
	Message:    "Please authenticate yourself to use this endpoint.",
}

// MethodNotAllowed is a standard unauthorized error response.
var MethodNotAllowed = &Error{
	StatusCode: http.StatusMethodNotAllowed,
	ErrCode:    riposo.ErrCodeMethodNotAllowed,
	Text:       "Method Not Allowed",
	Message:    "Method not allowed on this endpoint.",
}

// Forbidden is a standard forbidden error response.
var Forbidden = &Error{
	StatusCode: http.StatusForbidden,
	ErrCode:    riposo.ErrCodeForbidden,
	Text:       "Forbidden",
	Message:    "This user cannot access this resource.",
}

// NotModified is a standard not modified error response.
var NotModified = &Error{
	StatusCode: http.StatusNotModified,
}

// InvalidResource generates a specific resource not found Error.
func InvalidResource(path riposo.Path) *Error {
	return &Error{
		StatusCode: http.StatusNotFound,
		ErrCode:    riposo.ErrCodeInvalidResourceID,
		Text:       http.StatusText(http.StatusNotFound),
		Details:    invalidResourceDetails{ID: path.ObjectID(), ResourceName: path.ResourceName()},
	}
}

// modifiedMeanwhileDetails are used by ModifiedMeanwhile.
type modifiedMeanwhileDetails struct {
	Existing interface{} `json:"existing"`
}

// ModifiedMeanwhile generates a PreconditionFailed Error.
func ModifiedMeanwhile(existing *Object) *Error {
	resp := &Error{
		StatusCode: http.StatusPreconditionFailed,
		ErrCode:    riposo.ErrCodeModifiedMeanwhile,
		Text:       http.StatusText(http.StatusPreconditionFailed),
		Message:    "Resource was modified meanwhile",
	}
	if existing != nil {
		resp.Details = modifiedMeanwhileDetails{Existing: existing}
	}
	return resp
}
