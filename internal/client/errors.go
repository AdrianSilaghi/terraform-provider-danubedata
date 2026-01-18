package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Message    string
	Errors     map[string][]string
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		var msgs []string
		for field, errs := range e.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", field, strings.Join(errs, ", ")))
		}
		return fmt.Sprintf("API error %d: %s", e.StatusCode, strings.Join(msgs, "; "))
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

type apiErrorResponse struct {
	Message string              `json:"message"`
	Error   string              `json:"error"`
	Errors  map[string][]string `json:"errors"`
}

func parseAPIError(statusCode int, body []byte) error {
	var errResp apiErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}

	message := errResp.Message
	if message == "" {
		message = errResp.Error
	}
	if message == "" {
		message = "Unknown error"
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Errors:     errResp.Errors,
	}
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}
