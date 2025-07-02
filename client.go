package jackett

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// Client is used to interact with the Jackett API
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
	mu      sync.RWMutex
}

// SearchResult represents a torrent search result from Jackett
type SearchResult struct {
	Title                string  `json:"Title"`
	Size                 int64   `json:"Size"`
	Seeders              int     `json:"Seeders"`
	Peers                int     `json:"Peers"`
	Link                 string  `json:"Link"`
	MagnetURI            string  `json:"MagnetUri"`
	GUID                 string  `json:"Guid"`
	PublishDate          string  `json:"PublishDate"`
	Tracker              string  `json:"Tracker"`
	Category             []int   `json:"Category"`
	CategoryDesc         string  `json:"CategoryDesc"`
	BlackholeLink        string  `json:"BlackholeLink"`
	Gain                 float64 `json:"Gain"`
	InfoHash             string  `json:"InfoHash"`
	MinimumRatio         float64 `json:"MinimumRatio"`
	MinimumSeedTime      int64   `json:"MinimumSeedTime"`
	DownloadVolumeFactor float64 `json:"DownloadVolumeFactor"`
	UploadVolumeFactor   float64 `json:"UploadVolumeFactor"`
}

// SearchResponse represents the response from a search query
type SearchResponse struct {
	Results  []SearchResult `json:"Results"`
	Indexers []struct {
		ID      string `json:"ID"`
		Name    string `json:"Name"`
		Status  int    `json:"Status"`
		Results int64  `json:"Results"`
		Error   string `json:"Error"`
	} `json:"Indexers"`
}

// Indexer represents a configured indexer in Jackett
type Indexer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Configured  bool   `json:"configured"`
	SiteLink    string `json:"site_link"`
	Language    string `json:"language"`
	LastError   string `json:"last_error"`
	Potatoe     bool   `json:"potatoe"`
	Caps        []Cap  `json:"caps"`
}

// Cap represents indexer capabilities
type Cap struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	SubCats []Cap  `json:"SubCats,omitempty"`
}

// NewClient initializes a new Jackett client.
// baseURL should be the full URL to the Jackett instance, e.g. "http://localhost:9117"
// If httpClient is nil, http.DefaultClient is used.
func NewClient(baseURL, apiKey string, httpClient ...*http.Client) (*Client, error) {
	client := http.DefaultClient
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}

	jClient := &Client{
		client:  client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}

	return jClient, nil
}

// Search performs a search query across all configured indexers
func (c *Client) Search(query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("Query", query)

	respData, err := c.doGet("/api/v2.0/indexers/all/results", params)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %v", err)
	}

	return &response, nil
}

// SearchWithIndexer performs a search query on a specific indexer
func (c *Client) SearchWithIndexer(indexerID, query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("Query", query)

	endpoint := fmt.Sprintf("/api/v2.0/indexers/%s/results", indexerID)
	respData, err := c.doGet(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %v", err)
	}

	return &response, nil
}

// GetIndexers retrieves all configured indexers
func (c *Client) GetIndexers() ([]Indexer, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("configured", "true")

	respData, err := c.doGet("/api/v2.0/indexers", params)
	if err != nil {
		return nil, fmt.Errorf("get indexers error: %v", err)
	}

	var indexers []Indexer
	if err := json.Unmarshal(respData, &indexers); err != nil {
		return nil, fmt.Errorf("failed to decode indexers response: %v", err)
	}

	return indexers, nil
}

// DownloadTorrent downloads a torrent file from the given link
func (c *Client) DownloadTorrent(link string) ([]byte, error) {
	// Parse the link to check if it's a Jackett URL
	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid download link: %v", err)
	}

	// If it's not already pointing to this Jackett instance, use it as-is
	baseURL, _ := url.Parse(c.baseURL)
	if linkURL.Host != baseURL.Host {
		// External link, download directly
		resp, err := c.client.Get(link)
		if err != nil {
			return nil, fmt.Errorf("download error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(body))
		}

		return io.ReadAll(resp.Body)
	}

	// It's a Jackett link, ensure API key is present
	query := linkURL.Query()
	if query.Get("apikey") == "" {
		query.Set("apikey", c.apiKey)
		linkURL.RawQuery = query.Encode()
	}

	resp, err := c.client.Get(linkURL.String())
	if err != nil {
		return nil, fmt.Errorf("download error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// doGet is a helper method for making GET requests to the Jackett API
func (c *Client) doGet(endpoint string, query url.Values) ([]byte, error) {
	apiURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %v", err)
	}

	apiURL.Path = endpoint
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response code: %d, response: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// TestConnection verifies the connection to Jackett
func (c *Client) TestConnection() error {
	params := url.Values{}
	params.Set("apikey", c.apiKey)

	_, err := c.doGet("/api/v2.0/indexers", params)
	if err != nil {
		return fmt.Errorf("connection test failed: %v", err)
	}

	return nil
}

// GetServerConfig retrieves the Jackett server configuration
func (c *Client) GetServerConfig() (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)

	respData, err := c.doGet("/api/v2.0/server/config", params)
	if err != nil {
		return nil, fmt.Errorf("get server config error: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(respData, &config); err != nil {
		return nil, fmt.Errorf("failed to decode server config: %v", err)
	}

	return config, nil
}
