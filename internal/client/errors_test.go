package client

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "simple error message",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
			},
			expected: "API error 400: Bad request",
		},
		{
			name: "error with field errors",
			err: &APIError{
				StatusCode: 422,
				Message:    "Validation failed",
				Errors: map[string][]string{
					"name": {"Name is required", "Name must be unique"},
				},
			},
			expected: "API error 422: name: Name is required, Name must be unique",
		},
		{
			name: "error with multiple field errors",
			err: &APIError{
				StatusCode: 422,
				Message:    "Validation failed",
				Errors: map[string][]string{
					"name":  {"Name is required"},
					"email": {"Email is invalid"},
				},
			},
			// Note: map iteration order is not guaranteed, so we just check it contains expected parts
			expected: "API error 422:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if tt.name == "error with multiple field errors" {
				// For multiple fields, just check prefix
				if got[:14] != tt.expected {
					t.Errorf("APIError.Error() = %v, want prefix %v", got, tt.expected)
				}
			} else {
				if got != tt.expected {
					t.Errorf("APIError.Error() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "404 API error",
			err: &APIError{
				StatusCode: 404,
				Message:    "Not found",
			},
			expected: true,
		},
		{
			name: "400 API error",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
			},
			expected: false,
		},
		{
			name: "500 API error",
			err: &APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: false,
		},
		{
			name: "NotFoundError",
			err: &NotFoundError{
				Resource: "VPS",
				ID:       "123",
			},
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{
		Resource: "VPS instance",
		ID:       "vps-123",
	}

	expected := "VPS instance with ID vps-123 not found"
	if err.Error() != expected {
		t.Errorf("NotFoundError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantMsg    string
	}{
		{
			name:       "JSON error response with message",
			statusCode: 400,
			body:       []byte(`{"message": "Invalid request"}`),
			wantMsg:    "Invalid request",
		},
		{
			name:       "JSON error response with error field",
			statusCode: 403,
			body:       []byte(`{"error": "Forbidden"}`),
			wantMsg:    "Forbidden",
		},
		{
			name:       "JSON error response with validation errors",
			statusCode: 422,
			body:       []byte(`{"message": "Validation failed", "errors": {"name": ["Name is required"]}}`),
			wantMsg:    "name: Name is required",
		},
		{
			name:       "non-JSON response",
			statusCode: 500,
			body:       []byte("Internal Server Error"),
			wantMsg:    "Internal Server Error",
		},
		{
			name:       "empty JSON response",
			statusCode: 400,
			body:       []byte(`{}`),
			wantMsg:    "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.statusCode, tt.body)
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, tt.statusCode)
			}
			// Check that the error message contains the expected message
			errStr := apiErr.Error()
			if tt.name != "JSON error response with validation errors" {
				if apiErr.Message != tt.wantMsg {
					t.Errorf("Message = %v, want %v", apiErr.Message, tt.wantMsg)
				}
			} else {
				// For validation errors, check the full error string
				if errStr != "API error 422: "+tt.wantMsg {
					t.Errorf("Error() = %v, want contains %v", errStr, tt.wantMsg)
				}
			}
		})
	}
}
