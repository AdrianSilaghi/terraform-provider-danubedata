package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient creates a client configured to use a test server
func newTestClient(server *httptest.Server) *Client {
	return New(Config{
		BaseURL:   server.URL,
		APIToken:  "test-token",
		UserAgent: "test-agent",
	})
}

// newTestServer creates a test HTTP server with the given handler
func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestClient_New(t *testing.T) {
	c := New(Config{
		BaseURL:   "https://api.example.com",
		APIToken:  "my-token",
		UserAgent: "my-agent",
	})

	if c.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %v, want %v", c.baseURL, "https://api.example.com")
	}
	if c.apiToken != "my-token" {
		t.Errorf("apiToken = %v, want %v", c.apiToken, "my-token")
	}
	if c.userAgent != "my-agent" {
		t.Errorf("userAgent = %v, want %v", c.userAgent, "my-agent")
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestClient_DoRequest_Success(t *testing.T) {
	type testResponse struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header = %v, want %v", r.Header.Get("Authorization"), "Bearer test-token")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept header = %v, want %v", r.Header.Get("Accept"), "application/json")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type header = %v, want %v", r.Header.Get("Content-Type"), "application/json")
		}
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("User-Agent header = %v, want %v", r.Header.Get("User-Agent"), "test-agent")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResponse{
			Message: "success",
			ID:      "123",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	var resp testResponse
	err := c.doRequest(context.Background(), "GET", "/test", nil, &resp)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "success" {
		t.Errorf("Message = %v, want %v", resp.Message, "success")
	}
	if resp.ID != "123" {
		t.Errorf("ID = %v, want %v", resp.ID, "123")
	}
}

func TestClient_DoRequest_WithBody(t *testing.T) {
	type requestBody struct {
		Name string `json:"name"`
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}

		var body requestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Name != "test-name" {
			t.Errorf("Name = %v, want %v", body.Name, "test-name")
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.doRequest(context.Background(), "POST", "/test", requestBody{Name: "test-name"}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DoRequest_APIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedStatus int
	}{
		{
			name:           "400 Bad Request",
			statusCode:     400,
			responseBody:   `{"message": "Bad request"}`,
			expectedStatus: 400,
		},
		{
			name:           "401 Unauthorized",
			statusCode:     401,
			responseBody:   `{"error": "Invalid token"}`,
			expectedStatus: 401,
		},
		{
			name:           "404 Not Found",
			statusCode:     404,
			responseBody:   `{"message": "Resource not found"}`,
			expectedStatus: 404,
		},
		{
			name:           "422 Validation Error",
			statusCode:     422,
			responseBody:   `{"message": "Validation failed", "errors": {"name": ["Name is required"]}}`,
			expectedStatus: 422,
		},
		{
			name:           "500 Internal Server Error",
			statusCode:     500,
			responseBody:   `{"message": "Internal server error"}`,
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			})
			defer server.Close()

			c := newTestClient(server)
			err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestClient_DoRequest_ContextCancellation(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := c.doRequest(ctx, "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestClient_DoRequest_EmptyResponse(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	var resp struct{}
	err := c.doRequest(context.Background(), "DELETE", "/test", nil, &resp)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
