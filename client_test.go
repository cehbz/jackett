package jackett

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// Common XML response constants for testing
const (
	// Basic indexer XML structure
	basicIndexerXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="test-indexer" configured="true">
    <title>Test Indexer</title>
    <description>Test Description</description>
    <link>https://test.example.com</link>
    <language>en-US</language>
    <type>private</type>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="yes" supportedParams="q,season,ep" />
        <movie-search available="yes" supportedParams="q" />
        <music-search available="yes" supportedParams="q" />
        <audio-search available="yes" supportedParams="q" />
        <book-search available="yes" supportedParams="q" />
      </searching>
      <categories>
        <category id="2000" name="Movies" />
        <category id="5000" name="TV" />
      </categories>
    </caps>
  </indexer>
</indexers>`

	// Complex indexer with nested categories
	complexIndexerXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="complex-indexer" configured="true">
    <title>Complex Indexer</title>
    <description>Indexer with complex categories</description>
    <link>https://complex.example.com</link>
    <language>en-US</language>
    <type>private</type>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="yes" supportedParams="q,season,ep" />
        <movie-search available="yes" supportedParams="q,imdbid" />
        <music-search available="yes" supportedParams="q,album,artist" />
        <audio-search available="yes" supportedParams="q,album,artist" />
        <book-search available="no" supportedParams="q" />
      </searching>
      <categories>
        <category id="2000" name="Movies">
          <subcat id="2040" name="Movies/HD" />
          <subcat id="2045" name="Movies/UHD" />
          <subcat id="2050" name="Movies/BluRay" />
        </category>
        <category id="3000" name="Audio">
          <subcat id="3020" name="Audio/Video" />
        </category>
        <category id="5000" name="TV">
          <subcat id="5040" name="TV/HD" />
          <subcat id="5045" name="TV/UHD" />
          <subcat id="5060" name="TV/Sport" />
          <subcat id="5070" name="TV/Anime" />
          <subcat id="5080" name="TV/Documentary" />
        </category>
        <category id="6000" name="XXX" />
        <category id="7000" name="Books">
          <subcat id="7020" name="Books/EBook" />
          <subcat id="7030" name="Books/Comics" />
        </category>
        <category id="8000" name="Other" />
        <category id="100001" name="Custom Category 1" />
        <category id="100002" name="Custom Category 2" />
      </categories>
    </caps>
  </indexer>
</indexers>`

	// All indexers (configured and unconfigured)
	allIndexersXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="configured-indexer" configured="true">
    <title>Configured Indexer</title>
    <description>This indexer is configured</description>
    <link>https://configured.example.com</link>
    <language>en-US</language>
    <type>private</type>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="yes" supportedParams="q,season,ep" />
        <movie-search available="yes" supportedParams="q" />
        <music-search available="yes" supportedParams="q" />
        <audio-search available="yes" supportedParams="q" />
        <book-search available="yes" supportedParams="q" />
      </searching>
      <categories>
        <category id="2000" name="Movies" />
        <category id="5000" name="TV" />
      </categories>
    </caps>
  </indexer>
  <indexer id="unconfigured-indexer" configured="false">
    <title>Unconfigured Indexer</title>
    <description>This indexer is not configured</description>
    <link>https://unconfigured.example.com</link>
    <language>en-US</language>
    <type>public</type>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="no" supportedParams="q" />
        <movie-search available="yes" supportedParams="q" />
        <music-search available="no" supportedParams="q" />
        <audio-search available="no" supportedParams="q" />
        <book-search available="no" supportedParams="q" />
      </searching>
      <categories>
        <category id="2000" name="Movies" />
      </categories>
    </caps>
  </indexer>
</indexers>`

	// Error responses
	invalidAPIKeyXML = `<?xml version="1.0" encoding="UTF-8"?>
<error code="100" description="Invalid API Key" />`

	serverErrorXML = `<?xml version="1.0" encoding="UTF-8"?>
<error code="500" description="Internal Server Error" />`

	// Empty responses
	emptyIndexersXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers />`

	// Malformed XML
	malformedXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="test" configured="true">
    <title>Test Indexer</title>
    <description>Test Description
    <!-- Missing closing tags -->
`

	// Minimal indexer (missing optional fields)
	minimalIndexerXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="minimal" configured="true">
    <title>Minimal Indexer</title>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="no" supportedParams="q" />
        <movie-search available="no" supportedParams="q" />
        <music-search available="no" supportedParams="q" />
        <audio-search available="no" supportedParams="q" />
        <book-search available="no" supportedParams="q" />
      </searching>
      <categories>
        <category id="8000" name="Other" />
      </categories>
    </caps>
  </indexer>
</indexers>`

	// Indexer with no categories
	noCategoriesXML = `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
  <indexer id="no-categories" configured="true">
    <title>No Categories Indexer</title>
    <description>Indexer with no categories</description>
    <link>https://nocats.example.com</link>
    <language>en-US</language>
    <type>public</type>
    <caps>
      <server title="Jackett" />
      <limits default="100" max="100" />
      <searching>
        <search available="yes" supportedParams="q" />
        <tv-search available="no" supportedParams="q" />
        <movie-search available="no" supportedParams="q" />
        <music-search available="no" supportedParams="q" />
        <audio-search available="no" supportedParams="q" />
        <book-search available="no" supportedParams="q" />
      </searching>
      <categories>
      </categories>
    </caps>
  </indexer>
</indexers>`
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

func TestNewClient(t *testing.T) {
	// Test with default HTTP client
	client, err := NewClient("http://localhost:9117", "test-api-key")
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
	client2, err := NewClient("http://localhost:9117", "test-api-key", customHTTP)
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
				Title:                "Test Movie 2024 1080p",
				Size:                 1073741824,
				Seeders:              10,
				Peers:                5,
				Tracker:              "TestTracker",
				Category:             []int{2000},
				CategoryDesc:         "Movies",
				Link:                 "http://example.com/torrent",
				MagnetURI:            "magnet:?xt=urn:btih:...",
				GUID:                 "guid-123",
				PublishDate:          "2024-01-01T00:00:00Z",
				BlackholeLink:        nil,
				Gain:                 0,
				InfoHash:             "hash123",
				MinimumRatio:         nil,
				MinimumSeedTime:      nil,
				DownloadVolumeFactor: 1,
				UploadVolumeFactor:   1,
				FirstSeen:            "2024-01-01T00:00:00Z",
				TrackerId:            "testtrackerid",
				TrackerType:          "private",
				Details:              "http://example.com/details",
				Files:                nil,
				Grabs:                nil,
				Description:          nil,
				RageID:               nil,
				TVDBId:               nil,
				Imdb:                 nil,
				TMDb:                 nil,
				TVMazeId:             nil,
				TraktId:              nil,
				DoubanId:             nil,
				Genres:               nil,
				Languages:            []string{},
				Subs:                 []string{},
				Year:                 nil,
				Author:               nil,
				BookTitle:            nil,
				Publisher:            nil,
				Artist:               nil,
				Album:                nil,
				Label:                nil,
				Track:                nil,
				Poster:               nil,
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
				Title:                "Test Show S01E01",
				Size:                 536870912,
				Seeders:              5,
				Peers:                8,
				Tracker:              "SpecificTracker",
				Category:             []int{5000},
				CategoryDesc:         "TV",
				Link:                 "http://example.com/tv",
				MagnetURI:            "magnet:?xt=urn:btih:...",
				GUID:                 "guid-456",
				PublishDate:          "2024-01-02T00:00:00Z",
				BlackholeLink:        nil,
				Gain:                 0,
				InfoHash:             "hash456",
				MinimumRatio:         nil,
				MinimumSeedTime:      nil,
				DownloadVolumeFactor: 1,
				UploadVolumeFactor:   1,
				FirstSeen:            "2024-01-02T00:00:00Z",
				TrackerId:            "specifictrackerid",
				TrackerType:          "private",
				Details:              "http://example.com/tvdetails",
				Files:                nil,
				Grabs:                nil,
				Description:          nil,
				RageID:               nil,
				TVDBId:               nil,
				Imdb:                 nil,
				TMDb:                 nil,
				TVMazeId:             nil,
				TraktId:              nil,
				DoubanId:             nil,
				Genres:               nil,
				Languages:            []string{},
				Subs:                 []string{},
				Year:                 nil,
				Author:               nil,
				BookTitle:            nil,
				Publisher:            nil,
				Artist:               nil,
				Album:                nil,
				Label:                nil,
				Track:                nil,
				Poster:               nil,
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
	// Mock successful response using basic XML
	endpointResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {statusCode: http.StatusOK, responseBody: basicIndexerXML},
	}
	expectedRequests := []expectedRequest{
		{method: "GET", url: "/api/v2.0/indexers/all/results/torznab", query: url.Values{"apikey": []string{"test-api-key"}, "t": []string{"indexers"}, "configured": []string{"true"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 1 {
		t.Errorf("Expected 1 indexer, got %d", len(indexers))
	}

	if indexers[0].Name != "Test Indexer" {
		t.Errorf("Expected indexer name 'Test Indexer', got '%s'", indexers[0].Name)
	}

	if indexers[0].ID != "test-indexer" {
		t.Errorf("Expected indexer ID 'test-indexer', got '%s'", indexers[0].ID)
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

func TestGetIndexers_InvalidAPIKey(t *testing.T) {
	// Test with invalid API key - should return XML error response
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   401,
			responseBody: invalidAPIKeyXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"invalid-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Override the API key to be invalid
	client.apiKey = "invalid-key"

	indexers, err := client.GetIndexers()
	if err == nil {
		t.Error("Expected error for invalid API key, got nil")
	}

	if !strings.Contains(err.Error(), "Invalid API Key") {
		t.Errorf("Expected error to contain 'Invalid API Key', got: %v", err)
	}

	if indexers != nil {
		t.Error("Expected nil indexers for invalid API key")
	}
}

func TestGetIndexers_EmptyResponse(t *testing.T) {
	// Test with empty response (no indexers configured)
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: emptyIndexersXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error for empty response, got %v", err)
	}

	if len(indexers) != 0 {
		t.Errorf("Expected 0 indexers for empty response, got %d", len(indexers))
	}
}

func TestGetIndexers_MalformedXML(t *testing.T) {
	// Test with malformed XML response
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: malformedXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err == nil {
		t.Error("Expected error for malformed XML, got nil")
	}

	if !strings.Contains(err.Error(), "XML") {
		t.Errorf("Expected XML parsing error, got: %v", err)
	}

	if indexers != nil {
		t.Error("Expected nil indexers for malformed XML")
	}
}

func TestGetIndexers_HTTPError(t *testing.T) {
	// Test with HTTP error response
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   500,
			responseBody: serverErrorXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err == nil {
		t.Error("Expected error for HTTP 500, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain status code 500, got: %v", err)
	}

	if indexers != nil {
		t.Error("Expected nil indexers for HTTP error")
	}
}

func TestGetIndexers_AllIndexers(t *testing.T) {
	// Test getting all indexers (configured=false)
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: allIndexersXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"false"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Call the internal method directly to test configured=false
	// We need to modify the client to use configured=false
	params := url.Values{}
	params.Set("apikey", client.apiKey)
	params.Set("t", "indexers")
	params.Set("configured", "false")

	respData, err := client.doGet("/api/v2.0/indexers/all/results/torznab", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var torznabResponse TorznabIndexersResponse
	if err := xml.Unmarshal(respData, &torznabResponse); err != nil {
		t.Fatalf("Failed to decode indexers response: %v", err)
	}

	// Convert TorznabIndexer to Indexer
	indexers := make([]Indexer, len(torznabResponse.Indexers))
	for i, tIdx := range torznabResponse.Indexers {
		// Convert caps
		caps := &Caps{
			Server: tIdx.Caps.Server.Title,
			Limits: Limits{
				Default: tIdx.Caps.Limits.Default,
				Max:     tIdx.Caps.Limits.Max,
			},
			Searching: Searching{
				Search:      convertSearchType(tIdx.Caps.Searching.Search),
				TVSearch:    convertSearchType(tIdx.Caps.Searching.TVSearch),
				MovieSearch: convertSearchType(tIdx.Caps.Searching.MovieSearch),
				MusicSearch: convertSearchType(tIdx.Caps.Searching.MusicSearch),
				AudioSearch: convertSearchType(tIdx.Caps.Searching.AudioSearch),
				BookSearch:  convertSearchType(tIdx.Caps.Searching.BookSearch),
			},
		}

		// Convert categories
		categories := make([]Category, len(tIdx.Caps.Categories.Categories))
		for j, tCat := range tIdx.Caps.Categories.Categories {
			subcats := make([]Subcat, len(tCat.Subcats))
			for k, tSubcat := range tCat.Subcats {
				subcats[k] = Subcat(tSubcat)
			}
			categories[j] = Category{
				ID:      tCat.ID,
				Name:    tCat.Name,
				Subcats: subcats,
			}
		}

		indexers[i] = Indexer{
			ID:          tIdx.ID,
			Name:        tIdx.Title,
			Description: tIdx.Description,
			Type:        tIdx.Type,
			Configured:  tIdx.Configured,
			SiteLink:    tIdx.Link,
			Language:    tIdx.Language,
			Caps:        caps,
			Categories:  categories,
		}
	}
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 2 {
		t.Errorf("Expected 2 indexers, got %d", len(indexers))
	}

	// Check configured indexer
	if indexers[0].ID != "configured-indexer" {
		t.Errorf("Expected first indexer ID 'configured-indexer', got '%s'", indexers[0].ID)
	}
	if !indexers[0].Configured {
		t.Error("Expected first indexer to be configured")
	}

	// Check unconfigured indexer
	if indexers[1].ID != "unconfigured-indexer" {
		t.Errorf("Expected second indexer ID 'unconfigured-indexer', got '%s'", indexers[1].ID)
	}
	if indexers[1].Configured {
		t.Error("Expected second indexer to be unconfigured")
	}
}

func TestGetIndexers_ComplexCategories(t *testing.T) {
	// Test with complex nested categories
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: complexIndexerXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 1 {
		t.Fatalf("Expected 1 indexer, got %d", len(indexers))
	}

	indexer := indexers[0]
	if indexer.ID != "complex-indexer" {
		t.Errorf("Expected indexer ID 'complex-indexer', got '%s'", indexer.ID)
	}

	// Check categories
	if len(indexer.Categories) != 8 {
		t.Errorf("Expected 8 categories, got %d", len(indexer.Categories))
	}

	// Find Movies category with subcategories
	var moviesCat *Category
	for _, cat := range indexer.Categories {
		if cat.ID == 2000 && cat.Name == "Movies" {
			moviesCat = &cat
			break
		}
	}
	if moviesCat == nil {
		t.Error("Expected Movies category not found")
	} else if len(moviesCat.Subcats) != 3 {
		t.Errorf("Expected 3 subcategories for Movies, got %d", len(moviesCat.Subcats))
	}

	// Find TV category with subcategories
	var tvCat *Category
	for _, cat := range indexer.Categories {
		if cat.ID == 5000 && cat.Name == "TV" {
			tvCat = &cat
			break
		}
	}
	if tvCat == nil {
		t.Error("Expected TV category not found")
	} else if len(tvCat.Subcats) != 5 {
		t.Errorf("Expected 5 subcategories for TV, got %d", len(tvCat.Subcats))
	}

	// Find custom categories
	var customCat1 *Category
	for _, cat := range indexer.Categories {
		if cat.ID == 100001 && cat.Name == "Custom Category 1" {
			customCat1 = &cat
			break
		}
	}
	if customCat1 == nil {
		t.Error("Expected Custom Category 1 not found")
	}
}

func TestGetIndexers_EmptyCategories(t *testing.T) {
	// Test with indexer that has no categories
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: noCategoriesXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 1 {
		t.Fatalf("Expected 1 indexer, got %d", len(indexers))
	}

	indexer := indexers[0]
	if len(indexer.Categories) != 0 {
		t.Errorf("Expected 0 categories, got %d", len(indexer.Categories))
	}
}

func TestGetIndexers_MissingOptionalFields(t *testing.T) {
	// Test with indexer missing optional fields
	mockResponses := map[string]mockResponse{
		"/api/v2.0/indexers/all/results/torznab": {
			statusCode:   200,
			responseBody: minimalIndexerXML,
		},
	}

	expectedRequests := []expectedRequest{
		{
			method: "GET",
			url:    "/api/v2.0/indexers/all/results/torznab",
			query: url.Values{
				"apikey":     []string{"test-api-key"},
				"t":          []string{"indexers"},
				"configured": []string{"true"},
			},
		},
	}

	client, _, err := newMockClient(mockResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	indexers, err := client.GetIndexers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(indexers) != 1 {
		t.Fatalf("Expected 1 indexer, got %d", len(indexers))
	}

	indexer := indexers[0]
	if indexer.ID != "minimal" {
		t.Errorf("Expected indexer ID 'minimal', got '%s'", indexer.ID)
	}
	if indexer.Name != "Minimal Indexer" {
		t.Errorf("Expected name 'Minimal Indexer', got '%s'", indexer.Name)
	}
	// Optional fields should be empty strings
	if indexer.Description != "" {
		t.Errorf("Expected empty description, got '%s'", indexer.Description)
	}
	if indexer.SiteLink != "" {
		t.Errorf("Expected empty site link, got '%s'", indexer.SiteLink)
	}
	if indexer.Language != "" {
		t.Errorf("Expected empty language, got '%s'", indexer.Language)
	}
	if indexer.Type != "" {
		t.Errorf("Expected empty type, got '%s'", indexer.Type)
	}
}
