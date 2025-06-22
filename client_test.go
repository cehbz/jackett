package jackett

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test with default HTTP client
	client, err := NewClient("test-api-key", "localhost", "9117")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.apiKey)
	}

	if client.baseURL != "http://localhost:9117" {
		t.Errorf("Expected base URL 'http://localhost:9117', got '%s'", client.baseURL)
	}

	// Test with custom HTTP client
	customHTTP := &http.Client{}
	client2, err := NewClient("test-api-key", "localhost", "9117", customHTTP)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client2.client != customHTTP {
		t.Error("Expected custom HTTP client to be used")
	}
}

func TestSearch(t *testing.T) {
	mockSearchResponse := &SearchResponse{
		Results: []SearchResult{
			{
				Title:   "Test Movie 2024 1080p",
				Size:    1073741824,
				Seeders: 10,
				Peers:   5,
				Tracker: "TestTracker",
			},
		},
		Indexers: []struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Status  int    `json:"Status"`
			Results int64  `json:"Results"`
			Error   string `json:"Error"`
		}{
			{
				ID:      "test-indexer",
				Name:    "Test Indexer",
				Status:  0,
				Results: 1,
			},
		},
	}

	responseBody, _ := json.Marshal(mockSearchResponse)

	// Mock successful response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results": {statusCode: http.StatusOK, responseBody: string(responseBody)},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers/all/results", query: url.Values{"apikey": []string{"test-api-key"}, "Query": []string{"test movie"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	results, err := client.Search("test movie")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results.Results))
	}

	if results.Results[0].Title != "Test Movie 2024 1080p" {
		t.Errorf("Expected title 'Test Movie 2024 1080p', got '%s'", results.Results[0].Title)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestSearchWithIndexer(t *testing.T) {
	mockSearchResponse := &SearchResponse{
		Results: []SearchResult{
			{
				Title:   "Test Show S01E01",
				Size:    536870912,
				Seeders: 5,
				Peers:   8,
				Tracker: "SpecificTracker",
			},
		},
	}

	responseBody, _ := json.Marshal(mockSearchResponse)

	// Mock successful response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers/specific-indexer/results": {statusCode: http.StatusOK, responseBody: string(responseBody)},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers/specific-indexer/results", query: url.Values{"apikey": []string{"test-api-key"}, "Query": []string{"test show"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	results, err := client.SearchWithIndexer("specific-indexer", "test show")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results.Results))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestGetIndexers(t *testing.T) {
	mockIndexers := []Indexer{
		{
			ID:          "indexer1",
			Name:        "Test Indexer 1",
			Description: "A test indexer",
			Type:        "public",
			Configured:  true,
			SiteLink:    "https://example1.com",
		},
		{
			ID:          "indexer2",
			Name:        "Test Indexer 2",
			Description: "Another test indexer",
			Type:        "private",
			Configured:  true,
			SiteLink:    "https://example2.com",
		},
	}

	responseBody, _ := json.Marshal(mockIndexers)

	// Mock successful response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers": {statusCode: http.StatusOK, responseBody: string(responseBody)},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers", query: url.Values{"apikey": []string{"test-api-key"}, "configured": []string{"true"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 2 {
		t.Errorf("Expected 2 indexers, got %d", len(indexers))
	}

	if indexers[0].Name != "Test Indexer 1" {
		t.Errorf("Expected first indexer name 'Test Indexer 1', got '%s'", indexers[0].Name)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestDownloadTorrent(t *testing.T) {
	expectedData := "torrent file data"

	// Test downloading from Jackett URL
	endpointResponses := map[string]mockResponse{
		"/dl/test": {statusCode: http.StatusOK, responseBody: expectedData},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/dl/test", query: url.Values{"apikey": []string{"test-api-key"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with Jackett URL
	data, err := client.DownloadTorrent("http://localhost:9117/dl/test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != expectedData {
		t.Errorf("Expected %s, got %s", expectedData, string(data))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestDownloadTorrent_ExternalURL(t *testing.T) {
	expectedData := "external torrent data"

	// For external URLs, the mock transport will handle any URL
	endpointResponses := map[string]mockResponse{
		"": {statusCode: http.StatusOK, responseBody: expectedData}, // Empty key for external URLs
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "https://external.com/torrent.torrent"},
	}

	client, mockTransport, err := newMockClientWithExternalURL(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with external URL
	data, err := client.DownloadTorrent("https://external.com/torrent.torrent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != expectedData {
		t.Errorf("Expected %s, got %s", expectedData, string(data))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTestConnection(t *testing.T) {
	// Mock successful response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers": {statusCode: http.StatusOK, responseBody: "[]"},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers", query: url.Values{"apikey": []string{"test-api-key"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TestConnection()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTestConnection_Error(t *testing.T) {
	// Mock error response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers": {statusCode: http.StatusUnauthorized, responseBody: "Invalid API key"},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TestConnection()
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestGetServerConfig(t *testing.T) {
	mockConfig := map[string]interface{}{
		"notices":          []string{},
		"port":             9117,
		"external":         false,
		"api_key":          "test-api-key",
		"blackholedir":     "/downloads",
		"updatedisabled":   false,
		"prerelease":       false,
		"password":         "",
		"logging":          false,
		"basepathoverride": "",
		"omdbkey":          "",
		"omdburl":          "",
		"app_version":      "0.20.0",
		"can_run_netcore":  true,
		"isfreebsd":        false,
		"runtime":          "6.0.0",
	}

	responseBody, _ := json.Marshal(mockConfig)

	// Mock successful response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/server/config": {statusCode: http.StatusOK, responseBody: string(responseBody)},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/server/config", query: url.Values{"apikey": []string{"test-api-key"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	config, err := client.GetServerConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config["app_version"] != "0.20.0" {
		t.Errorf("Expected app_version '0.20.0', got '%v'", config["app_version"])
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestSearch_Error(t *testing.T) {
	// Mock error response
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results": {statusCode: http.StatusInternalServerError, responseBody: "Server error"},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers/all/results"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = client.Search("test")
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}
