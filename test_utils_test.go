package jackett

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// mockRoundTripper is used to mock http.Client responses
type mockRoundTripper struct {
	responses        map[string]mockResponse
	expectedRequests []expectedRequest
	requestIndex     int
	t                *testing.T
	allowExternal    bool // Allow external URLs
}

// mockResponse represents a mock HTTP response
type mockResponse struct {
	statusCode   int
	responseBody string
}

// expectedRequest represents an expected HTTP request
type expectedRequest struct {
	method string
	url    string
	query  url.Values
}

// RoundTrip implements the RoundTripper interface
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.requestIndex >= len(m.expectedRequests) {
		m.t.Errorf("Unexpected request: %s %s", req.Method, req.URL.String())
		return nil, fmt.Errorf("unexpected request")
	}

	expected := m.expectedRequests[m.requestIndex]
	m.requestIndex++

	// Check method
	if req.Method != expected.method {
		m.t.Errorf("Expected method %s, got %s", expected.method, req.Method)
	}

	// For external URLs, just check the full URL
	if m.allowExternal && strings.HasPrefix(req.URL.String(), "http") {
		if req.URL.String() != expected.url {
			m.t.Errorf("Expected URL %s, got %s", expected.url, req.URL.String())
		}
		// Return response for external URL (use empty key)
		resp := m.responses[""]
		return &http.Response{
			StatusCode: resp.statusCode,
			Body:       io.NopCloser(strings.NewReader(resp.responseBody)),
			Header:     make(http.Header),
		}, nil
	}

	// Check path
	if req.URL.Path != expected.url {
		m.t.Errorf("Expected URL %s, got %s", expected.url, req.URL.Path)
	}

	// Check query parameters if specified
	if expected.query != nil {
		if !reflect.DeepEqual(req.URL.Query(), expected.query) {
			m.t.Errorf("Expected query params %v, got %v", expected.query, req.URL.Query())
		}
	}

	resp := m.responses[req.URL.Path]
	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.responseBody)),
		Header:     make(http.Header),
	}, nil
}

// newMockClient creates a mock client with predefined responses
func newMockClient(responses map[string]mockResponse, expectedRequests []expectedRequest) (*Client, *mockRoundTripper, error) {
	transport := &mockRoundTripper{
		responses:        responses,
		expectedRequests: expectedRequests,
		t:                &testing.T{},
	}

	httpClient := &http.Client{Transport: transport}
	client, err := NewClient("http://localhost:9117", "test-api-key", httpClient)
	return client, transport, err
}

// newMockClientWithExternalURL creates a mock client that allows external URLs
func newMockClientWithExternalURL(responses map[string]mockResponse, expectedRequests []expectedRequest) (*Client, *mockRoundTripper, error) {
	transport := &mockRoundTripper{
		responses:        responses,
		expectedRequests: expectedRequests,
		t:                &testing.T{},
		allowExternal:    true,
	}

	httpClient := &http.Client{Transport: transport}
	client, err := NewClient("http://localhost:9117", "test-api-key", httpClient)
	return client, transport, err
}
